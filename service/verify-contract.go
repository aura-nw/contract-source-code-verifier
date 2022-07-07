package service

import (
	"encoding/json"
	"os/exec"
	"smart-contract-verify/model"
	"smart-contract-verify/util"
	"strings"

	"log"
)

func GetContractId(contractAddress string, rpc string) string {
	out, err := exec.Command("aurad", "query", "wasm", "contract", contractAddress, "--node", rpc, "--output", "json").CombinedOutput()
	if err != nil {
		return ""
	}
	log.Println("Contract Info: " + string(out))

	contract := &model.Contract{}
	json.Unmarshal([]byte(out), contract)

	return contract.ContractInfo.CodeId
}

func GetContractHash(contractId string, rpc string) (string, string) {
	// Load config
	config, _ := util.LoadConfig(".")

	dir, out := util.MakeTempDir()

	out, err := exec.Command("aurad", "query", "wasm", "code", contractId, dir+config.UPLOAD_CONTRACT, "--node", rpc).CombinedOutput()
	if err != nil {
		_ = util.RemoveTempDir(dir)
		log.Println("Execute command error: " + string(out))
		log.Println("Error download contract: " + err.Error())
		return "", ""
	}
	log.Println("Result call contract with ID: " + string(out))

	out, err = exec.Command("sha256sum", dir+config.UPLOAD_CONTRACT).CombinedOutput()
	if err != nil {
		_ = util.RemoveTempDir(dir)
		log.Println("Execute command error: " + string(out))
		log.Println("Error get contract hash: " + err.Error())
		return "", ""
	}
	log.Println("Result GetContractHash: " + string(out))

	hash := strings.Split(string(out), " ")[0]
	return hash, dir
}

func VerifyContractCode(contractUrl string, commit string, contractHash string, compilerVersion string, rpc string, wasmFile string, contractDir string, codeId string) (bool, string, string) {
	contractFolder := contractUrl[strings.LastIndex(contractUrl, "/")+1 : len([]rune(contractUrl))]

	dir, out := util.MakeTempDir()
	tempDir := strings.Split(dir, "/")[len(strings.Split(dir, "/"))-1]

	out, err := exec.Command("/bin/bash", "./script/verify-contract.sh", contractUrl, commit, contractHash, dir, contractFolder, compilerVersion, wasmFile, contractDir, tempDir, codeId).CombinedOutput()
	if err != nil {
		_ = util.RemoveTempDir(dir)
		log.Println("Execute command error: " + string(out))
		log.Println("Error verify smart contract code: " + err.Error())
		return false, dir, contractFolder
	}
	log.Println("Result VerifyContractCode: " + string(out))
	return true, dir, contractFolder
}
