package main

import (
	"fmt"
)

type Bus struct {
	mem  *Mem
	uart *UART
}

const (
	uartBase = 0x20000000
	uartTop  = 0x20000fff
	ramBase = 0x80000000
	ramTop  = 0x800fffff
)

var _ = fmt.Println

func NewBus() *Bus {
	mem := NewMem()
	uart := NewUART()
	return &Bus{mem, uart}
}

// Memory Map
// - Reserved : 0x00000000 - 0x1fffffff
// - UART     : 0x20000000 - 0x20000fff
// - Reserved : 0x20001000 - 0x7fffffff
// - Program  : 0x80000000 - 0x800fffff
// - Reserved : 0x80100000 - 0xffffffff

func (p *Bus) WriteByte(addr uint32, data uint8) {
	if (ramBase <= addr) && (addr <= ramTop) {
		t := addr - ramBase
		p.mem.WriteByte(t, data)
	} else if (uartBase <= addr) && (addr <= uartTop) {
		t := addr - uartBase
		p.uart.WriteByte(t, data)
	}
}

func (p *Bus) WriteHalf(addr uint32, data uint16) {
	if (ramBase <= addr) && (addr <= ramTop) {
		t := addr - ramBase
		p.mem.WriteHalf(t, data)
	} else if (uartBase <= addr) && (addr <= uartTop) {
		t := addr - uartBase
		p.uart.WriteHalf(t, data)
	}
}

func (p *Bus) WriteWord(addr uint32, data uint32) {
	if (ramBase <= addr) && (addr <= ramTop) {
		t := addr - ramBase
		p.mem.WriteWord(t, data)
	} else if (uartBase <= addr) && (addr <= uartTop) {
		t := addr - uartBase
		p.uart.WriteWord(t, data)
	}
}

func (p *Bus) ReadByte(addr uint32) uint8 {
	var ret uint8 = 0

	if (ramBase <= addr) && (addr <= ramTop) {
		t := addr - ramBase
		ret = p.mem.ReadByte(t)
	} else if (uartBase <= addr) && (addr <= uartTop) {
		t := addr - uartBase
		ret = p.uart.ReadByte(t)
	}
	
	return ret
}

func (p *Bus) ReadHalf(addr uint32) uint16 {
	var ret uint16 = 0
	
	if (ramBase <= addr) && (addr <= ramTop) {
		t := addr - ramBase
		ret = p.mem.ReadHalf(t)
	} else if (uartBase <= addr) && (addr <= uartTop) {
		t := addr - uartBase
		ret = p.uart.ReadHalf(t)
	}
	
	return ret
}

func (p *Bus) ReadWord(addr uint32) uint32 {
	var ret uint32 = 0
	
	if (ramBase <= addr) && (addr <= ramTop) {
		t := addr - ramBase
		ret = p.mem.ReadWord(t)
	} else if (uartBase <= addr) && (addr <= uartTop) {
		t := addr - uartBase
		ret = p.uart.ReadWord(t)
	}
	
	return ret
}
