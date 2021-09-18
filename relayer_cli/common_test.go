package main

import (
	"github.com/ethereum/go-ethereum/common"
	"log"
	"math/big"
	"testing"
	"time"
)

func Test_CommpassInfo_preWork(t *testing.T) {
	commpassInfo := commpassInfo{}
	commpassInfo.relayerData = []*relayerInfo{
		{url: keystore1},
	}
	conn := getAtlasConn()
	commpassInfo.client = conn
	p, from := loadprivate(commpassInfo.relayerData[0].url)
	var acc common.Address
	acc.SetBytes(from.Bytes())
	a := getBalance(conn, acc)
	b, _ := a.Int64()
	commpassInfo.relayerData[0].from = acc
	commpassInfo.relayerData[0].registerValue = b
	commpassInfo.relayerData[0].priKey = p
	value := ethToWei(RegisterAmount)
	if commpassInfo.relayerData[0].registerValue < RegisterAmount {
		log.Fatal("Amount must bigger than ", RegisterAmount)
	}
	checkFee(new(big.Int).SetUint64(0))
	input := packInput(abiRelayer, "register", value)
	sendContractTransaction(conn, commpassInfo.relayerData[0].from, RelayerAddress, nil, commpassInfo.relayerData[0].priKey, input)
	for {
		time.Sleep(time.Second)
		commpassInfo.queryCommpassInfo(QueryRelayerinfo)
	}
}
