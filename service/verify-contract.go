package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/exec"
	"smart-contract-verify/model"
	"smart-contract-verify/util"
	"strings"

	"log"

	"github.com/gin-gonic/gin"
)

func GetContractId(contractAddress string, rpc string) string {
	out, err := exec.Command("aurad", "query", "wasm", "contract", contractAddress, "--node", rpc, "--output", "json").CombinedOutput() // , "| jq"
	if err != nil {
		log.Println("Execute command error: " + string(out))
		log.Println("Error get contract Id: " + err.Error())
	}
	log.Println("Contract Info: " + string(out))

	contract := &model.Contract{}
	json.Unmarshal([]byte(out), contract)

	return contract.ContractInfo.CodeId
}

func GetContractHash(contractAddress string, rpc string) string {
	contractId := GetContractId(contractAddress, rpc)

	out, err := exec.Command("aurad", "query", "wasm", "code", contractId, "tmpdir/contract.wasm", "--node", rpc).CombinedOutput()
	if err != nil {
		log.Println("Execute command error: " + string(out))
		log.Println("Error download contract: " + err.Error())
	}
	log.Println("Result call contract with ID: " + string(out))

	out, err = exec.Command("sha256sum", "tmpdir/contract.wasm").CombinedOutput()
	if err != nil {
		log.Println("Execute command error: " + string(out))
		log.Println("Error get contract hash: " + err.Error())
	}
	log.Println("Result GetContractHash: " + string(out))

	hash := strings.Split(string(out), " ")[0]
	return hash
}

func VerifyContractCode(contractUrl string, dockerImage string, contractAddress string, isGithubUrl bool, rpc string) bool {
	hash := GetContractHash(contractAddress, rpc)
	urlOption := "0"
	if isGithubUrl {
		urlOption = "1"
	}

	out, err := exec.Command("/bin/bash", "./script/verify-contract.sh", contractUrl, dockerImage, hash, urlOption).CombinedOutput()
	if err != nil {
		log.Println("Execute command error: " + string(out))
		log.Println("Error verify smart contract code: " + err.Error())
		return false
	}
	log.Println("Result VerifyContractCode: " + string(out))
	return true
}

// @BasePath /api/v1
// CallVerifyContractCode godoc
// @Summary Verify a smart contract source code
// @Description Compare if source code truely belongs to deployed smart contract
// @Tags smart-contract
// @Accept  json
// @Produce  json
// @Param verify-contract-request body model.VerifyContractRequest true "Verify smart contract source code"
// @Success 200 {object} model.JsonResponse
// @Router /smart-contract/verify [post]
func CallVerifyContractCode(g *gin.Context) {
	response := model.JsonResponse{}

	// Load config
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Cannot load config:", err)
	}

	params, err := ioutil.ReadAll(g.Request.Body)
	var request model.VerifyContractRequest
	err = json.Unmarshal(params, &request)
	if err != nil {
		fmt.Println("Can't unmarshal the byte array")
		return
	}

	verify := VerifyContractCode(request.ContractUrl, request.Image, request.ContractAddress, request.IsGithubUrl, config.RPC)

	if verify {
		// _, err := exec.Command("cargo", "schema").CombinedOutput()
		// if err != nil {
		// 	log.Println("Error generate schema: " + err.Error())
		// }

		// _, err = exec.Command("tar", "cvf", "schema.tar", "./schema/*.json").CombinedOutput()
		// if err != nil {
		// 	log.Println("Error compress schema: " + err.Error())
		// }

		response = util.CustomResponse(model.ResponseCode["SUCCESSFUL"], model.ResponseMessage["SUCCESSFUL"])
	} else {
		response = util.CustomResponse(model.ResponseCode["FAILED"], model.ResponseMessage["FAILED"])
	}

	g.IndentedJSON(http.StatusOK, response)
}
