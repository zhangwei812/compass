package main

import (
	"context"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
)

var (
	maxCount = config.MaxCount
)

func initCfg(ctx *cli.Context) {
	initConfig1(ctx)
	maxCount = config.MaxCount
	init2()
	init3()
	init4()
	init5()
}

func getAtlasConn() *ethclient.Client {
	url := AtlasUrl
	conn, err := ethclient.Dial(url)
	client = conn
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
		for {
			select {
			case <-d.atlasBackendCh:
				count++
				if count >= maxCount {
					canNext = false
				} else if canNext {
					// 次数没到 继续执行
					number, err := conn.BlockNumber(context.Background())
					if err != nil {
						Fatal("get BlockNumber err ", err)
					}
					currentEpoch := number / epochHeight
					go func() { d.notifyCh <- currentEpoch }()
				}
			}
		}

	}()

	for {
		//-------保存,阶段-------
		number, err := conn.BlockNumber(context.Background())
		if err != nil {
			Fatal("get BlockNumber err ", err)
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
