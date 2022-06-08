package main

import (
	"os"

	"github.com/zhufuyi/mpc/cmd"
)

func main() {
	rootCMD := cmd.NewRootCMD()
	if err := rootCMD.Execute(); err != nil {
		rootCMD.PrintErrln("Error:", err)
		os.Exit(1)
	}
}
