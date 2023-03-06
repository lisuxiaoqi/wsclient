package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"time"
)

func ethwsconn2(wsURL string, rpcClient *ethclient.Client) (chan error, error) {
	fmt.Println("---Start with ws as the rpc")
	c, err := rpc.DialContext(context.Background(), wsURL)
	if err != nil {
		utils.Fatalf("Failed to connect to Ethereum node: %v", err)
		return nil, err
	}
	wsc := ethclient.NewClient(c)

	//client, err := ethclient.Dial(wsURL)
	//if err != nil {
	//	utils.Fatalf("Failed to connect to Ethereum node: %v", err)
	//	return nil, err
	//}

	//headers := make(chan *types.Header)
	//sub, err := client.SubscribeNewHead(context.Background(), headers)

	okcHeaders := make(chan *OKCHeader)
	sub, err := c.EthSubscribe(context.Background(), okcHeaders, "newHeads")
	if err != nil {
		utils.Fatalf("SubscribeNewHead error:%v", err)
		return nil, err
	}

	errChan := make(chan error)
	go func() {
		for {
			select {
			case err := <-sub.Err():
				utils.Fatalf("SubscribeNewHead error:%v", err)
				errChan <- err
				return
			case header := <-okcHeaders:
				//fmt.Println(header.ParentHash.Hex())
				fmt.Println(time.Now(), "WS Get header:", header.Number, "Hash:", header.Hash.Hex())

				//getBlockByHash
				block, err := wsc.BlockByHash(context.Background(), header.Hash)
				if err != nil {
					fmt.Printf("BlockByNumber error:%v\n", err)
					continue
				} else {
					fmt.Println("WS RPC get block by Hash:", block.Number(), "txs:", block.Transactions().Len())
				}

				if block.Transactions() == nil || block.Transactions().Len() == 0 {
					if (header.TxHash != common.Hash{}) {
						fmt.Println("Get Block but empty tx,  error", block.Number())
						continue
					}
				}

				//getTxReceipt
				for _, tx := range block.Transactions() {
					//fmt.Println("Raw txHash", tx.Hash().Hex())
					_, err := wsc.TransactionReceipt(context.Background(), tx.Hash())
					if err != nil {
						fmt.Printf("Get Tx Receipt error:%v\n", err)
					} else {
						//fmt.Println("tx Hash in receipt", receipt.TxHash.Hex())
					}

				}
			}
		}
	}()

	return errChan, nil
}
