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
func (d *compassInfo) getEthHeaders() []types.Header {
	Ethconn, _ := dialEthConn()
	startNum, _ := getCurrentNumberAbi(d.client, ChainTypeETH, d.relayerData[0].from)
	nowEthBlockInAtlas = int(startNum)
	a, _ := Ethconn.BlockNumber(context.Background())
	nowEthBlock = int(a)
	Headers := make([]types.Header, 0)
	// 让eth处理完分叉再进行同步数据 只同步比eth小13的块
	if a <= startNum+12 {
		return Headers
	}

	var i uint64
	LimitOnce = min(config.LimitOnce, a-startNum-12)
	//fmt.Println("LimitOnce:", LimitOnce, "a", a, "startNum", startNum) //LimitOnce: 1 a 11186031 startNum 11186017
	for i = 1; i <= LimitOnce; i++ {
		Header, err := Ethconn.HeaderByNumber(context.Background(), big.NewInt(int64(startNum+i)))
		if err != nil {
			return Headers
		}
		Headers = append(Headers, *Header)
	}
	return Headers
}

func min(x, y uint64) uint64 {
	if x > y {
		return y
	}
	return x
}
