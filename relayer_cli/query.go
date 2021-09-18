package main

import (
	"context"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mapprotocol/atlas/core/rawdb"
	"log"
	"math"
	"math/big"
)

const (
	BALANCE          = "balance"
	RegisterBalance  = "registerBalance"
	QueryRelayerinfo = "relayerInfo"
	REWARD           = "reward"
	ChaintypeHeight  = "chainTypeHeight"
	NextStep         = "next step"
)

func (d *commpassInfo) queryCommpassInfo(ss string) {
	conn := d.client
	switch ss {
	case BALANCE:
		for k, _ := range d.relayerData {
			fmt.Println("ADDRESS:", d.relayerData[k].from, " OLD BALANCE :", d.relayerData[k].preBalance, " NOW BALANCE :", getBalance(conn, d.relayerData[k].from))
		}
	case RegisterBalance:
		for k, _ := range d.relayerData {
			registered, unregistering, unregistered := getRegisterBalance(conn, d.relayerData[k].from)
			fmt.Println("ADDRESS:", d.relayerData[k].from,
				" NOW registerValue BALANCE :", registered, " registerING BALANCE :", unregistering, "registerED BALANCE :", unregistered)
		}
	case QueryRelayerinfo:
		for k, _ := range d.relayerData {
			bool1, bool2, relayerEpoch, _ := queryRegisterInfo(conn, d.relayerData[k].from)
			fmt.Println("ADDRESS:", d.relayerData[k].from, "ISREGISTER:", bool1, " ISRELAYER :", bool2, " RELAYER_EPOCH :", relayerEpoch)
		}
	case REWARD:

	case ChaintypeHeight:
		for k, _ := range d.relayerData {
			currentTypeHeight, hash := getCurrentNumberAbi(conn, ChainTypeETH, d.relayerData[k].from)
			fmt.Println("ADDRESS:", d.relayerData[k].from, " TYPE HEIGHT:", currentTypeHeight, "  HASH:   ", hash)
		}
	}

}

func getBalance(conn *ethclient.Client, address common.Address) *big.Float {
	balance, err := conn.BalanceAt(context.Background(), address, nil)
	if err != nil {
		panic(err)
	}
	balance2 := new(big.Float)
	balance2.SetString(balance.String())
	Value := new(big.Float).Quo(balance2, big.NewFloat(math.Pow10(18)))
	return Value
}

func queryRegisterInfo(conn *ethclient.Client, from common.Address) (bool, bool, *big.Int, error) {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	var input []byte
	input = packInput(abiRelayer, "getRelayer", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error", err)
	}

	method, _ := abiRelayer.Methods["getRelayer"]
	ret, err := method.Outputs.Unpack(output)
	if len(ret) != 0 {
		args := struct {
			register bool
			relayer  bool
			epoch    *big.Int
		}{
			ret[0].(bool),
			ret[1].(bool),
			ret[2].(*big.Int),
		}
		return args.register, args.relayer, args.epoch, nil
	} else {
		fmt.Println("Contract query failed result len == 0")
		return false, false, nil, contractQueryFailedErr
	}
}

//  getCurrent type chain number by abi
func getCurrentNumberAbi(conn *ethclient.Client, chainType rawdb.ChainType, from common.Address) (uint64, string) {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	input := packInput(abiHeaderStore, CurNbrAndHash, big.NewInt(int64(chainType)))
	msg := ethchain.CallMsg{From: from, To: &HeaderStoreAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error", err)
	}
	method, _ := abiHeaderStore.Methods[CurNbrAndHash]
	ret, err := method.Outputs.Unpack(output)
	ret1 := ret[0].(*big.Int).Uint64()
	ret2 := common.BytesToHash(ret[1].([]byte))
	return ret1, ret2.String()

}
func getRegisterBalance(conn *ethclient.Client, from common.Address) (uint64, uint64, uint64) {
	header, err := conn.HeaderByNumber(context.Background(), nil)
	if err != nil {
		panic(err)
	}
	input := packInput(abiRelayer, "getRelayerBalance", from)
	msg := ethchain.CallMsg{From: from, To: &RelayerAddress, Data: input}
	output, err := conn.CallContract(context.Background(), msg, header.Number)
	if err != nil {
		log.Fatal("method CallContract error", err)
	}
	method, _ := abiRelayer.Methods["getRelayerBalance"]
	ret, err := method.Outputs.Unpack(output)
	if len(ret) != 0 {
		args := struct {
			registered    *big.Int
			unregistering *big.Int
			unregistered  *big.Int
		}{
			ret[0].(*big.Int),
			ret[1].(*big.Int),
			ret[2].(*big.Int),
		}
		return weiToEth(args.registered), weiToEth(args.unregistering), weiToEth(args.unregistered)
	}
	log.Fatal("Contract query failed result len == 0")
	return 0, 0, 0
}
