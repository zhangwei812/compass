package main

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/log"
	"io/ioutil"
)

type cfg struct {
	Keystore                  string //keystore 路径
	RouterContractAddress     string //以太坊合约
	ERC20ContractAddress      string // 以太坊 ERC20/ token 地址
	RouterContractAddress_map string // map合约
	ERC20ContractAddress_map  string // map ERC20/ token 地址
	MaxCount                  int    // 一届同步多少次
	RegisterValue             uint64 // 质押多少钱
	StartVerityNum            uint64 // 开始验证区块
	LimitOnce                 uint64 // 一次同步多少个
	AtlasUrl                  string
	EthUrl                    string
	Password                  string
	DefaultTransactionAddress string // 默认往这个账号发钱 swapOut
	DefaultAmount             uint64 // 默认发的钱
}

var config cfg

func initConfig1() {
	fmt.Println("init config...")
	data, err := ioutil.ReadFile(fmt.Sprintf("compass_config.json"))
	if err != nil {
		log.Crit("compass config readFile Err", err.Error())
	}
	_ = json.Unmarshal(data, &config)
	initConfig2()
}

type PersonInfo struct {
	Address  string
	Count    int64
	Txverity int64
}

var person []PersonInfo

func initConfig2() {
	data, err := ioutil.ReadFile(fmt.Sprintf("person_info_txverify.json"))
	if err != nil {
		log.Crit("compass personInfo config readFile Err", err.Error())
	}
	_ = json.Unmarshal(data, &person)
	if person == nil || len(person) == 0 {
		for i := 0; i < 10; i++ {
			person = append(person, PersonInfo{})
		}
	}
	//n := uint64(person[0].Txverity)
	////if config.StartVerityNum < n {
	////	config.StartVerityNum = n
	////}
	fmt.Println("init over")
}

func saveConfig(file string) {
	// 创建文件
	data, err := json.Marshal(person)
	if err != nil {
		fmt.Println("saveConfig file failed", err.Error(), "    ", file)
		return
	}
	if err := ioutil.WriteFile(file, data, 1000); err != nil {
		fmt.Println("saveConfig file failed", err.Error(), "    ", file)
	}
}
