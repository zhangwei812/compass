package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/core/rawdb"
	"gopkg.in/urfave/cli.v1"
	"math/big"
)

func save(ctx *cli.Context) error {
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
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println()
			fmt.Println("=================DO SAVE========================current epoch :", currentEpoch)
			d.queryCommpassInfo(ChaintypeHeight)
			d.queryCommpassInfo(QueryRelayerinfo)
			fmt.Println("doSave....")
			d.doSave(d.getEthHeaders())
			fmt.Println("doSave over")
			d.queryCommpassInfo(ChaintypeHeight)
			d.atlasBackendCh <- NextStep
		}
	}
}

func (d *commpassInfo) doSave(chains []types.Header) {
	l := len(chains)
	if l == 0 {
		fmt.Println("ignore  header len :", len(chains))
		return
	}
	marshal, _ := rlp.EncodeToBytes(chains)
	conn := d.client
	for k, _ := range d.relayerData {
		person[0].Count += int64(l)
		d.relayerData[k].realSave(conn, ChainTypeETH, marshal)
	}
	saveConfig("person_info_save.json")
}
func (r *relayerInfo) realSave(conn *ethclient.Client, chainType rawdb.ChainType, marshal []byte) bool {
	input := packInput(abiHeaderStore, "save", big.NewInt(int64(chainType)), big.NewInt(int64(ChainTypeMAP)), marshal)
	b := sendContractTransaction(conn, r.from, HeaderStoreAddress, nil, r.priKey, input)
	return b
}
