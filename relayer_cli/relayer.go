package main

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"time"
)

var (
	registerValue          = config.RegisterValue
	Base                   = new(big.Int).SetUint64(10000)
	contractQueryFailedErr = errors.New("Contract query failed result ")
)

func init3() {
	registerValue = config.RegisterValue
}

const (
	RegisterAmount = 100000
)

func register(ctx *cli.Context, conn *ethclient.Client, info relayerInfo) {
	value := ethToWei(info.registerValue)
	if info.registerValue < RegisterAmount {
		Fatal("register", "Amount must bigger than ", RegisterAmount)
	}
	fee := ctx.GlobalUint64(FeeFlag.Name)
	checkFee(new(big.Int).SetUint64(fee))
	input := packInput(abiRelayer, "register", value)
	sendContractTransaction(conn, info.from, RelayerAddress, nil, info.priKey, input)
}

func checkFee(fee *big.Int) {
	if fee.Sign() < 0 || fee.Cmp(Base) > 0 {
		Fatal("Please set correct fee value")
	}
}

func (d *compassInfo) relayerRegister() {
	fmt.Println()
	log.Info("--------------------do relayer Register-----------------------------------")
	conn := d.client
	ctx := d.ctx
	for k := range d.relayerData {
		register(ctx, d.client, *d.relayerData[k])
		for {
			bool1, bool2, relayerEpoch, _ := queryRegisterInfo(conn, d.relayerData[k].from)
			log.Info("relayerInfo", "ADDRESS", d.relayerData[k].from, "ISREGISTER", bool1, " ISRELAYER", bool2, " RELAYER_EPOCH", relayerEpoch)
			if bool2 {
				break
			}
			log.Info("waiting to become relayer...................................")
			for {
				time.Sleep(time.Second * 10)
				bool1, bool2, relayerEpoch, _ := queryRegisterInfo(conn, d.relayerData[k].from)
				if bool2 {
					log.Info("relayerInfo", "ADDRESS", d.relayerData[k].from, "ISREGISTER", bool1, " ISRELAYER ", bool2, " RELAYER_EPOCH", relayerEpoch)
					break
				}
			}
			break
		}
	}
}
func (d *compassInfo) waitBecomeRelayer(info relayerInfo) {
	conn := d.client
	register(d.ctx, d.client, info)
	for {
		bool1, bool2, relayerEpoch, _ := queryRegisterInfo(conn, info.from)
		log.Info("relayerInfo", "ADDRESS", info.from, "ISREGISTER", bool1, " ISRELAYER ", bool2, " RELAYER_EPOCH", relayerEpoch)
		if bool2 {
			break
		}
		log.Info("waiting to become relayer...................................")
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
