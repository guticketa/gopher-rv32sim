package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	// "reflect"
)

var verbose = flag.Bool("v", false, "")

func main() {
	flag.Parse()
	
	if flag.NArg() != 1 {
		log.Fatalf("ERROR: %v", errors.New("Argument Error"))
	}
	
	filename := flag.Args()[0]
	sim := NewCPU()
	sim.Reset()
	sim.LoadElf(filename)
	for i := 0; i < 5000; i++ {
		inst := sim.Fetch()
		ops := sim.Decode(inst)
		if *verbose {
			fmt.Printf("0x%08x:  %08x  %s\n", sim.PC, inst, ops.Name)
		}
		sim.Execute(&ops)
	}

	// Result
	if sim.Regs[3] == 1 {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
