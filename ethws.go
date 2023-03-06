package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
	"time"
)

type OKCHeader struct {
	//types.Header
	ParentHash  common.Hash      `json:"parentHash"       gencodec:"required"`
	UncleHash   common.Hash      `json:"sha3Uncles"       gencodec:"required"`
	Hash        common.Hash      `json:"hash"  gencodec:"required"`
	Coinbase    common.Address   `json:"miner"`
	Root        common.Hash      `json:"stateRoot"        gencodec:"required"`
	TxHash      common.Hash      `json:"transactionsRoot" gencodec:"required"`
	ReceiptHash common.Hash      `json:"receiptsRoot"     gencodec:"required"`
	Bloom       types.Bloom      `json:"logsBloom"        gencodec:"required"`
	Difficulty  *big.Int         `json:"difficulty"       gencodec:"required"`
	Number      *big.Int         `json:"number"           gencodec:"required"`
	GasLimit    uint64           `json:"gasLimit"         gencodec:"required"`
	GasUsed     uint64           `json:"gasUsed"          gencodec:"required"`
	Time        uint64           `json:"timestamp"        gencodec:"required"`
	Extra       []byte           `json:"extraData"        gencodec:"required"`
	MixDigest   common.Hash      `json:"mixHash"`
	Nonce       types.BlockNonce `json:"nonce"`

	// BaseFee was added by EIP-1559 and is ignored in legacy headers.
	BaseFee *big.Int `json:"baseFeePerGas" rlp:"optional"`
}

// UnmarshalJSON unmarshals from JSON.
func (h *OKCHeader) UnmarshalJSON(input []byte) error {
	type Header struct {
		Hash        *common.Hash      `json:"hash"       gencodec:"required"`
		ParentHash  *common.Hash      `json:"parentHash"       gencodec:"required"`
		UncleHash   *common.Hash      `json:"sha3Uncles"       gencodec:"required"`
		Coinbase    *common.Address   `json:"miner"`
		Root        *common.Hash      `json:"stateRoot"        gencodec:"required"`
		TxHash      *common.Hash      `json:"transactionsRoot" gencodec:"required"`
		ReceiptHash *common.Hash      `json:"receiptsRoot"     gencodec:"required"`
		Bloom       *types.Bloom      `json:"logsBloom"        gencodec:"required"`
		Difficulty  *hexutil.Big      `json:"difficulty"       gencodec:"required"`
		Number      *hexutil.Big      `json:"number"           gencodec:"required"`
		GasLimit    *hexutil.Uint64   `json:"gasLimit"         gencodec:"required"`
		GasUsed     *hexutil.Uint64   `json:"gasUsed"          gencodec:"required"`
		Time        *hexutil.Uint64   `json:"timestamp"        gencodec:"required"`
		Extra       *hexutil.Bytes    `json:"extraData"        gencodec:"required"`
		MixDigest   *common.Hash      `json:"mixHash"`
		Nonce       *types.BlockNonce `json:"nonce"`
		BaseFee     *hexutil.Big      `json:"baseFeePerGas" rlp:"optional"`
	}
	var dec Header
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.ParentHash == nil {
		return errors.New("missing required field 'parentHash' for Header")
	}
	h.ParentHash = *dec.ParentHash
	if dec.UncleHash == nil {
		return errors.New("missing required field 'sha3Uncles' for Header")
	}
	h.UncleHash = *dec.UncleHash

	if dec.Hash == nil {
		return errors.New("missing required field 'Hash' for Header")
	}
	h.Hash = *dec.Hash

	if dec.Coinbase != nil {
		h.Coinbase = *dec.Coinbase
	}
	if dec.Root == nil {
		return errors.New("missing required field 'stateRoot' for Header")
	}
	h.Root = *dec.Root
	if dec.TxHash == nil {
		return errors.New("missing required field 'transactionsRoot' for Header")
	}
	h.TxHash = *dec.TxHash
	if dec.ReceiptHash == nil {
		return errors.New("missing required field 'receiptsRoot' for Header")
	}
	h.ReceiptHash = *dec.ReceiptHash
	if dec.Bloom == nil {
		return errors.New("missing required field 'logsBloom' for Header")
	}
	h.Bloom = *dec.Bloom
	if dec.Difficulty == nil {
		return errors.New("missing required field 'difficulty' for Header")
	}
	h.Difficulty = (*big.Int)(dec.Difficulty)
	if dec.Number == nil {
		return errors.New("missing required field 'number' for Header")
	}
	h.Number = (*big.Int)(dec.Number)
	if dec.GasLimit == nil {
		return errors.New("missing required field 'gasLimit' for Header")
	}
	h.GasLimit = uint64(*dec.GasLimit)
	if dec.GasUsed == nil {
		return errors.New("missing required field 'gasUsed' for Header")
	}
	h.GasUsed = uint64(*dec.GasUsed)
	if dec.Time == nil {
		return errors.New("missing required field 'timestamp' for Header")
	}
	h.Time = uint64(*dec.Time)
	if dec.Extra == nil {
		return errors.New("missing required field 'extraData' for Header")
	}
	h.Extra = *dec.Extra
	if dec.MixDigest != nil {
		h.MixDigest = *dec.MixDigest
	}
	if dec.Nonce != nil {
		h.Nonce = *dec.Nonce
	}
	if dec.BaseFee != nil {
		h.BaseFee = (*big.Int)(dec.BaseFee)
	}
	return nil
}

func ethwsconn(wsURL string, rpcClient *ethclient.Client) (chan error, error) {
	fmt.Println("---Start with normal rpc")

	c, err := rpc.DialContext(context.Background(), wsURL)
	if err != nil {
		utils.Fatalf("Failed to connect to Ethereum node: %v", err)
		return nil, err
	}
	//client := ethclient.NewClient(c)

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
				block, err := rpcClient.BlockByHash(context.Background(), header.Hash)
				if err != nil {
					fmt.Printf("BlockByNumber error:%v\n", err)
					continue
				} else {
					fmt.Println("RPC get block by Hash:", block.Number(), "txs:", block.Transactions().Len())
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
					_, err := rpcClient.TransactionReceipt(context.Background(), tx.Hash())
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
