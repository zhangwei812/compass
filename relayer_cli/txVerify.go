package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"time"
)

var (
	RouterContractAddress     = config.RouterContractAddress
	RouterContractAddress_map = config.RouterContractAddress_map
	EventSwapOutHash          = crypto.Keccak256Hash([]byte("LogSwapOut(uint256,address,address,address,uint256,uint256,uint256)"))
	currentVerityNum          = config.StartVerityNum // 开始验证区块
)

func init() {
	RouterContractAddress = config.RouterContractAddress
	RouterContractAddress_map = config.RouterContractAddress_map
	currentVerityNum = config.StartVerityNum // 开始验证区块
}
func (d *commpassInfo) doTxVerity() {
	for {
		//------验证,开始块 ------
		num, _ := getCurrentNumberAbi(d.client, ChainTypeETH, d.relayerData[0].from)
		if num > currentVerityNum {
			d.doTxVerity1(currentVerityNum, num)
			//currentVerityNum = num
		}
		time.Sleep(time.Second)
	}
}

func (d *commpassInfo) doTxVerity1(fromBlock uint64, toBlock uint64) {
	fmt.Println("=================DO TxVerity========================")
	EthConn, _ := dialEthConn()
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{common.HexToAddress(RouterContractAddress)},
	}
	logs, err := EthConn.FilterLogs(context.Background(), query)
	if err != nil {
		panic(err)
	}
	for _, aLog := range logs {
		if EventSwapOutHash != aLog.Topics[0] {
			continue
		}
		d.HandleLogSwapOut(&aLog, EthConn)
	}
}

func (d *commpassInfo) HandleLogSwapOut(aLog *types.Log, ethConn *ethclient.Client) {
	err := abiRouter.UnpackIntoInterface(&eventResponse, "LogSwapOut", aLog.Data)
	if err != nil {
		panic(err)
	}
	txProve := GetTxProve(*ethConn, aLog, &eventResponse)

	token := common.BytesToAddress(aLog.Topics[1].Bytes())
	to := common.BytesToAddress(aLog.Topics[3].Bytes())
	conn := d.client
	input := packInput(abiRouter, "swapIn",
		eventResponse.OrderId,
		token,
		to,
		eventResponse.Amount,
		eventResponse.FromChainID,
		aLog.Address,
		txProve)

	//input, err := abiRouter.Pack("swapIn",
	//	big.NewInt(0),
	//	token,
	//	aLog.Address,
	//	big.NewInt(1),
	//	eventResponse.FromChainID,
	//	//eventResponse.ToChainID,
	//	common.HexToAddress(RouterContractAddress),
	//	txProve)
	//input1, _ := abiRouter.Pack("txVerify",
	//	aLog.Address,
	//	token,
	//	eventResponse.FromChainID,
	//	eventResponse.ToChainID,
	//	txProve)

	if err != nil {
		log.Fatal(abiRouter, " error ", err)
	}
	relayer := d.relayerData[0]
	RouterContractAddress_map1 := common.HexToAddress(RouterContractAddress_map)
	b := sendContractTransaction(conn, relayer.from, RouterContractAddress_map1, nil, relayer.priKey, input)
	fmt.Println("TxVerity result:", b, "   eth blockNumber ", aLog.BlockNumber, "  transactionIndex: ", aLog.TxIndex)
}
