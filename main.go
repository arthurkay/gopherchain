package main

import (
	"flag"
	"fmt"
	"gopherchain/chain"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct {
	blockchain *chain.BlockChain
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage: ")
	fmt.Println(" add -block BLOCK_DATA - Adds a block to the blockchain")
	fmt.Print(" print - Prints the blocks in the chain")
}

func (cli *CommandLine) addBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("Added block")
}

func (cli *CommandLine) printChain() {
	iter := cli.blockchain.Iterator()

	for {
		block := iter.Next()
		fmt.Printf("Previous hash: %X \n", block.PrevHash)
		fmt.Printf("Data in current block: %s \n", block.Data)
		fmt.Printf("Current Block Hash: %X \n", block.Hash)

		pow := chain.NewProof(block)
		fmt.Printf("Proof Of Work: %s \n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "Block Data")

	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		chain.HandleErr(err)

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		chain.HandleErr(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func main() {
	defer os.Exit(0)
	fmt.Print("Gopherchain started \n")

	link := chain.InitBlockChain()
	defer link.Database.Close()

	cli := CommandLine{link}
	cli.run()

}
