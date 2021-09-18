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
	RouterContractAddress_map string // map合约
	MaxCount                  int    // 一届同步多少次
	RegisterValue             uint64 // 质押多少钱
	StartVerityNum            uint64 // 开始验证区块
	LimitOnce                 uint64 // 一次同步多少个
	AtlasUrl                  string
	EthUrl                    string
	Password                  string
}

var config cfg

func initConfig() {
	data, err := ioutil.ReadFile(fmt.Sprintf("compass_config.json"))
	if err != nil {
		log.Crit("compass config readFile Err", err.Error())
	}
	_ = json.Unmarshal(data, &config)
}
