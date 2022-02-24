package main

import (
	"mpc/cmd"
	"os"
)

func main() {
	rootCMD := cmd.NewRootCMD()
	if err := rootCMD.Execute(); err != nil {
		rootCMD.PrintErrln("Error:", err)
		os.Exit(1)
	}
}
