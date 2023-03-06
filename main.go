package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"os"
	"os/signal"
	"sync"
)

const (
	rawWSURL  = "wss://exchainws.okex.org:8443"
	rawRPCURL = "https://exchainrpc.okex.org"
)

var wsMode bool

func main() {
	//
	if os.Args[1] == "ws" {
		wsMode = true
	}

	//open rpc connection
	rpcClient, err := ethclient.Dial(rawRPCURL)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		//open ws connection
		//errChan, err := wsConnect(rawWSURL, rpcClient)
		//errChan, err := ethwsconn(rawWSURL, rpcClient)
		var errChan chan error
		var err error
		if wsMode {
			errChan, err = ethwsconn2(rawWSURL, rpcClient)
		} else {
			errChan, err = ethwsconn(rawWSURL, rpcClient)
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		for {
			select {
			case <-errChan:
				fmt.Fprintln(os.Stderr, err)
				wg.Done()
				return
			case <-done:
				fmt.Println("Quit routine:")
				wg.Done()
				return
			}
		}
	}()

	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt)

	_ = <-interrupt
	// We received a SIGINT (Ctrl + C). Terminate gracefully...
	log.Println("Received SIGINT interrupt signal. Closing all pending connections")
	close(done)

}
