package main

import (
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/mapprotocol/atlas/core/rawdb"
	"math/big"
)

func (d *commpassInfo) saveMock() {
	for {
		select {
		case currentEpoch := <-d.notifyCh:
			fmt.Println("current epoch ========>", currentEpoch)
			d.queryCommpassInfo(ChaintypeHeight)
			d.queryCommpassInfo(QueryRelayerinfo)
			d.doSave(d.getEthHeaders())
			d.queryCommpassInfo(ChaintypeHeight)
			d.atlasBackendCh <- NextStep

		}
	}
}

func (d *commpassInfo) doSave(chains []types.Header) {
	fmt.Println("=================DO SAVE========================")
	if len(chains) == 0 {
		fmt.Println("error ! header len :", len(chains))
		return
	}
	marshal, _ := rlp.EncodeToBytes(chains)
	conn := d.client
	for k, _ := range d.relayerData {
		fmt.Println("ADDRESS:", d.relayerData[k].from)
		d.relayerData[k].realSave(conn, ChainTypeETH, marshal)
	}
}
func (r *relayerInfo) realSave(conn *ethclient.Client, chainType rawdb.ChainType, marshal []byte) bool {
	input := packInput(abiHeaderStore, "save", big.NewInt(int64(chainType)), big.NewInt(int64(ChainTypeMAP)), marshal)
	b := sendContractTransaction(conn, r.from, HeaderStoreAddress, nil, r.priKey, input)
	return b
}
