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
	"github.com/ethereum/go-ethereum/log"
	"github.com/mapprotocol/atlas/chains"
	"github.com/mapprotocol/atlas/core/vm"
	"github.com/mapprotocol/atlas/params"
	params2 "github.com/mapprotocol/atlas/params"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
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
	abiERC20, _        = abi.JSON(strings.NewReader(MapERC20ContractAbi))
	abiTxVerity, _     = abi.JSON(strings.NewReader(params2.TxVerifyABIJSON))
	RelayerAddress     = params2.RelayerAddress
	HeaderStoreAddress = params2.HeaderStoreAddress
	TxVerifyAddress    = params2.TxVerifyAddress
	LimitOnce          = config.LimitOnce // 一次最多同步多少个
	AtlasUrl           = config.AtlasUrl
	EthUrl             = config.EthUrl
	client             *ethclient.Client
)

func init2() {
	keystore1 = config.Keystore
	password = config.Password
	LimitOnce = config.LimitOnce // 一次最多同步多少个
	AtlasUrl = config.AtlasUrl
	EthUrl = config.EthUrl
}

const (
	ChainTypeETH = chains.ChainTypeETHTest
	ChainTypeMAP = chains.ChainTypeMAP
	// method name
	CurNbrAndHash = vm.CurNbrAndHash

	ChainTypeMAPTemp = 177
)

type compassInfo struct {
	atlasBackendCh chan string
	notifyCh       chan uint64 // notify can do save
	currentEpoch   uint64
	ethData        []types.Header
	client         *ethclient.Client
	relayerData    []*relayerInfo
	ctx            *cli.Context
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

func (d *compassInfo) preWork(ctx *cli.Context) {
	conn := getAtlasConn()
	d.atlasBackendCh = make(chan string)
	d.notifyCh = make(chan uint64)
	d.client = conn
	d.currentEpoch = 0
	d.ctx = ctx
	for k := range d.relayerData {
		Ele := d.relayerData[k]
		priKey, from := loadPrivate(Ele.url)
		var acc common.Address
		acc.SetBytes(from.Bytes())
		Ele.registerValue = int64(registerValue)
		Ele.from = acc
		Ele.priKey = priKey
		Ele.fee = uint64(0)
		bb := getBalance(conn, Ele.from)
		Ele.preBalance = bb
		Ele.nowBalance = bb
		//---- init person
		person[k].Address = acc.String()
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

func loadPrivate(keyfile string) (*ecdsa.PrivateKey, common.Address) {
	keyjson, err := ioutil.ReadFile(keyfile)
	if err != nil {
		Fatal("loadPrivate ReadFile", fmt.Errorf("failed to read the keyfile at '%s': %v", keyfile, err))
	}
	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		Fatal("loadPrivate DecryptKey", fmt.Errorf("error decrypting key: %v", err))
	}
	priKey1 := key.PrivateKey
	return priKey1, crypto.PubkeyToAddress(priKey1.PublicKey)
}
func sendContractTransaction(client *ethclient.Client, from, toAddress common.Address, value *big.Int, privateKey *ecdsa.PrivateKey, input []byte) bool {
	// Ensure a valid value field and resolve the account nonce
	nonce, err := client.PendingNonceAt(context.Background(), from)
	if err != nil {
		log.Error("sendContractTransaction PendingNonceAt", "err", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Error("sendContractTransaction SuggestGasPrice", "err", err)
	}

	gasLimit := uint64(2100000) // in units
	// If the contract surely has code (or code is not needed), estimate the transaction
	// 如果合同确实有代码（或不需要代码），则估计交易
	msg := ethchain.CallMsg{From: from, To: &toAddress, GasPrice: gasPrice, Value: value, Data: input}
	gasLimit, err = client.EstimateGas(context.Background(), msg)
	if err != nil {
		log.Info("client.EstimateGas Contract exec ", "failed", err)
	}
	//log.Info("EstimateGas gasLimit : ", gasLimit)
	if gasLimit < 1 {
		gasLimit = 866328
	}

	// Create the transaction, sign it and schedule it for execution
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, input)

	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Error("sendContractTransaction ChainID", "client.ChainID err ", err)
	}
	//log.Info("TX data nonce ", nonce, " transfer value ", value, " gasLimit ", gasLimit, " gasPrice ", gasPrice, " chainID ", chainID)
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		log.Error("sendContractTransaction signedTx", "types.SignTx err", err)
	}

	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Error("sendContractTransaction ", "client.SendTransaction err", err)
	}
	txHash := signedTx.Hash()
	count := 0
	for {
		time.Sleep(time.Millisecond * 200)
		_, isPending, err := client.TransactionByHash(context.Background(), txHash)

		if err != nil {
			log.Error("sendContractTransaction TransactionByHash", "err", err)
		}
		count++
		if !isPending {
			break
		}
	}
	receipt, err := client.TransactionReceipt(context.Background(), txHash)
	count1 := 0
	if err != nil {
		log.Error("TransactionReceipt receipt1", "err", err)
		for {
			time.Sleep(time.Second * 5)
			count1++
			receipt, err = client.TransactionReceipt(context.Background(), txHash)
			if err != nil {
				log.Error("TransactionReceipt receipt finding...", "err", err)
			} else {
				break
			}
			if count1 > 10 {
				return false
			}
		}
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		block, err := client.BlockByHash(context.Background(), receipt.BlockHash)
		if err != nil {
			log.Error("Transaction", "err", err, "receipt.BlockHash", receipt.BlockHash)
		}
		log.Info("Transaction Success", "block Number", receipt.BlockNumber.Uint64(), " block txs", len(block.Transactions()), "blockhash", block.Hash().Hex())
		return true
	} else if receipt.Status == types.ReceiptStatusFailed {
		log.Info("TX data  ", "nonce", nonce, " transfer value", value, " gasLimit", gasLimit, " gasPrice", gasPrice, " chainID", chainID)
		log.Info("Transaction Failed", "Block Number", receipt.BlockNumber.Uint64())
		return false
	}
	return false
}

func packInput(abiHeaderStore abi.ABI, abiMethod string, params ...interface{}) []byte {
	input, err := abiHeaderStore.Pack(abiMethod, params...)
	if err != nil {
		Fatal(abiMethod, " error", err)
	}
	return input
}

func reconnection(c *ethclient.Client) {
	conn, err := ethclient.Dial(AtlasUrl)
	for err != nil {
		log.Error("reconnection", "err", err)
		time.Sleep(1 * time.Second)
	}
	c = conn
}
func Fatal(errName string, args ...interface{}) {
	s := args[0].(string)
	if strings.HasPrefix(s, "Post") {
		reconnection(client)
	}
	log.Crit(errName, args)
}
