package main

import (
	"fmt"
)

var regName = [...]string{
	"zero",
	"ra",
	"sp",
	"gp",
	"tp",
	"t0",
	"t1",
	"t2",
	"s0",
	"s1",
	"a0",
	"a1",
	"a2",
	"a3",
	"a4",
	"a5",
	"a6",
	"a7",
	"s2",
	"s3",
	"s4",
	"s5",
	"s6",
	"s7",
	"s8",
	"s9",
	"s10",
	"s11",
	"t3",
	"t4",
	"t5",
	"t6",
}

var csrName = map[int]string{
	0xf11: "mvenorid",
	0xf12: "marchid",
	0xf13: "mimpid",
	0xf14: "mhartid",
	0x300: "mstatus",
	0x301: "misa",
	0x302: "medeleg",
	0x303: "mideleg",
	0x304: "mie",
	0x305: "mtvec",
	0x306: "mcounteren",
	0x341: "mepc",
	0x342: "mcause",
	0x343: "mtval",
	0x344: "mip",
}

var disasms = map[string]func(ops *Ops, pc uint32) string{
	"lui": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,0x%x", ops.Name, regName[ops.Rd], (ops.Imm >> 12) & 0xfffff)
	},
	"auipc": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,0x%x", ops.Name, regName[ops.Rd], (ops.Imm >> 12) & 0xfffff)
	},
	"jal": func(ops *Ops, pc uint32) string {
		if ops.Rd == 0 {
			return fmt.Sprintf("j\t%08x", pc+ops.Imm)
		} else {
			return fmt.Sprintf("%v\t%v,%08x", ops.Name, regName[ops.Rd], pc+ops.Imm)
		}
	},
	"jalr": func(ops *Ops, pc uint32) string {
		if ops.Rs1 == 0 && ops.Imm == 0 {
			return fmt.Sprintf("jr\t%v", regName[ops.Rs1])
		} else {
			return fmt.Sprintf("%v\t%v,%d(%v)", ops.Name, regName[ops.Rd], int32(ops.Imm), regName[ops.Rs1])
		}
	},
	"beq": func(ops *Ops, pc uint32) string {
		if ops.Rs2 == 0 {
			return fmt.Sprintf("beqz\t%v,%x", regName[ops.Rs1], pc+ops.Imm)
		} else {
			return fmt.Sprintf("%v\t%v,%v,%x", ops.Name, regName[ops.Rs1], regName[ops.Rs2], pc+ops.Imm)
		}
	},
	"bne": func(ops *Ops, pc uint32) string {
		if ops.Rs2 == 0 {
			return fmt.Sprintf("bnez\t%v,%x", regName[ops.Rs1], pc+ops.Imm)
		} else {
			return fmt.Sprintf("%v\t%v,%v,%x", ops.Name, regName[ops.Rs1], regName[ops.Rs2], pc+ops.Imm)
		}
	},
	"blt": func(ops *Ops, pc uint32) string {
		if ops.Rs2 == 0 {
			return fmt.Sprintf("bltz\t%v,%x", regName[ops.Rs1], pc+ops.Imm)
		} else {
			return fmt.Sprintf("%v\t%v,%v,%x", ops.Name, regName[ops.Rs1], regName[ops.Rs2], pc+ops.Imm)
		}
	},
	"bge": func(ops *Ops, pc uint32) string {
		if ops.Rs1 == 0 {
			return fmt.Sprintf("blez\t%v,%x", regName[ops.Rs2], pc+ops.Imm)
		} else if ops.Rs2 == 0 {
			return fmt.Sprintf("bgez\t%v,%x", regName[ops.Rs1], pc+ops.Imm)
		} else {
			return fmt.Sprintf("%v\t%v,%v,%x", ops.Name, regName[ops.Rs1], regName[ops.Rs2], pc+ops.Imm)
		}
	},
	"bltu": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%x", ops.Name, regName[ops.Rs1], regName[ops.Rs2], pc+ops.Imm)
	},
	"bgeu": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%x", ops.Name, regName[ops.Rs1], regName[ops.Rs2], pc+ops.Imm)
	},
	"lb": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%d(%v)", ops.Name, regName[ops.Rd], int32(ops.Imm), regName[ops.Rs1])
	},
	"lh": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%d(%v)", ops.Name, regName[ops.Rd], int32(ops.Imm), regName[ops.Rs1])
	},
	"lw": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%d(%v)", ops.Name, regName[ops.Rd], int32(ops.Imm), regName[ops.Rs1])
	},
	"lbu": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%d(%v)", ops.Name, regName[ops.Rd], int32(ops.Imm), regName[ops.Rs1])
	},
	"lhu": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%d(%v)", ops.Name, regName[ops.Rd], int32(ops.Imm), regName[ops.Rs1])
	},
	"sb": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%d(%v)", ops.Name, regName[ops.Rd], int32(ops.Imm), regName[ops.Rs1])
	},
	"sh": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%d(%v)", ops.Name, regName[ops.Rd], int32(ops.Imm), regName[ops.Rs1])
	},
	"sw": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%d(%v)", ops.Name, regName[ops.Rs2], int32(ops.Imm), regName[ops.Rs1])
	},
	"addi": func(ops *Ops, pc uint32) string {
		if ops.Rs1 == 0 {
			if ops.Rd == 0 && ops.Imm == 0 {
				return "nop"
			} else {
				return fmt.Sprintf("li\t%v,%v", regName[ops.Rd], int32(ops.Imm))
			}
		} else {
			return fmt.Sprintf("%v\t%v,%v,%d", ops.Name, regName[ops.Rd], regName[ops.Rs1], int32(ops.Imm))
		}
	},
	"slti": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%d", ops.Name, regName[ops.Rd], regName[ops.Rs1], int32(ops.Imm))
	},
	"sltiu": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%d", ops.Name, regName[ops.Rd], regName[ops.Rs1], ops.Imm)
	},
	"xori": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%d", ops.Name, regName[ops.Rd], regName[ops.Rs1], int32(ops.Imm))
	},
	"ori": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%d", ops.Name, regName[ops.Rd], regName[ops.Rs1], int32(ops.Imm))
	},
	"andi": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%d", ops.Name, regName[ops.Rd], regName[ops.Rs1], int32(ops.Imm))
	},
	"slli": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,0x%x", ops.Name, regName[ops.Rd], regName[ops.Rs1], ops.Shamt)
	},
	"srli": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,0x%x", ops.Name, regName[ops.Rd], regName[ops.Rs1], ops.Shamt)
	},
	"srai": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,0x%x", ops.Name, regName[ops.Rd], regName[ops.Rs1], ops.Shamt)
	},
	"add": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], regName[ops.Rs1], regName[ops.Rs2])
	},
	"sub": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], regName[ops.Rs1], regName[ops.Rs2])
	},
	"sll": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], regName[ops.Rs1], regName[ops.Rs2])
	},
	"slt": func(ops *Ops, pc uint32) string {
		if ops.Rs2 == 0 {
			return fmt.Sprintf("sltz\t%v,%v", regName[ops.Rd], regName[ops.Rs1])
		} else {
			return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], regName[ops.Rs1], regName[ops.Rs2])
		}
	},
	"sltu": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], regName[ops.Rs1], regName[ops.Rs2])
	},
	"xor": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], regName[ops.Rs1], regName[ops.Rs2])
	},
	"srl": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], regName[ops.Rs1], regName[ops.Rs2])
	},
	"sra": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], regName[ops.Rs1], regName[ops.Rs2])
	},
	"or": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], regName[ops.Rs1], regName[ops.Rs2])
	},
	"and": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], regName[ops.Rs1], regName[ops.Rs2])
	},
	"fence": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v", ops.Name)
	},
	"fence_i": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v", ops.Name)
	},
	"ecall": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v", ops.Name)
	},
	"ebreak": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v", ops.Name)
	},
	"mret": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v", ops.Name)
	},
	"csrrw": func(ops *Ops, pc uint32) string {
		if ops.Rd == 0 {
			return fmt.Sprintf("csrw\t%v,%v", toCsrName(ops.Csr), regName[ops.Rs1])
		} else {
			return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], toCsrName(ops.Csr), regName[ops.Rs1])
		}
	},
	"csrrs": func(ops *Ops, pc uint32) string {
		if ops.Rs1 == 0 {
			return fmt.Sprintf("csrr\t%v,%v", regName[ops.Rd], toCsrName(ops.Csr))
		} else if ops.Rd == 0 {
			return fmt.Sprintf("csrs\t%v,%v", toCsrName(ops.Csr), regName[ops.Rs1])
		} else {
			return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], toCsrName(ops.Csr), regName[ops.Rs1])
		}
	},
	"csrrc": func(ops *Ops, pc uint32) string {
		if ops.Rd == 0 {
			return fmt.Sprintf("csrc\t%v,%v", toCsrName(ops.Csr), regName[ops.Rs1])
		} else {
			return fmt.Sprintf("%v\t%v,%v,%v", ops.Name, regName[ops.Rd], toCsrName(ops.Csr), regName[ops.Rs1])
		}
	},
	"csrrwi": func(ops *Ops, pc uint32) string {
		if ops.Rd == 0 {
			return fmt.Sprintf("csrwi\t%v,%v", toCsrName(ops.Csr), ops.Rs1)
		} else {
			return fmt.Sprintf("%v\t%v,%v,%d", ops.Name, regName[ops.Rd], toCsrName(ops.Csr), ops.Rs1)
		}
	},
	"csrrsi": func(ops *Ops, pc uint32) string {
		if ops.Rd == 0 {
			return fmt.Sprintf("csrsi\t%v,%v", toCsrName(ops.Csr), ops.Rs1)
		} else {
			return fmt.Sprintf("%v\t%v,%v,%d", ops.Name, regName[ops.Rd], toCsrName(ops.Csr), ops.Rs1)
		}
	},
	"csrrci": func(ops *Ops, pc uint32) string {
		if ops.Rd == 0 {
			return fmt.Sprintf("csrci\t%v,%v", toCsrName(ops.Csr), ops.Rs1)
		} else {
			return fmt.Sprintf("%v\t%v,%v,%d", ops.Name, regName[ops.Rd], toCsrName(ops.Csr), ops.Rs1)
		}
	},
	"illegal_instruction": func(ops *Ops, pc uint32) string {
		return fmt.Sprintf("%v", ops.Name)
	},
}

func toCsrName(addr uint32) string {
	if v, ok := csrName[int(addr)]; ok {
		return v
	}
	return fmt.Sprintf("csr_0x%x", addr)
}

func disasm(pc uint32, inst uint32, ops *Ops) {
	instStr := disasms[ops.Name](ops, pc)
	info := fmt.Sprintf("%8x:\t%08x\t%v", pc, inst, instStr)
	fmt.Println(info)
}
