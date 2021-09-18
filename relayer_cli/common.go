package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	ethchain "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
	params2 "github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"math/big"
	"strings"
	"time"
)

var (
	epochHeight = params.NewEpochLength
	//keystore1   = "D:/work/atlas002/atlas/data/keystore/UTC--2021-07-19T02-04-57.993791200Z--df945e6ffd840ed5787d367708307bd1fa3d40f4"
	//keystore1          = "D:/work/atlas_master/atlas/data1/keystore/UTC--2021-09-17T08-33-00.520132000Z--6c6df7b8309e94aecce535d5e879e97708fea16b"
	//keystore1          = "D:/work/atlas002/atlas/data1/keystore/UTC--2021-09-18T02-37-29.614751180Z--5fc316bc118026f3839ddd737cae6838f9dc992b"
	keystore1          = config.Keystore
	password           = config.Password
	abiRelayer, _      = abi.JSON(strings.NewReader(params2.RelayerABIJSON))
	abiHeaderStore, _  = abi.JSON(strings.NewReader(params2.HeaderStoreABIJSON))
	abiRouter, _       = abi.JSON(strings.NewReader(RouterContractAbi))
	abiTxVerity, _     = abi.JSON(strings.NewReader(params2.TxVerifyABIJSON))
	RelayerAddress     = params2.RelayerAddress
	HeaderStoreAddress = params2.HeaderStoreAddress
	TxVerifyAddress    = params2.TxVerifyAddress
	LimitOnce          = config.LimitOnce // 一次最多同步多少个
	AtlasUrl           = config.AtlasUrl
	EthUrl             = config.EthUrl
)

func init() {
	keystore1 = config.Keystore
	password = config.Password
	LimitOnce = config.LimitOnce // 一次最多同步多少个
	AtlasUrl = config.AtlasUrl
	EthUrl = config.EthUrl
}

const (
	ChainTypeETH = 3
	ChainTypeMAP = 211
	// method name
	CurNbrAndHash = vm.CurNbrAndHash
)

type commpassInfo struct {
	atlasBackendCh chan string
	notifyCh       chan uint64 // notify can do save
	currentEpoch   uint64
	ethData        []types.Header
	client         *ethclient.Client
	relayerData    []*relayerInfo
}
type relayerInfo struct {
	url           string
	from          common.Address
	preBalance    *big.Float
	nowBalance    *big.Float
	registerValue int64
	priKey        *ecdsa.PrivateKey
	fee           uint64
}

func (d *commpassInfo) preWork(ctx *cli.Context) {
	conn := getAtlasConn()
	d.atlasBackendCh = make(chan string)
	d.notifyCh = make(chan uint64)
	d.client = conn
	d.currentEpoch = 0
	for k, _ := range d.relayerData {
		Ele := d.relayerData[k]
		priKey, from := loadprivate(Ele.url)
		var acc common.Address
		acc.SetBytes(from.Bytes())
		Ele.registerValue = int64(registerValue)
		Ele.from = acc
		Ele.priKey = priKey
		Ele.fee = uint64(0)
		bb := getBalance(conn, Ele.from)
		Ele.preBalance = bb
		Ele.nowBalance = bb
		register(ctx, d.client, *d.relayerData[k])
		for {
			bool1, bool2, relayerEpoch, _ := queryRegisterInfo(conn, d.relayerData[k].from)
			fmt.Println("ADDRESS:", d.relayerData[k].from, "ISREGISTER:", bool1, " ISRELAYER :", bool2, " RELAYER_EPOCH :", relayerEpoch)
			if bool2 {
				break
			}
			fmt.Println("waiting to become relayer...................................")
			for {
				time.Sleep(time.Second * 10)
				bool1, bool2, relayerEpoch, _ := queryRegisterInfo(conn, d.relayerData[k].from)
				if bool2 {
					fmt.Println("ADDRESS:", d.relayerData[k].from, "ISREGISTER:", bool1, " ISRELAYER :", bool2, " RELAYER_EPOCH :", relayerEpoch)
				}
			}
			break
		}
	}
}

func ethToWei(registerValue int64) *big.Int {
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	value := new(big.Int).Mul(big.NewInt(registerValue), baseUnit)
	return value
}

func weiToEth(value *big.Int) uint64 {
	baseUnit := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	valueT := new(big.Int).Div(value, baseUnit).Uint64()
	return valueT
}

func loadprivate(keyfile string) (*ecdsa.PrivateKey, common.Address) {
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		log.Fatal(fmt.Errorf("failed to read the keyfile at '%s': %v", keyfile, err))
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		log.Fatal(fmt.Errorf("error decrypting key: %v", err))
	}
	priKey1 := key.PrivateKey
	return priKey1, crypto.PubkeyToAddress(priKey1.PublicKey)
}
func sendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte) bool {
	// Ensure a valid value field and resolve the account nonce
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		panic(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		panic(err)
	}

	gasLimit := uint64(2100000) // in units
	// If the contract surely has code (or code is not needed), estimate the transaction
	msg := ethchain.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		fmt.Println("Contract exec failed", err)
	}
	if gasLimit < 1 {
		gasLimit = 866328
	}

	// Create the transaction, sign it and schedule it for execution
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, input)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		panic(err)
	}
	//fmt.Println("TX data nonce ", nonce, " transfer value ", value, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		panic(err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		panic(err)
	}
	txHash := signedTx.Hash()
	count := 0
	for {
		time.Sleep(time.Millisecond * 200)
		_, isPending, err := client.TransactionByHash(context.Background(), txHash)
		if err != nil {
			panic(err)
		}
		count++
		if !isPending {
			break
		}
	}
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	if err != nil {
		panic(err)
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		block, err := client.BlockByHash(context.Background(), receipt.BlockHash)
		if err != nil {
			panic(err)
		}
		fmt.Println("Transaction Success", " block Number", receipt.BlockNumber.Uint64(), " block txs", len(block.Transactions()), "blockhash", block.Hash().Hex())
		return true
	} else if receipt.Status == types.ReceiptStatusFailed {
		fmt.Println("Transaction Failed ", " Block Number", receipt.BlockNumber.Uint64())
		return false
	}
	return false
}

func packInput(abiHeaderStore abi.ABI, abiMethod string, params ...interface{}) []byte {
	input, err := abiHeaderStore.Pack(abiMethod, params...)
	if err != nil {
		log.Fatal(abiMethod, " error ", err)
	}
	return input
}
