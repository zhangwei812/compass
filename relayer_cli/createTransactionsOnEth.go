package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"time"
)

var (
	DefaultTransactionAddress = config.DefaultTransactionAddress
	DefaultAmount             = config.DefaultAmount
)

func init() {
	DefaultTransactionAddress = config.DefaultTransactionAddress
	DefaultAmount = config.DefaultAmount
}
func sendTransation(ctx *cli.Context) error {
	commpassInfo := commpassInfo{}
	commpassInfo.relayerData = []*relayerInfo{
		{url: keystore1},
	}
	commpassInfo.preWork(ctx)
	// 创造交易
	commpassInfo.sendTransationOnEth()
	return nil
}

func (d *commpassInfo) sendTransationOnEth() {
	for {
		fmt.Println()
		fmt.Println("================= sendTransation to Eth========================")
		token := common.HexToAddress(Erc20ContractAddress)
		to := common.HexToAddress(DefaultTransactionAddress)
		amount := big.NewInt(int64(DefaultAmount))
		EthConn, _ := dialEthConn()
		input := packInput(abiRouter, "swapOut",
			token,
			to,
			amount,
			big.NewInt(int64(ChainTypeMAP)))
		relayer := d.relayerData[0]
		RouterContractAddress1 := common.HexToAddress(RouterContractAddress)
		fmt.Println("from:", relayer.from.String(), "   to:", DefaultTransactionAddress, "  amount:", amount)
		b := sendContractTransaction(EthConn, relayer.from, RouterContractAddress1, nil, relayer.priKey, input)
		if !b {
			panic("sendTransationOnEth err")
		}
		fmt.Println("waiting next time(Once an hour) to sendTranstion............")
		// 一个小时转一次
		time.Sleep(3600 * time.Second)
	}
}
