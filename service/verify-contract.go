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
	// Get contract id based on contract address
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

	// Get uploaded wasm file based on code id
	out, err := exec.Command("aurad", "query", "wasm", "code", contractId, dir+config.UPLOAD_CONTRACT, "--node", rpc).CombinedOutput()
	if err != nil {
		_ = util.RemoveTempDir(dir)
		log.Println("Execute command error: " + string(out))
		log.Println("Error download contract: " + err.Error())
		return "", ""
	}
	log.Println("Result call contract with ID: " + string(out))

	// Generate contract hash
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

func VerifyContractCode(request model.VerifyContractRequest, contractHash string, contractDir string, codeId string) (bool, string, string) {
	// Load config
	config, _ := util.LoadConfig(".")

	contractFolder := request.ContractUrl[strings.LastIndex(request.ContractUrl, "/")+1 : len([]rune(request.ContractUrl))]

	dir, out := util.MakeTempDir()
	artifactsWasm := dir + "/" + contractFolder
	artifactsWasm = artifactsWasm + config.ARTIFACTS + request.WasmFile

	pwd, _ := exec.Command("pwd").CombinedOutput()

	// Clone and check out commit of contract
	util.CloneAndCheckOutContract(dir+"/"+contractFolder, request.ContractUrl, request.Commit)

	// Compile contract
	compiled := util.CompileSourceCode(request.CompilerVersion, strings.TrimSuffix(string(pwd), "\n")+"/"+dir+"/"+contractFolder, contractFolder+"_cache")
	if !compiled {
		return false, dir, contractFolder
	}

	// Get hash of compiled wasm file
	codeHash, err := exec.Command("sha256sum", artifactsWasm).CombinedOutput()
	if err != nil {
		_ = util.RemoveTempDir(dir)
		log.Println("Execute command error: " + string(codeHash))
		log.Println("Error get contract hash: " + err.Error())
		return false, dir, contractFolder
	}
	log.Println("Result GetContractHash: " + string(codeHash))

	// Check if hashes are match
	if strings.Split(string(codeHash), " ")[0] != contractHash {
		_ = util.RemoveTempDir(dir)
		return false, dir, contractFolder
	}

	// Generate schema file
	out, err = exec.Command("sh", "-c", "cd "+dir+"/"+contractFolder+"/"+contractDir+" && cargo clean && cargo schema").CombinedOutput()
	if err != nil {
		_ = util.RemoveTempDir(dir)
		log.Println("Execute command error: " + string(out))
		log.Println("Error generate schema files: " + err.Error())
		return false, dir, contractFolder
	}
	log.Println("Result generate schema files: " + string(out))

	// Zip contract source code
	cmd := exec.Command("sh", "-c", "cd "+dir+" && zip -r "+config.ZIP_PREFIX+codeId+"_"+request.ContractAddress+".zip "+contractFolder)
	err = cmd.Run()
	if err != nil {
		_ = util.RemoveTempDir(dir)
		log.Println("Error zip contract: " + err.Error())
		return false, dir, contractFolder
	}

	return true, dir, contractFolder
}
