package matic_data

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"log"
	"math/big"
	"signmap/libs"
	"signmap/libs/contracts"
	"strings"
)

func GetLastSign() *big.Int {
	client := libs.GetClient()

	fromAddress := BindAddress()
	var abiStaking, _ = abi.JSON(strings.NewReader(curAbi))
	input := contracts.PackInput(abiStaking, "getLastSign", fromAddress)
	ret := contracts.CallContract(client, fromAddress, libs.DataContractAddress, input)
	var res = big.NewInt(0)
	if len(ret) == 0 {
		log.Println("getLastSign error.")
		return res
	}
	err := abiStaking.UnpackIntoInterface(&res, "getLastSign", ret)

	if err != nil {
		log.Println("abi error", err)
		return res
	}
	return res
}
