package main

import (
	"fmt"
)

type Bus struct {
	mem  *Mem
}

var _ = fmt.Println

func NewBus() *Bus {
	mem := NewMem()
	return &Bus{mem}
}

/*
 * - Reserved : 0x00040000 - 0xffffffff
 * - Program  : 0x00000000 - 0x0003ffff
 */

func (p *Bus) WriteByte(addr uint32, data uint8) {
	if (0x00000000 <= addr) && (addr < 0x00040000) {
		t := addr
		p.mem.WriteByte(t, data)
	}
}

func (p *Bus) WriteHalf(addr uint32, data uint16) {
	if (0x00000000 <= addr) && (addr < 0x00040000) {
		t := addr
		p.mem.WriteHalf(t, data)
	}
}

func (p *Bus) WriteWord(addr uint32, data uint32) {
	if (0x00000000 <= addr) && (addr < 0x00040000) {
		t := addr
		p.mem.WriteWord(t, data)
	}
}

func (p *Bus) ReadByte(addr uint32) uint8 {
	var ret uint8 = 0

	if (0x00000000 <= addr) && (addr < 0x00040000) {
		t := addr
		ret = p.mem.ReadByte(t)
	}
	
	return ret
}

func (p *Bus) ReadHalf(addr uint32) uint16 {
	var ret uint16 = 0
	
	if (0x00000000 <= addr) && (addr < 0x00040000) {
		t := addr
		ret = p.mem.ReadHalf(t)
	}
	
	return ret
}

func (p *Bus) ReadWord(addr uint32) uint32 {
	var ret uint32 = 0
	
	if (0x00000000 <= addr) && (addr < 0x00040000) {
		t := addr
		ret = p.mem.ReadWord(t)
	}
	
	return ret
}
