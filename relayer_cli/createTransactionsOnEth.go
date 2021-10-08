package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"time"
)

var (
	DefaultTransactionAddress = config.DefaultTransactionAddress
	DefaultAmount             = config.DefaultAmount
	transSuccessNum           = 0
	transFailNum              = 0
	transStartTime            = ""
)

func init4() {
	DefaultTransactionAddress = config.DefaultTransactionAddress
	DefaultAmount = config.DefaultAmount
}
func sendTransaction(ctx *cli.Context) error {
	initCfg(ctx)
	compassInfo := compassInfo{}
	compassInfo.relayerData = []*relayerInfo{
		{url: keystore1},
	}
	compassInfo.preWork(ctx)
	// 创造交易
	compassInfo.sendTransationOnEth()
	return nil
}

func (d *compassInfo) sendTransationOnEth() {
	count := 0
	transStartTime = time.Now().Format("2006/1/2 15:04:05")
	for {
		fmt.Println()
		count++
		log.Info("================= sendTransaction to Eth========================", "Number", count)
		token := common.HexToAddress(Erc20ContractAddress)
		to := common.HexToAddress(DefaultTransactionAddress)
		amount := big.NewInt(int64(DefaultAmount))
		EthConn, _ := dialEthConn()
		input := packInput(abiRouter, "swapOut",
			token,
			to,
			amount,
			big.NewInt(int64(ChainTypeMAPTemp)))
		relayer := d.relayerData[0]
		RouterContractAddress1 := common.HexToAddress(RouterContractAddress)
		log.Info("sendTransactionOnEth", "from", relayer.from.String(), "   to", DefaultTransactionAddress, "  amount", amount)
		b := sendContractTransaction(EthConn, relayer.from, RouterContractAddress1, nil, relayer.priKey, input)
		if !b {
			transFailNum++
			log.Error("sendTransactionOnEth err")
		} else {
			transSuccessNum++
		}

		log.Info("sendTransactionOnEth", "transStartTime", transStartTime, "transaction count", transFailNum+transSuccessNum, "success", transSuccessNum, "fail", transFailNum)
		log.Info("waiting next time(Once an hour) to sendTranstion............")
		// 一个小时转一次
		time.Sleep(3600 * time.Second)
	}
}
