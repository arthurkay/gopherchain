package cli

import (
	"flag"
	"fmt"
	"gopherchain/chain"
	"gopherchain/wallet"
	"log"
	"os"
	"runtime"
	"strconv"
)

type CommandLine struct{}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage: ")
	fmt.Println(" balance --addr ADDRESS get the account balance")
	fmt.Println(" createblockchain --addr ADDRESS creates a blockchain")
	fmt.Println(" print - Prints the blocks in the chain")
	fmt.Println(" send --from FROM --to TO --amount AMOUNT Send money from FROM to TO")
	fmt.Println(" new-wallet - Creates a new wallet")
	fmt.Println(" list-addr - Lists the addresses in our wallet file")
}

func (cli *CommandLine) printChain() {
	myChain := chain.ContinueBlockChain("")
	defer myChain.Database.Close()

	iter := myChain.Iterator()

	for {
		block := iter.Next()
		fmt.Printf("Previous hash: %X \n", block.PrevHash)
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

func (cli *CommandLine) listAddresses() {
	wallets, _ := wallet.CreateWallets()
	addresses := wallets.GetAllAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

func (cli *CommandLine) createWallet() {
	wallets, _ := wallet.CreateWallets()
	address := wallets.AddWallet()
	wallets.SaveFile()

	fmt.Printf("New address is: %s\n", address)
}

func (cli *CommandLine) createBlockChain(address string) {
	myChain := chain.InitBlockChain(address)
	myChain.Database.Close()
	fmt.Println("Finished")
}

func (cli *CommandLine) getBalance(address string) {
	myChain := chain.ContinueBlockChain(address)
	defer myChain.Database.Close()

	balance := 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := myChain.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s : %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	if !wallet.ValidateAddress(to) {
		log.Panic("Address not valid")
	}

	if !wallet.ValidateAddress(from) {
		log.Panic("Address not valid")
	}

	log.Printf("Sending %d to %s", amount, to)
	myChain := chain.ContinueBlockChain(from)
	defer myChain.Database.Close()

	tx := chain.NewTransaction(from, to, amount, myChain)
	myChain.AddBlock([]*chain.Transaction{tx})
	fmt.Println("Success")
}

func (cli *CommandLine) Run() {
	cli.validateArgs()
	getBalanceCmd := flag.NewFlagSet("balance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createBlockChain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("list-addresses", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("create-wallet", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("addr", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("addr", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "balance":
		err := getBalanceCmd.Parse(os.Args[2:])
		chain.HandleErr(err)

	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "list-addr":
		err := listAddressesCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	case "new-wallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}
