package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/core/rawdb"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"time"
)

var (
	saveCount      = 0
	saveSuccessNum = 0
	saveFailNum    = 0
	saveFailRecord = map[int]struct {
		start int
		end   int
		time  string
	}{}
	saveStartTime = ""
)

func save(ctx *cli.Context) error {
	initCfg(ctx)
	commpassInfo := commpassInfo{}
	commpassInfo.relayerData = []*relayerInfo{
		{url: keystore1},
	}
	commpassInfo.preWork(ctx)
	// 同步数据
	commpassInfo.relayerRegister()
	go commpassInfo.atlasBackend()
	go commpassInfo.saveMock()
	select {}
	return nil
}

func (d *commpassInfo) saveMock() {
	saveStartTime = time.Now().Format("2006/1/2 15:04:05")
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println()
			log.Info("=================DO SAVE========================", "current epoch ", currentEpoch)
			d.queryCommpassInfo(ChaintypeHeight)
			d.queryCommpassInfo(QueryRelayerinfo)
			log.Info("doSave....")
			d.doSave(d.getEthHeaders())
			log.Info("doSave over")
			d.queryCommpassInfo(ChaintypeHeight)
			go func() { d.atlasBackendCh <- NextStep }()
		}
	}
}

func (d *commpassInfo) doSave(chains []types.Header) {
	l := len(chains)
	if l == 0 {
		log.Info("ignore  header", " len ", len(chains))
		return
	}
	marshal, _ := rlp.EncodeToBytes(chains)
	log.Info("chains bytes size", " len ", len(marshal), "chains length", l)
	conn := d.client
	for k, _ := range d.relayerData {
		person[0].Count += int64(l)
		b := d.relayerData[k].realSave(conn, ChainTypeETH, marshal)
		saveRecord(b, int(chains[0].Number.Uint64()), int(chains[l-1].Number.Uint64()), d)
	}
	saveConfig("person_info_save.json")
}
func (r *relayerInfo) realSave(conn *ethclient.Client, chainType rawdb.ChainType, marshal []byte) bool {
	input := packInput(abiHeaderStore, "save", big.NewInt(int64(chainType)), big.NewInt(int64(ChainTypeMAP)), marshal)
	b := sendContractTransaction(conn, r.from, HeaderStoreAddress, nil, r.priKey, input)

	return b
}

func saveRecord(result bool, start, end int, d *commpassInfo) {
	if result {
		saveCount = end - start + 1
		saveSuccessNum++
	} else {
		saveFailNum++
		saveFailRecord[len(saveFailRecord)] = struct {
			start int
			end   int
			time  string
		}{start: start, end: end, time: time.Now().String()}
	}
	nowNum, _ := d.client.BlockNumber(context.Background())
	log.Info("数据保存", "开始时间", saveStartTime, " 成功次数", saveSuccessNum, "失败次数", saveFailNum, "本次同步数量", saveCount, "当前atlas上Eth高度", nowEthBlockInAtlas, "atals的块高", nowNum, "Eth当前的块高", nowEthBlock)
	l := len(saveFailRecord)
	for l > 0 {
		l--
		log.Info("数据保存", "eth上第", saveFailRecord[l].start, "到", saveFailRecord[l].end, "保存失败, 时间:", saveFailRecord[l].time)
	}
}
