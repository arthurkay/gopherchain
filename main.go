package main

import (
	"fmt"
	"gopherchain/cli"
	"os"
)

func main() {
	defer os.Exit(0)
	fmt.Print("Gopherchain started\n")
	cli := cli.CommandLine{}
	cli.Run()

}
