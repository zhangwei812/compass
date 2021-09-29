package main

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
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
	// 记录
	txSuccessNum = 0
	txFailNum    = 0
	txFailRecord = map[int]struct {
		ethNum   int
		logIndex int
	}{}
	nowEthBlockInAtlas = 0
	nowEthBlock        = 0
	txStartTime        = ""
)

func txverify(ctx *cli.Context) error {
	initCfg(ctx)
	commpassInfo := commpassInfo{}
	commpassInfo.relayerData = []*relayerInfo{
		{url: keystore1},
	}
	commpassInfo.preWork(ctx)
	// 验证
	commpassInfo.doTxVerity()
	return nil
}
func init5() {
	RouterContractAddress = config.RouterContractAddress
	Erc20ContractAddress = config.ERC20ContractAddress
	RouterContractAddressMap = config.RouterContractAddress_map
	Erc20ContractAddressMap = config.ERC20ContractAddress_map
	currentVerityNum = config.StartVerityNum // 开始验证区块起始头
}

func (d *commpassInfo) doTxVerity() {
	tempCount := 0
	txStartTime = time.Now().Format("2006/1/2 15:04:05")
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
	nowEthBlockInAtlas = int(toBlock)
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
		log.Error(err)
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
		fmt.Println("在以太坊块:", aLog.BlockNumber, "交易ID:", aLog.Index)
		d.HandleLogSwapOut(&aLog, EthConn)
		fmt.Println()
	}
}

func (d *commpassInfo) HandleLogSwapOut(aLog *types.Log, ethConn *ethclient.Client) {
	err := abiRouter.UnpackIntoInterface(&eventResponse, "LogSwapOut", aLog.Data)
	if err != nil {
		log.Error(err)
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
	//	Fatal(abiRouter, " error ", err)
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
	//fmt.Println(eventResponse.OrderId,
	//	token,
	//	to,
	//	eventResponse.Amount,
	//	eventResponse.FromChainID,
	//	aLog.Address)
	//fmt.Println(common.Bytes2Hex(txProve))

	if err != nil {
		Fatal(abiRouter, " error ", err)
	}
	relayer := d.relayerData[0]
	RouterContractAddressMap1 := common.HexToAddress(RouterContractAddressMap)
	balance1 := getTargetAddressBalance(conn, relayer.from, to)
	fmt.Println("target mint1 Address:", to.String(), "  balance:", balance1, "will mint:", eventResponse.Amount)
	result := sendContractTransaction(conn, relayer.from, RouterContractAddressMap1, nil, relayer.priKey, input)
	fmt.Println("TxVerify result:", result, "   eth blockNumber ", aLog.BlockNumber, "  transactionIndex: ", aLog.TxIndex)
	txRecord(result, aLog, d)
	balance2 := getTargetAddressBalance(conn, relayer.from, to)
	c := balance2 - balance1
	fmt.Println("target mint2 Address:", to.String(), "  balance:", balance2, "change money:", balance2-balance1)
	if big.NewInt(int64(c)).Cmp(eventResponse.Amount) != 0 {
		fmt.Println("err: abnormal mint---> Address:", to.String(), "  balance:", balance2, "change money:", balance2-balance1)
	}
}

func txRecord(result bool, aLog *types.Log, d *commpassInfo) {
	if result {
		txSuccessNum++
	} else {
		txFailNum++
		txFailRecord[len(txFailRecord)] = struct {
			ethNum   int
			logIndex int
		}{ethNum: int(aLog.BlockNumber), logIndex: int(aLog.TxIndex)}
	}
	l := len(txFailRecord)
	nowNum, _ := d.client.BlockNumber(context.Background())
	fmt.Println("验证了", txSuccessNum+txFailNum, " 成功了:", txSuccessNum, " 失败了:", txFailNum, "验证到了", currentVerityNum, "当前atlas上Eth高度是:", nowEthBlockInAtlas, "atals的块高", nowNum)
	for l > 0 {
		l--
		fmt.Println("eth上第", txFailRecord[l].ethNum, "个块", "交易index是:", txFailRecord[l].logIndex)
	}
}
