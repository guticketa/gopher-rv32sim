package main

import (
	"fmt"
)

// Memory Map:
// 0x000: txdata transmit data register
// 0x008: txctrl transmit control register

type UART struct {
	reg []uint32
}

func NewUART() *UART {
	reg := make([]uint32, 4096)
	return &UART{reg}
}

func (p *UART) ReadByte(addr uint32) uint8 {
	sel := addr & 0x00000003
	maskAddr := (addr & 0xfffffffc) >> 2
	switch sel {
	case 0:
		return uint8((p.reg[maskAddr] & 0x000000ff) >> 0)
	case 1:
		return uint8((p.reg[maskAddr] & 0x0000ff00) >> 8)
	case 2:
		return uint8((p.reg[maskAddr] & 0x00ff0000) >> 16)
	case 3:
		return uint8((p.reg[maskAddr] & 0xff000000) >> 24)
	default:
		return uint8((p.reg[maskAddr] & 0x000000ff) >> 0)
	}
}

func (p *UART) ReadHalf(addr uint32) uint16 {
	sel := addr & 0x00000002
	maskAddr := addr & 0xfffffffc
	var t0, t1 uint16
	switch sel {
	case 2:
		t0 = uint16(p.ReadByte(maskAddr + 2))
		t1 = uint16(p.ReadByte(maskAddr + 3))
	default:
		t0 = uint16(p.ReadByte(maskAddr + 0))
		t1 = uint16(p.ReadByte(maskAddr + 1))
	}

	return (t1 << 8) | t0
}

func (p *UART) ReadWord(addr uint32) uint32 {
	maskAddr := addr & 0xfffffffc
	t0 := uint32(p.ReadByte(maskAddr + 0))
	t1 := uint32(p.ReadByte(maskAddr + 1))
	t2 := uint32(p.ReadByte(maskAddr + 2))
	t3 := uint32(p.ReadByte(maskAddr + 3))
	return (t3 << 24) | (t2 << 16) | (t1 << 8) | t0
}

func (p *UART) WriteByte(addr uint32, data uint8) {
	sel := addr & 0x00000003
	maskAddr := (addr & 0xfffffffc) >> 2
	switch sel {
	case 0:
		p.reg[maskAddr] = (p.reg[maskAddr] & 0xffffff00) | (uint32(data) << 0)
	case 1:
		p.reg[maskAddr] = (p.reg[maskAddr] & 0xffff00ff) | (uint32(data) << 8)
	case 2:
		p.reg[maskAddr] = (p.reg[maskAddr] & 0xff00ffff) | (uint32(data) << 16)
	case 3:
		p.reg[maskAddr] = (p.reg[maskAddr] & 0x00ffffff) | (uint32(data) << 24)
	default:
		p.reg[maskAddr] = (p.reg[maskAddr] & 0xffffff00) | (uint32(data) << 0)
	}

	if (p.reg[2]&0x01 != 0) && (addr == 0) {
		fmt.Printf("%c", data)
	}
}

func (p *UART) WriteHalf(addr uint32, data uint16) {
	sel := addr & 0x00000002
	maskAddr := addr & 0xfffffffc
	switch sel {
	case 2:
		p.WriteByte(maskAddr+2, uint8((data>>0)&0x000000ff))
		p.WriteByte(maskAddr+3, uint8((data>>8)&0x000000ff))
	default:
		p.WriteByte(maskAddr+0, uint8((data>>0)&0x000000ff))
		p.WriteByte(maskAddr+1, uint8((data>>8)&0x000000ff))
	}
}

func (p *UART) WriteWord(addr uint32, data uint32) {
	maskAddr := addr & 0xfffffffc
	p.WriteByte(maskAddr+0, uint8((data>>0)&0x000000ff))
	p.WriteByte(maskAddr+1, uint8((data>>8)&0x000000ff))
	p.WriteByte(maskAddr+2, uint8((data>>16)&0x000000ff))
	p.WriteByte(maskAddr+3, uint8((data>>24)&0x000000ff))
}
