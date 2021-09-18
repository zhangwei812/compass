package main

import (
	"errors"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
	"math/big"
)

var (
	registerValue          = config.RegisterValue
	Base                   = new(big.Int).SetUint64(10000)
	contractQueryFailedErr = errors.New("Contract query failed result ")
)

func init() {
	registerValue = config.RegisterValue
}

const (
	RegisterAmount = 100000
)

func register(ctx *cli.Context, conn *ethclient.Client, info relayerInfo) {
	value := ethToWei(info.registerValue)
	if info.registerValue < RegisterAmount {
		log.Fatal("Amount must bigger than ", RegisterAmount)
	}
	fee := ctx.GlobalUint64(FeeFlag.Name)
	checkFee(new(big.Int).SetUint64(fee))
	input := packInput(abiRelayer, "register", value)
	sendContractTransaction(conn, info.from, RelayerAddress, nil, info.priKey, input)
}

func checkFee(fee *big.Int) {
	if fee.Sign() < 0 || fee.Cmp(Base) > 0 {
		log.Fatal("Please set correct fee value")
	}
}
