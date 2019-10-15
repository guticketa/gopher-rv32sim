package main

import (
	"fmt"
	"encoding/binary"
	"os"
)

const (
	CSR_ADDR_MVENDORID       = 0xF11
	CSR_ADDR_MARCHID         = 0xF12
	CSR_ADDR_MIMPID          = 0xF13
	CSR_ADDR_MHARTID         = 0xF14
	CSR_ADDR_MSTATUS         = 0x300
	CSR_ADDR_MISA            = 0x301
	CSR_ADDR_MEDELEG         = 0x302
	CSR_ADDR_MIDELEG         = 0x303
	CSR_ADDR_MIE             = 0x304
	CSR_ADDR_MTVEC           = 0x305
	CSR_ADDR_MCOUNTEREN      = 0x306
	CSR_ADDR_MEPC            = 0x341
	CSR_ADDR_MCAUSE          = 0x342
	CSR_ADDR_MTVAL           = 0x343
	CSR_ADDR_MIP             = 0x344
	EXCEPT_CODE_ILLEGAL_INST = 0x00000002
	EXCEPT_CODE_BREAKPOINT   = 0x00000003
	EXCEPT_CODE_ECALL_FROM_M = 0x0000000b
)

const (
	resetVec = 0x80000000
)

const (
	EI_CLASS  = 4
	EI_DATA   = 5
	EI_NIDENT = 16
)

type FileHeader struct {
	Ident     [EI_NIDENT]byte
	Type      uint16
	Machine   uint16
	Version   uint32
	Entry     uint32
	Phoff     uint32
	Shoff     uint32
	Flags     uint32
	Ehsize    uint16
	Phentsize uint16
	Phnum     uint16
	Shentsize uint16
	Shnum     uint16
	Shstrndx  uint16
}

type ProgramHeader struct {
	Type   uint32
	Off    uint32
	Vaddr  uint32
	Paddr  uint32
	Filesz uint32
	Memsz  uint32
	Flags  uint32
	Align  uint32
}

type Ops struct {
	Name   string
	Imm    uint32
	Rs1    uint32
	Rs2    uint32
	Rd     uint32
	Funct3 uint32
	Funct7 uint32
	Shamt  uint32
	Csr    uint32
}

type CPU struct {
	PC   uint32
	Regs []uint32
	CSRs []uint32
	bus *Bus
}

var _ = fmt.Println

func sext(imm uint32, bitwidth uint32) int32 {
	shift := 32 - bitwidth
	t := imm << shift
	signed := int32(t)
	return signed >> shift
}

func NewCPU() *CPU {
	bus := NewBus()
	regs := make([]uint32, 32)
	csrs := make([]uint32, 4096)
	return &CPU{0, regs, csrs, bus}
}

func (p *CPU) LoadElf(filename string) {
	t := make([]byte, 1)
	file, _ := os.Open(filename)

	var ident [16]uint8
	if _, err := file.ReadAt(ident[0:], 0); err != nil {
		return
	}
	if ident[0] != '\x7f' || ident[1] != 'E' || ident[2] != 'L' || ident[3] != 'F' {
		return
	}

	eiClass := ident[EI_CLASS]
	if eiClass != 1 { /* not LFCLASS32 */
		return
	}

	eiData := ident[EI_DATA]
	var byteOrder binary.ByteOrder

	switch eiData {
	case 1: /* little-endian. */
		byteOrder = binary.LittleEndian
	case 2: /* big-endian */
		byteOrder = binary.BigEndian
	default:
		return
	}

	fileHdr := new(FileHeader)
	file.Seek(0, os.SEEK_SET)
	if err := binary.Read(file, byteOrder, fileHdr); err != nil {
		return
	}

	progHdr := new(ProgramHeader)
	phOff := 0
	for i := 0; i < int(fileHdr.Phnum); i++ {
		file.Seek(int64(fileHdr.Phoff+uint32(phOff)), os.SEEK_SET)
		if err := binary.Read(file, byteOrder, progHdr); err != nil {
			return
		}
		phOff += 32
		pVaddr := uint32(progHdr.Vaddr)
		file.Seek(int64(progHdr.Off), os.SEEK_SET)
		for j := 0; j < int(progHdr.Memsz); j++ {
			file.Read(t)
			p.bus.WriteByte(pVaddr, uint8(t[0]))
			pVaddr++
		}
	}
	p.PC = uint32(fileHdr.Entry)
}

func (p *CPU) Reset() {
	p.PC = resetVec
}

var instructions = map[string]func(cpu *CPU, ops *Ops){
	"lui": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, ops.Imm)
		cpu.PC = cpu.PC + 4
	},
	"auipc": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.PC+ops.Imm)
		cpu.PC = cpu.PC + 4
	},
	"jal": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.PC+4)
		cpu.PC = cpu.PC + ops.Imm
	},
	"jalr": func(cpu *CPU, ops *Ops) {
		t := cpu.PC + 4
		cpu.PC = (cpu.Regs[ops.Rs1] + ops.Imm) & 0xfffffffe
		cpu.RegWrite(ops.Rd, t)
	},
	"beq": func(cpu *CPU, ops *Ops) {
		if cpu.Regs[ops.Rs1] == cpu.Regs[ops.Rs2] {
			cpu.PC = cpu.PC + ops.Imm
		} else {
			cpu.PC = cpu.PC + 4
		}
	},
	"bne": func(cpu *CPU, ops *Ops) {
		if cpu.Regs[ops.Rs1] != cpu.Regs[ops.Rs2] {
			cpu.PC = cpu.PC + ops.Imm
		} else {
			cpu.PC = cpu.PC + 4
		}
	},
	"blt": func(cpu *CPU, ops *Ops) {
		if int32(cpu.Regs[ops.Rs1]) < int32(cpu.Regs[ops.Rs2]) {
			cpu.PC = cpu.PC + ops.Imm
		} else {
			cpu.PC = cpu.PC + 4
		}
	},
	"bge": func(cpu *CPU, ops *Ops) {
		if int32(cpu.Regs[ops.Rs1]) >= int32(cpu.Regs[ops.Rs2]) {
			cpu.PC = cpu.PC + ops.Imm
		} else {
			cpu.PC = cpu.PC + 4
		}
	},
	"bltu": func(cpu *CPU, ops *Ops) {
		if cpu.Regs[ops.Rs1] < cpu.Regs[ops.Rs2] {
			cpu.PC = cpu.PC + ops.Imm
		} else {
			cpu.PC = cpu.PC + 4
		}
	},
	"bgeu": func(cpu *CPU, ops *Ops) {
		if cpu.Regs[ops.Rs1] >= cpu.Regs[ops.Rs2] {
			cpu.PC = cpu.PC + ops.Imm
		} else {
			cpu.PC = cpu.PC + 4
		}
	},
	"lb": func(cpu *CPU, ops *Ops) {
		t := cpu.bus.ReadByte(cpu.Regs[ops.Rs1] + ops.Imm)
		cpu.RegWrite(ops.Rd, uint32(sext(uint32(t), 8)))
		cpu.PC = cpu.PC + 4
	},
	"lh": func(cpu *CPU, ops *Ops) {
		t := cpu.bus.ReadHalf(cpu.Regs[ops.Rs1] + ops.Imm)
		cpu.RegWrite(ops.Rd, uint32(sext(uint32(t), 16)))
		cpu.PC = cpu.PC + 4
	},
	"lw": func(cpu *CPU, ops *Ops) {
		t := cpu.bus.ReadWord(cpu.Regs[ops.Rs1]+ops.Imm) & 0xffffffff
		cpu.RegWrite(ops.Rd, uint32(sext(t, 32)))
		cpu.PC = cpu.PC + 4
	},
	"lbu": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, uint32(cpu.bus.ReadByte(cpu.Regs[ops.Rs1]+ops.Imm)))
		cpu.PC = cpu.PC + 4
	},
	"lhu": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, uint32(cpu.bus.ReadHalf(cpu.Regs[ops.Rs1]+ops.Imm)))
		cpu.PC = cpu.PC + 4
	},
	"sb": func(cpu *CPU, ops *Ops) {
		cpu.bus.WriteByte(cpu.Regs[ops.Rs1]+ops.Imm, uint8(cpu.Regs[ops.Rs2]&0xff))
		cpu.PC = cpu.PC + 4
	},
	"sh": func(cpu *CPU, ops *Ops) {
		cpu.bus.WriteHalf(cpu.Regs[ops.Rs1]+ops.Imm, uint16(cpu.Regs[ops.Rs2]&0xffff))
		cpu.PC = cpu.PC + 4
	},
	"sw": func(cpu *CPU, ops *Ops) {
		cpu.bus.WriteWord(cpu.Regs[ops.Rs1]+ops.Imm, cpu.Regs[ops.Rs2]&0xffffffff)
		cpu.PC = cpu.PC + 4
	},
	"addi": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]+ops.Imm)
		cpu.PC = cpu.PC + 4
	},
	"slti": func(cpu *CPU, ops *Ops) {
		if int32(cpu.Regs[ops.Rs1]) < int32(ops.Imm) {
			cpu.RegWrite(ops.Rd, 1)
		} else {
			cpu.RegWrite(ops.Rd, 0)
		}
		cpu.PC = cpu.PC + 4
	},
	"sltiu": func(cpu *CPU, ops *Ops) {
		if cpu.Regs[ops.Rs1] < ops.Imm {
			cpu.RegWrite(ops.Rd, 1)
		} else {
			cpu.RegWrite(ops.Rd, 0)
		}
		cpu.PC = cpu.PC + 4
	},
	"xori": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]^ops.Imm)
		cpu.PC = cpu.PC + 4
	},
	"ori": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]|ops.Imm)
		cpu.PC = cpu.PC + 4
	},
	"andi": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]&ops.Imm)
		cpu.PC = cpu.PC + 4
	},
	"slli": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]<<ops.Shamt)
		cpu.PC = cpu.PC + 4
	},
	"srli": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]>>ops.Shamt)
		cpu.PC = cpu.PC + 4
	},
	"srai": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, uint32(int32(cpu.Regs[ops.Rs1])>>ops.Shamt))
		cpu.PC = cpu.PC + 4
	},
	"add": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]+cpu.Regs[ops.Rs2])
		cpu.PC = cpu.PC + 4
	},
	"sub": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]-cpu.Regs[ops.Rs2])
		cpu.PC = cpu.PC + 4
	},
	"sll": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]<<(cpu.Regs[ops.Rs2]&0x1f))
		cpu.PC = cpu.PC + 4
	},
	"slt": func(cpu *CPU, ops *Ops) {
		if int32(cpu.Regs[ops.Rs1]) < int32(cpu.Regs[ops.Rs2]) {
			cpu.RegWrite(ops.Rd, 1)
		} else {
			cpu.RegWrite(ops.Rd, 0)
		}
		cpu.PC = cpu.PC + 4
	},
	"sltu": func(cpu *CPU, ops *Ops) {
		if cpu.Regs[ops.Rs1] < cpu.Regs[ops.Rs2] {
			cpu.RegWrite(ops.Rd, 1)
		} else {
			cpu.RegWrite(ops.Rd, 0)
		}
		cpu.PC = cpu.PC + 4
	},
	"xor": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]^cpu.Regs[ops.Rs2])
		cpu.PC = cpu.PC + 4
	},
	"srl": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]>>(cpu.Regs[ops.Rs2]&0x1f))
		cpu.PC = cpu.PC + 4
	},
	"sra": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, uint32(int32(cpu.Regs[ops.Rs1])>>(cpu.Regs[ops.Rs2]&0x1f)))
		cpu.PC = cpu.PC + 4
	},
	"or": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]|cpu.Regs[ops.Rs2])
		cpu.PC = cpu.PC + 4
	},
	"and": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.Regs[ops.Rs1]&cpu.Regs[ops.Rs2])
		cpu.PC = cpu.PC + 4
	},
	"fence": func(cpu *CPU, ops *Ops) {
		cpu.PC = cpu.PC + 4
	},
	"fence_i": func(cpu *CPU, ops *Ops) {
		cpu.PC = cpu.PC + 4
	},
	"ecall": func(cpu *CPU, ops *Ops) {
		cpu.CSRWrite(CSR_ADDR_MEPC, &cpu.PC)
		var t uint32 = EXCEPT_CODE_ECALL_FROM_M
		cpu.CSRWrite(CSR_ADDR_MCAUSE, &t)
		var jumpAddr uint32
		cpu.CSRRead(CSR_ADDR_MTVEC, &jumpAddr)
		cpu.PC = jumpAddr
	},
	"ebreak": func(cpu *CPU, ops *Ops) {
		cpu.CSRWrite(CSR_ADDR_MEPC, &cpu.PC)
		var t uint32 = EXCEPT_CODE_BREAKPOINT
		cpu.CSRWrite(CSR_ADDR_MCAUSE, &t)
		var jumpAddr uint32
		cpu.CSRRead(CSR_ADDR_MTVEC, &jumpAddr)
		cpu.PC = jumpAddr
	},
	"mret": func(cpu *CPU, ops *Ops) {
		var t uint32
		cpu.CSRRead(CSR_ADDR_MEPC, &t)
		cpu.PC = t
	},
	"csrrw": func(cpu *CPU, ops *Ops) {
		t := cpu.CSRs[ops.Csr]
		cpu.CSRs[ops.Csr] = cpu.Regs[ops.Rs1]
		cpu.RegWrite(ops.Rd, t)
		cpu.PC = cpu.PC + 4
	},
	"csrrs": func(cpu *CPU, ops *Ops) {
		t := cpu.CSRs[ops.Csr]
		cpu.CSRs[ops.Csr] = t | cpu.Regs[ops.Rs1]
		cpu.RegWrite(ops.Rd, t)
		cpu.PC = cpu.PC + 4
	},
	"csrrc": func(cpu *CPU, ops *Ops) {
		t := cpu.CSRs[ops.Csr]
		cpu.CSRs[ops.Csr] = t & (^cpu.Regs[ops.Rs1])
		cpu.RegWrite(ops.Rd, t)
		cpu.PC = cpu.PC + 4
	},
	"csrrwi": func(cpu *CPU, ops *Ops) {
		cpu.RegWrite(ops.Rd, cpu.CSRs[ops.Csr])
		cpu.CSRs[ops.Csr] = ops.Rs1 /* zimm[4:0] */
		cpu.PC = cpu.PC + 4
	},
	"csrrsi": func(cpu *CPU, ops *Ops) {
		t := cpu.CSRs[ops.Csr]
		cpu.CSRs[ops.Csr] = t | ops.Rs1 /* zimm[4:0] */
		cpu.RegWrite(ops.Rd, t)
		cpu.PC = cpu.PC + 4
	},
	"csrrci": func(cpu *CPU, ops *Ops) {
		t := cpu.CSRs[ops.Csr]
		cpu.CSRs[ops.Csr] = t & (^ops.Rs1) /* zimm[4:0] */
		cpu.RegWrite(ops.Rd, t)
		cpu.PC = cpu.PC + 4
	},
	"illegal_instruction": func(cpu *CPU, ops *Ops) {
		cpu.CSRWrite(CSR_ADDR_MEPC, &cpu.PC)
		var t uint32 = EXCEPT_CODE_ILLEGAL_INST
		cpu.CSRWrite(CSR_ADDR_MCAUSE, &t)
		var jumpAddr uint32
		cpu.CSRRead(CSR_ADDR_MTVEC, &jumpAddr)
		cpu.PC = jumpAddr
	},
}

func (p *CPU) RegWrite(addr uint32, data uint32) {
	if addr > 0 && addr < 32 {
		p.Regs[addr] = data
	}
}

func (cpu *CPU) Fetch() uint32 {
	t := cpu.bus.ReadWord(cpu.PC)
	// cpu.PC = cpu.PC + 4
	return t
}

func (cpu *CPU) Decode(inst uint32) Ops {

	opcode := inst & 0x7f
	var ops Ops

	ops.Name = "illegal_instruction"
	ops.Rd = (inst >> 7) & 0x1f
	ops.Funct3 = (inst >> 12) & 0x7
	ops.Rs1 = (inst >> 15) & 0x1f
	ops.Rs2 = (inst >> 20) & 0x1f
	ops.Funct7 = (inst >> 25) & 0x3f
	ops.Shamt = (inst >> 20) & 0x3f
	ops.Csr = (inst >> 20) & 0xfff

	iimm := (inst >> 20) & 0xfff
	if (inst & 0x80000000) != 0 {
		iimm = 0xfffff000 | iimm
	}
	simm := (((inst >> 25) & 0x7f) << 5) | ((inst >> 7) & 0x1f)
	if (inst & 0x80000000) != 0 {
		simm = 0xfffff000 | simm
	}
	uimm := ((inst >> 12) & 0xfffff) << 12 // RV32I
	bimm := (((inst >> 31) & 1) << 12) |
		(((inst >> 7) & 1) << 11) |
		(((inst >> 25) & 0x3f) << 5) |
		(((inst >> 8) & 0xf) << 1)
	if (inst & 0x80000000) != 0 {
		bimm = 0xffffe000 | bimm
	}
	jimm := (((inst >> 31) & 0x01) << 20) |
		(((inst >> 12) & 0xff) << 12) |
		(((inst >> 20) & 0x01) << 11) |
		(((inst >> 21) & 0x3ff) << 1)
	if (inst & 0x80000000) != 0 {
		jimm = 0xffe00000 | jimm
	}

	switch opcode {
	case 0x37:
		ops.Name = "lui"
		ops.Imm = uimm
	case 0x17:
		ops.Name = "auipc"
		ops.Imm = uimm
	case 0x6f:
		ops.Name = "jal"
		ops.Imm = jimm
	case 0x67:
		ops.Name = "jalr"
		ops.Imm = iimm
	case 0x63:
		switch ops.Funct3 {
		case 0:
			ops.Name = "beq"
		case 1:
			ops.Name = "bne"
		case 4:
			ops.Name = "blt"
		case 5:
			ops.Name = "bge"
		case 6:
			ops.Name = "bltu"
		case 7:
			ops.Name = "bgeu"
		}
		ops.Imm = bimm
	case 0x03:
		switch ops.Funct3 {
		case 0:
			ops.Name = "lb"
		case 1:
			ops.Name = "lh"
		case 2:
			ops.Name = "lw"
		case 4:
			ops.Name = "lbu"
		case 5:
			ops.Name = "lhu"
		default:
			ops.Name = "illegal_instruction"
		}
		ops.Imm = iimm
	case 0x23:
		switch ops.Funct3 {
		case 0:
			ops.Name = "sb"
		case 1:
			ops.Name = "sh"
		case 2:
			ops.Name = "sw"
		default:
			ops.Name = "illegal_instruction"
		}
		ops.Imm = simm
	case 0x13:
		switch ops.Funct3 {
		case 0:
			ops.Name = "addi"
		case 1:
			ops.Name = "slli"
		case 2:
			ops.Name = "slti"
		case 3:
			ops.Name = "sltiu"
		case 4:
			ops.Name = "xori"
		case 5:
			if ops.Funct7 == 0 {
				ops.Name = "srli"
			} else {
				ops.Name = "srai"
			}
		case 6:
			ops.Name = "ori"
		case 7:
			ops.Name = "andi"
		default:
			ops.Name = "illegal_instruction"
		}
		if (ops.Funct3 == 1) || (ops.Funct3 == 5) { // slli, srli, srai
			ops.Imm = ops.Shamt
		} else {
			ops.Imm = iimm
		}
	case 0x33:
		switch ops.Funct3 {
		case 0:
			if ops.Funct7 == 0 {
				ops.Name = "add"
			} else {
				ops.Name = "sub"
			}
		case 1:
			ops.Name = "sll"
		case 2:
			ops.Name = "slt"
		case 3:
			ops.Name = "sltu"
		case 4:
			ops.Name = "xor"
		case 5:
			if ops.Funct7 == 0 {
				ops.Name = "srl"
			} else {
				ops.Name = "sra"
			}
		case 6:
			ops.Name = "or"
		case 7:
			ops.Name = "and"
		default:
			ops.Name = "illegal_instruction"
		}
		ops.Imm = bimm
	case 0x0f: //
		switch ops.Funct3 {
		case 0:
			ops.Name = "fence"
		case 1:
			ops.Name = "fence_i"
		default:
			ops.Name = "illegal_instruction"
		}
		ops.Imm = iimm
	case 0x73: // I
		switch ops.Funct3 {
		case 0:
			if ops.Csr == 0x000 {
				ops.Name = "ecall"
			} else if ops.Csr == 0x001 {
				ops.Name = "ebreak"
				// } else if ops.Csr == 0x002 {
				// 	ops.Name = "uret"
				// } else if ops.Csr == 0x102 {
				// 	ops.Name = "sret"
			} else if ops.Csr == 0x302 {
				ops.Name = "mret"
			} else {
				ops.Name = "illegal_instruction"
			}
		case 1:
			ops.Name = "csrrw"
		case 2:
			ops.Name = "csrrs"
		case 3:
			ops.Name = "csrrc"
		case 5:
			ops.Name = "csrrwi"
		case 6:
			ops.Name = "csrrsi"
		case 7:
			ops.Name = "csrrci"
		default:
			ops.Name = "illegal_instruction"
		}
		ops.Imm = iimm
	default:
		ops.Imm = 0
	}
	return ops
}

func (p *CPU) CSRRead(addr uint16, data *uint32) {
	*data = p.CSRs[addr]
}

func (p *CPU) CSRWrite(addr uint16, data *uint32) {
	p.CSRs[addr] = *data
}

func (p *CPU) Execute(ops *Ops) {
	instructions[ops.Name](p, ops)
}
