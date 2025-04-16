package main

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// parseRevertReason extracts the revert reason from the error.
func parseRevertReason(err error) string {
	jsonErr, ok := err.(interface{ ErrorData() any })
	if !ok {
		return ""
	}

	datastr, ok := jsonErr.ErrorData().(string)
	if !ok || len(datastr) < 10 {
		return fmt.Sprintf("%v", jsonErr.ErrorData())
	}

	data := common.Hex2Bytes(datastr[2:])

	// "Error(string)" 함수의 selector: 0x08c379a0
	if len(data) >= 4 && fmt.Sprintf("%x", data[:4]) == "08c379a0" {
		// decode string from ABI
		arguments := abi.Arguments{
			{
				Type: mustStringType(),
			},
		}
		unpacked, err := arguments.Unpack(data[4:])
		if err == nil && len(unpacked) > 0 {
			return unpacked[0].(string)
		}
	}

	// fallback
	return datastr
}

func mustStringType() abi.Type {
	t, err := abi.NewType("string", "", nil)
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	client, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		log.Fatal(err)
	}

	privateKeyHex := "ac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(31337))
	if err != nil {
		log.Fatal(err)
	}

	// ABI 및 BIN 로드
	abiBytes, err := os.ReadFile("build/MyToken.abi")
	if err != nil {
		log.Fatal(err)
	}
	binBytes, err := os.ReadFile("build/MyToken.bin")
	if err != nil {
		log.Fatal(err)
	}

	tokenABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		log.Fatal(err)
	}

	// 컨트랙트 배포
	address, tx, _, err := bind.DeployContract(auth, tokenABI, common.FromHex(string(binBytes)), client, big.NewInt(1000000))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("ERC20 deployed at: %s\n", address.Hex())
	fmt.Printf("Tx hash: %s\n", tx.Hash().Hex())

	// 배포 트랜잭션 처리될 때까지 잠시 대기
	fmt.Println("Waiting for deployment...")
	time.Sleep(3 * time.Second)

	callOpts := &bind.CallOpts{
		From: auth.From,
	}

	tokenContract := bind.NewBoundContract(address, tokenABI, client, client, client)

	var result []interface{}
	err = tokenContract.Call(callOpts, &result, "testStringRevert")
	if err != nil {
		reason := parseRevertReason(err)
		fmt.Printf("Revert reason: %s\n", reason)
		return
	}

	fmt.Println("Transaction would succeed (unexpected).")
}
