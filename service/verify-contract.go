package service

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os/exec"
	"smart-contract-verify/model"
	"smart-contract-verify/util"
	"strings"

	"log"
)

func GetContractId(contractAddress string, rpc string) string {
	out, err := exec.Command("aurad", "query", "wasm", "contract", contractAddress, "--node", rpc, "--output", "json").CombinedOutput() // , "| jq"
	if err != nil {
		log.Println("Execute command error: " + string(out))
		log.Println("Error get contract Id: " + err.Error())
		return ""
	}
	log.Println("Contract Info: " + string(out))

	contract := &model.Contract{}
	json.Unmarshal([]byte(out), contract)

	return contract.ContractInfo.CodeId
}

func GetContractHash(contractId string, rpc string) (string, string) {
	// Load config
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Panic("Cannot load config:", err)
	}

	dir, out, err := MakeTempDir()
	if err != nil {
		log.Println("Execute command error: " + string(out))
		log.Println("Error create dir to store code: " + err.Error())
		return "", ""
	}

	out, err = exec.Command("aurad", "query", "wasm", "code", contractId, dir+config.UPLOAD_CONTRACT, "--node", rpc).CombinedOutput()
	if err != nil {
		_ = RemoveTempDir(dir)
		log.Println("Execute command error: " + string(out))
		log.Println("Error download contract: " + err.Error())
		return "", ""
	}
	log.Println("Result call contract with ID: " + string(out))

	out, err = exec.Command("sha256sum", dir+config.UPLOAD_CONTRACT).CombinedOutput()
	if err != nil {
		_ = RemoveTempDir(dir)
		log.Println("Execute command error: " + string(out))
		log.Println("Error get contract hash: " + err.Error())
		return "", ""
	}
	log.Println("Result GetContractHash: " + string(out))

	hash := strings.Split(string(out), " ")[0]
	return hash, dir
}

func VerifyContractCode(contractUrl string, commit string, contractHash string, rpc string) (bool, string) {
	// hash, dir := GetContractHash(contractId, rpc)
	// if hash == "" {
	// 	return false, dir
	// }

	var contractFolder string
	if strings.Contains(contractUrl, ".git") {
		contractFolder = contractUrl[strings.LastIndex(contractUrl, "/")+1 : strings.LastIndex(contractUrl, ".")]
	} else {
		contractFolder = contractUrl[strings.LastIndex(contractUrl, "/")+1 : len([]rune(contractUrl))]
	}

	dir, out, err := MakeTempDir()

	out, err = exec.Command("/bin/bash", "./script/verify-contract.sh", contractUrl, commit, contractHash, dir, contractFolder).CombinedOutput()
	if err != nil {
		_ = RemoveTempDir(dir)
		log.Println("Execute command error: " + string(out))
		log.Println("Error verify smart contract code: " + err.Error())
		return false, dir
	}
	log.Println("Result VerifyContractCode: " + string(out))
	return true, dir
}

func RemoveTempDir(dir string) error {
	_, err := exec.Command("rm", "-rf", dir).CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func MakeTempDir() (string, []byte, error) {
	dir := "tempdir" + fmt.Sprint(rand.Int())
	out, err := exec.Command("mkdir", dir).CombinedOutput()
	return dir, out, err
}
