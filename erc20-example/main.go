package main

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/hlibbc/go-practices/erc20-example/token"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ .env 파일 로드 실패")
	}
}

func main() {
	// 1. Ganache 연결
	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatalf("❌ Ganache 연결 실패: %v", err)
	}
	fmt.Println("✅ Ganache 연결됨")

	// 2. .env에서 키 및 주소 로딩
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	recipientHex := os.Getenv("RECIPIENT")

	privateKey, err := crypto.HexToECDSA(privateKeyHex[2:]) // "0x" 제거
	if err != nil {
		log.Fatalf("❌ 프라이빗 키 파싱 실패: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("❌ 공개 키 변환 실패")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	recipient := common.HexToAddress(recipientHex)

	fmt.Println("🧾 배포자 주소:", fromAddress.Hex())

	// 3. 트랜잭션 서명자 설정
	chainID := big.NewInt(1337)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatalf("❌ 트랜잭션 서명자 생성 실패: %v", err)
	}
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)

	// 4. 컨트랙트 배포
	initialSupply := new(big.Int).Mul(big.NewInt(1000), big.NewInt(1e18))
	address, tx, instance, err := token.DeployToken(auth, client, initialSupply)
	if err != nil {
		log.Fatalf("❌ 컨트랙트 배포 실패: %v", err)
	}
	fmt.Println("✅ 컨트랙트 배포 완료")
	fmt.Println("📦 컨트랙트 주소:", address.Hex())
	fmt.Println("🚀 배포 트랜잭션:", tx.Hash().Hex())

	// 5. 토큰 전송 (100 MTK)
	amount := new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18))
	tx2, err := instance.Transfer(auth, recipient, amount)
	if err != nil {
		log.Fatalf("❌ 토큰 전송 실패: %v", err)
	}
	fmt.Println("✅ 토큰 전송 완료")
	fmt.Println("💸 전송 트랜잭션:", tx2.Hash().Hex())

	// 6. 잔액 확인
	printBalance("배포자", instance, fromAddress)
	printBalance("수신자", instance, recipient)
}

// balanceOf() 호출 및 출력
func printBalance(name string, contract *token.Token, addr common.Address) {
	bal, err := contract.BalanceOf(&bind.CallOpts{}, addr)
	if err != nil {
		log.Fatalf("❌ %s 잔액 조회 실패: %v", name, err)
	}
	fmt.Printf("💰 %s 잔액: %s MTK\n", name, formatMTK(bal))
}

// wei -> MTK 변환
func formatMTK(amount *big.Int) string {
	f := new(big.Float).SetInt(amount)
	unit := new(big.Float).SetFloat64(1e18)
	val := new(big.Float).Quo(f, unit)
	return val.Text('f', 2)
}
