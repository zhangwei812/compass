package main

import (
	"context"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

func dialEthConn() (*ethclient.Client, string) {
	url := EthUrl
	conn, err := ethclient.Dial(EthUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the AtlasChain client: %v", err)
	}
	return conn, url
}
func (d *commpassInfo) getEthHeaders() []types.Header {
	Ethconn, _ := dialEthConn()
	startNum, _ := getCurrentNumberAbi(d.client, ChainTypeETH, d.relayerData[0].from)
	Headers := make([]types.Header, 0)
	var i uint64
	for i = 1; i <= LimitOnce; i++ {
		Header, err := Ethconn.HeaderByNumber(context.Background(), big.NewInt(int64(startNum+i)))
		if err != nil {
			panic(err)
			return Headers
		}
		Headers = append(Headers, *Header)
	}
	return Headers
}
