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
func sendTransation(ctx *cli.Context) error {
	initCfg(ctx)
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
	count := 0
	transStartTime = time.Now().Format("2006/1/2 15:04:05")
	for {
		fmt.Println()
		count++
		log.Info("================= sendTransation to Eth========================", "Number", count)
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
		log.Info("sendTransationOnEth", "from:", relayer.from.String(), "   to:", DefaultTransactionAddress, "  amount:", amount)
		b := sendContractTransaction(EthConn, relayer.from, RouterContractAddress1, nil, relayer.priKey, input)
		if !b {
			transFailNum++
			log.Error("sendTransationOnEth err")
		} else {
			transSuccessNum++
		}

		log.Info("sendTransationOnEth", "开始时间：", transStartTime, "交易数量:", transFailNum+transSuccessNum, "成功次数:", transSuccessNum, "失败次数:", transFailNum)
		log.Info("waiting next time(Once an hour) to sendTranstion............")
		// 一个小时转一次
		time.Sleep(3600 * time.Second)
	}
}
