package main

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
	"math/big"
	"time"
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

func (d *commpassInfo) relayerRegister() {
	fmt.Println()
	fmt.Println("--------------------relayerRegister-----------------------------------")
	conn := d.client
	ctx := d.ctx
	for k, _ := range d.relayerData {
		register(ctx, d.client, *d.relayerData[k])
		for {
			bool1, bool2, relayerEpoch, _ := queryRegisterInfo(conn, d.relayerData[k].from)
			fmt.Println("ADDRESS:", d.relayerData[k].from, "ISREGISTER:", bool1, " ISRELAYER :", bool2, " RELAYER_EPOCH :", relayerEpoch)
			if bool2 {
				break
			}
			fmt.Println("waiting to become relayer...................................")
			for {
				time.Sleep(time.Second * 10)
				bool1, bool2, relayerEpoch, _ := queryRegisterInfo(conn, d.relayerData[k].from)
				if bool2 {
					fmt.Println("ADDRESS:", d.relayerData[k].from, "ISREGISTER:", bool1, " ISRELAYER :", bool2, " RELAYER_EPOCH :", relayerEpoch)
					break
				}
			}
			break
		}
	}
}
func (d *commpassInfo) waitBecomeRelayer(info relayerInfo) {
	conn := d.client
	register(d.ctx, d.client, info)
	for {
		bool1, bool2, relayerEpoch, _ := queryRegisterInfo(conn, info.from)
		fmt.Println("ADDRESS:", info.from, "ISREGISTER:", bool1, " ISRELAYER :", bool2, " RELAYER_EPOCH :", relayerEpoch)
		if bool2 {
			break
		}
		fmt.Println("waiting to become relayer...................................")
		for {
			time.Sleep(time.Second * 10)
			_, bool2, _, _ := queryRegisterInfo(conn, info.from)
			if bool2 {
				break
			}
		}
		break
	}
}
