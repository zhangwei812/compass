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
	RouterContractAddress    = config.RouterContractAddress
	Erc20ContractAddress     = config.ERC20ContractAddress
	RouterContractAddressMap = config.RouterContractAddress_map
	Erc20ContractAddressMap  = config.RouterContractAddress_map
	EventSwapOutHash         = crypto.Keccak256Hash([]byte("LogSwapOut(uint256,address,address,address,uint256,uint256,uint256)"))
	currentVerityNum         = config.StartVerityNum // 开始验证区块
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
	Erc20ContractAddress = config.ERC20ContractAddress
	RouterContractAddressMap = config.RouterContractAddress_map
	Erc20ContractAddressMap = config.ERC20ContractAddress_map
	currentVerityNum = config.StartVerityNum // 开始验证区块起始头
}

func (d *commpassInfo) doTxVerity() {
	tempCount := 0
	for {
		//d.doTxVerity1(currentVerityNum, 11104090)
		//------验证,开始块 ------
		num, _ := getCurrentNumberAbi(d.client, ChainTypeETH, d.relayerData[0].from)
		if num > currentVerityNum {
			d.doTxVerity1(currentVerityNum, num)
			currentVerityNum = num
			person[0].Txverity = int64(num)
			saveConfig("person_info_txverify.json")
		} else {
			if tempCount == 0 {
				fmt.Println("waiting new Transation to Verity....... ")
			}
			//每过10分钟输出一次日志
			if tempCount%600 == 0 {
				d.queryCommpassInfo(ChaintypeHeight)
			}

			tempCount++
			if tempCount > 1800 {
				tempCount = 0
			}
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
	if len(logs) > 0 {
		fmt.Println("Discover new transactions!!!    from:", fromBlock, "  to:", toBlock)
	} else {
		fmt.Println("no transactions to verify    from:", fromBlock, "  to:", toBlock)
		fmt.Println("waiting new Transation to Verity....... ")
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

	//token := common.BytesToAddress(aLog.Topics[1].Bytes())
	token := common.HexToAddress(Erc20ContractAddressMap)
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

	to := common.BytesToAddress(aLog.Topics[3].Bytes())
	input := packInput(abiRouter, "swapIn",
		eventResponse.OrderId,
		token,
		to,
		eventResponse.Amount,
		eventResponse.FromChainID,
		aLog.Address,
		txProve)
	if err != nil {
		log.Fatal(abiRouter, " error ", err)
	}
	relayer := d.relayerData[0]
	RouterContractAddressMap1 := common.HexToAddress(RouterContractAddressMap)
	balance1 := getTargetAddressBalance(conn, relayer.from, to)
	fmt.Println("target mint1 Address:", to.String(), "  balance:", balance1, "will mint:", eventResponse.Amount)
	b := sendContractTransaction(conn, relayer.from, RouterContractAddressMap1, nil, relayer.priKey, input)
	fmt.Println("TxVerify result:", b, "   eth blockNumber ", aLog.BlockNumber, "  transactionIndex: ", aLog.TxIndex)
	balance2 := getTargetAddressBalance(conn, relayer.from, to)
	c := balance2 - balance1
	fmt.Println("target mint2 Address:", to.String(), "  balance:", balance2, "change money:", balance2-balance1)
	if big.NewInt(int64(c)).Cmp(eventResponse.Amount) != 0 {
		fmt.Println("err: abnormal mint---> Address:", to.String(), "  balance:", balance2, "change money:", balance2-balance1)
	}
}
