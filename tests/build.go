package main

import (
	"log"

	"github.com/t4ke0/gbs"
)

// buildNormalProgram
func buildNormalProgram() error {
	sh := &gbs.Sh{}
	const command string = "go build -o nprog ./programs/normal/main.go"
	return sh.Init(command).Run().Error()
}

// runNormalProgram
func runNormalProgram() error {
	const command string = "./nprog"
	return new(gbs.Sh).Init(command).Run().Error()
}

// cleanNormalProgram
func cleanNormalProgram() error {
	const command string = "rm nprog"
	return new(gbs.Sh).Init(command).Run().Error()
}

// buildInputProgram
func buildInputProgram() error {
	const command string = "go build -o inprog ./programs/input/main.go"
	return new(gbs.Sh).Init(command).Run().Error()
}

// runInputProgram
func runInputProgram() error {
	const command string = "./inprog"
	return new(gbs.Sh).Init(command).In("hello, world").Error()
}

// cleanInputProgram
func cleanInputProgram() error {
	const command string = "rm inprog"
	return new(gbs.Sh).Init(command).Run().Error()
}

func main() {

	opts := []gbs.BuildFuncOpt{
		gbs.BuildFuncOpt{
			FuncName: "buildNormalProgram",
			Func:     buildNormalProgram,
		},
		gbs.BuildFuncOpt{
			FuncName: "runNormalProgram",
			Func:     runNormalProgram,
		},
		gbs.BuildFuncOpt{
			FuncName: "cleanNormalProgram",
			Func:     cleanNormalProgram,
		},
	}

	if err := gbs.Build(opts...); err != nil {
		log.Fatal(err)
	}

	opts = []gbs.BuildFuncOpt{
		gbs.BuildFuncOpt{"buildInputProgram", buildInputProgram},
		gbs.BuildFuncOpt{"runInputProgram", runInputProgram},
		gbs.BuildFuncOpt{"cleanInputProgram", cleanInputProgram},
	}

	if err := gbs.Build(opts...); err != nil {
		log.Fatal(err)
	}

}
