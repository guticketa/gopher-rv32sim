package main

type Mem struct {
	mem []uint8
}

func NewMem() *Mem {
	mem := make([]uint8, 1024*1024)
	return &Mem{mem}
}

func (p *Mem) ReadByte(addr uint32) uint8 {
	return p.mem[addr]
}

func (p *Mem) ReadHalf(addr uint32) uint16 {
	maskAddr := addr & 0xfffffffe
	b0 := uint16(p.mem[maskAddr+0] & 0xff)
	b1 := uint16(p.mem[maskAddr+1] & 0xff)
	return (b1 << 8) | (b0)
}

func (p *Mem) ReadWord(addr uint32) uint32 {
	maskAddr := addr & 0xfffffffc
	b0 := uint32(p.mem[maskAddr+0] & 0xff)
	b1 := uint32(p.mem[maskAddr+1] & 0xff)
	b2 := uint32(p.mem[maskAddr+2] & 0xff)
	b3 := uint32(p.mem[maskAddr+3] & 0xff)
	return (b3 << 24) | (b2 << 16) | (b1 << 8) | (b0)
}

func (p *Mem) WriteByte(addr uint32, data uint8) {
	p.mem[addr] = data
}

func (p *Mem) WriteHalf(addr uint32, data uint16) {
	maskAddr := addr & 0xfffffffe
	b0 := uint8((data >> 0) & 0xff)
	b1 := uint8((data >> 8) & 0xff)
	p.mem[maskAddr+0] = b0
	p.mem[maskAddr+1] = b1
}

func (p *Mem) WriteWord(addr uint32, data uint32) {
	maskAddr := addr & 0xfffffffc
	b0 := uint8((data >> 0) & 0xff)
	b1 := uint8((data >> 8) & 0xff)
	b2 := uint8((data >> 16) & 0xff)
	b3 := uint8((data >> 24) & 0xff)
	p.mem[maskAddr+0] = b0
	p.mem[maskAddr+1] = b1
	p.mem[maskAddr+2] = b2
	p.mem[maskAddr+3] = b3
}
