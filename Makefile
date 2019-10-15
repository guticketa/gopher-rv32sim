SRC := main.go cpu.go mem.go uart.go bus.go disasm.go

all: build

build: $(SRC)
	go build -o gopher-rv32sim $^

run: $(SRC)
	go run $^

.PHONY: clean
clean:
	@$(RM) gopher-rv32sim

