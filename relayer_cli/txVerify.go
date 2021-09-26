package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/urfave/cli.v1"
	"log"
	"math/big"
	"time"
)

var (
	RouterContractAddress     = config.RouterContractAddress
	RouterContractAddress_map = config.RouterContractAddress_map
	EventSwapOutHash          = crypto.Keccak256Hash([]byte("LogSwapOut(uint256,address,address,address,uint256,uint256,uint256)"))
	currentVerityNum          = config.StartVerityNum // 开始验证区块
)

func txverify(ctx *cli.Context) error {
	commpassInfo := commpassInfo{}
	commpassInfo.relayerData = []*relayerInfo{
		{url: keystore1},
	}
	commpassInfo.preWork(ctx)
	// 验证
	commpassInfo.doTxVerity()
	return nil
}
func init() {
	RouterContractAddress = config.RouterContractAddress
	RouterContractAddress_map = config.RouterContractAddress_map
	currentVerityNum = config.StartVerityNum // 开始验证区块
}
func (d *commpassInfo) doTxVerity() {
	for {
		//d.doTxVerity1(currentVerityNum, 11104090)
		//------验证,开始块 ------
		num, _ := getCurrentNumberAbi(d.client, ChainTypeETH, d.relayerData[0].from)
		if num > currentVerityNum {
			d.doTxVerity1(currentVerityNum, num)
			currentVerityNum = num
			person[0].Txverity = int64(num)
			saveConfig("person_info_txverify.json")
		}
		time.Sleep(time.Second)
	}
}

func (d *commpassInfo) doTxVerity1(fromBlock uint64, toBlock uint64) {
	fmt.Println()
	fmt.Println("=================DO TxVerity========================")
	EthConn, _ := dialEthConn()
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		ToBlock:   big.NewInt(int64(toBlock)),
		Addresses: []common.Address{common.HexToAddress(RouterContractAddress)},
	}
	logs, err := EthConn.FilterLogs(context.Background(), query)
	if err != nil {
		panic(err)
	}
	for _, aLog := range logs {
		if EventSwapOutHash != aLog.Topics[0] {
			continue
		}
		fmt.Println("-----aLog index:", aLog.Index)
		d.HandleLogSwapOut(&aLog, EthConn)
		fmt.Println()
	}
}

func (d *commpassInfo) HandleLogSwapOut(aLog *types.Log, ethConn *ethclient.Client) {
	err := abiRouter.UnpackIntoInterface(&eventResponse, "LogSwapOut", aLog.Data)
	if err != nil {
		panic(err)
	}
	txProve := GetTxProve(*ethConn, aLog, &eventResponse)

	token := common.BytesToAddress(aLog.Topics[1].Bytes())

	conn := d.client

	//input, _ := abiTxVerity.Pack("txVerify",
	//	aLog.Address,
	//	token,
	//	eventResponse.FromChainID,
	//	eventResponse.ToChainID,
	//	txProve)
	//if err != nil {
	//	log.Fatal(abiRouter, " error ", err)
	//}
	//relayer := d.relayerData[0]
	////RouterContractAddress_map1:=common.HexToAddress(RouterContractAddress_map)
	////fmt.Println("RouterContractAddress_map1",RouterContractAddress_map1)
	//b := sendContractTransaction(conn, relayer.from, TxVerifyAddress, nil, relayer.priKey, input)

	//function swapIn(uint256 id, address token, address to, uint amount, uint fromChainID, address sourceRouter, bytes memory data) external onlyMPC {

	//swapverify.txVerify(sourceRouter,token,fromChainID,chainID,data);
	to := common.BytesToAddress(aLog.Topics[3].Bytes())
	//to := common.HexToAddress("0xb324c41ef2b839c7918553ecc0230cc279660299")
	input := packInput(abiRouter, "swapIn",
		eventResponse.OrderId,
		token,
		to,
		eventResponse.Amount,
		eventResponse.FromChainID,
		aLog.Address,
		txProve)
	//fmt.Println(eventResponse.OrderId,
	//	token,
	//	to,
	//	eventResponse.Amount,
	//	eventResponse.FromChainID,
	//	aLog.Address)
	//fmt.Println("txProve   ", common.Bytes2Hex(txProve))
	if err != nil {
		log.Fatal(abiRouter, " error ", err)
	}
	relayer := d.relayerData[0]
	RouterContractAddress_map1 := common.HexToAddress(RouterContractAddress_map)
	fmt.Println("RouterContractAddress_map", RouterContractAddress_map1)
	fmt.Println("target Address:", to.String(), "  balance:", getBalance(conn, to))
	//fmt.Println("relayer.from    ",relayer.from)
	b := sendContractTransaction(conn, relayer.from, RouterContractAddress_map1, nil, relayer.priKey, input)
	fmt.Println("TxVerify result:", b, "   eth blockNumber ", aLog.BlockNumber, "  transactionIndex: ", aLog.TxIndex)
	fmt.Println("target Address:", to.String(), "  balance:", getBalance(conn, to))
}
