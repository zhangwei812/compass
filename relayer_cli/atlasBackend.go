package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
)

var (
	maxCount = config.MaxCount
)

func init() {
	fmt.Println("atlas 1 ")
	initConfig()
	maxCount = config.MaxCount
}

func getAtlasConn() *ethclient.Client {
	url := AtlasUrl
	conn, err := ethclient.Dial(url)
	if err != nil {
		log.Fatalf("Failed to connect to the atlas chain client: %v", err)
	}
	return conn
}

func (d *commpassInfo) atlasBackend() {
	count := 0
	conn := d.client
	canNext := true
	go func() {
		select {
		case <-d.atlasBackendCh:
			count++
			if count >= maxCount {
				canNext = false
			} else if canNext {
				// 次数没到 继续执行
				number, err := conn.BlockNumber(context.Background())
				if err != nil {
					log.Fatal("get BlockNumber err ", err)
				}
				currentEpoch := number / epochHeight
				d.notifyCh <- currentEpoch
			}
		}
	}()

	for {
		//-------保存,阶段-------
		number, err := conn.BlockNumber(context.Background())
		if err != nil {
			log.Fatal("get BlockNumber err ", err)
		}
		currentEpoch := number/epochHeight + 1
		if currentEpoch != d.currentEpoch {
			canNext = true
			count = 0
			d.currentEpoch = currentEpoch
			d.notifyCh <- currentEpoch
		}
	}
}
