package main

import (
	"context"
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
	commpassInfo := compassInfo{}
	commpassInfo.relayerData = []*relayerInfo{
		{url: keystore1},
	}
	commpassInfo.preWork(ctx)
	// 同步数据
	commpassInfo.relayerRegister()
	go commpassInfo.atlasBackend()
	go commpassInfo.saveMock()
	select {}
}

func (d *compassInfo) saveMock() {
	saveStartTime = time.Now().Format("2006/1/2 15:04:05")
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			log.Info("")
			headers := d.getEthHeaders()
			l := len(headers)
			if l > 0 {
				log.Info("=================DO SAVE========================", "current epoch ", currentEpoch)
				d.queryCommpassInfo(ChaintypeHeight)
				d.queryCommpassInfo(QueryRelayerinfo)
				log.Info("doSave....")
				d.doSave(headers)
				log.Info("doSave over")
				d.queryCommpassInfo(ChaintypeHeight)
			} else {
				d.queryCommpassInfo(ChaintypeHeight)
				log.Info("waiting  new transactions....")
				time.Sleep(time.Second * 5)
			}
			go func() { d.atlasBackendCh <- NextStep }()
		}
	}
}

func (d *compassInfo) doSave(chains []types.Header) {
	l := len(chains)
	if l == 0 {
		log.Info("ignore header", "len", len(chains))
		return
	}
	marshal, _ := rlp.EncodeToBytes(chains)
	log.Info("chains bytes size", "bytesSize", len(marshal), "chains length", l)
	conn := d.client
	for k := range d.relayerData {
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

func saveRecord(result bool, start, end int, d *compassInfo) {
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
	log.Info("Save record1", "save chains length", saveCount)
	log.Info("Save record2", "start save time", saveStartTime, " success", saveSuccessNum, "fail", saveFailNum)
	log.Info("Save record3", "atlas's eth block number", nowEthBlockInAtlas, "atlas block number", nowNum, "eth block number", nowEthBlock)
	l := len(saveFailRecord)
	for l > 0 {
		l--
		log.Info("save fail records", "eth start", saveFailRecord[l].start, "end", saveFailRecord[l].end, " time", saveFailRecord[l].time)
	}
}
