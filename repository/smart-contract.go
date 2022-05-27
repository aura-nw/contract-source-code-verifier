package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"smart-contract-verify/database"
	"smart-contract-verify/model"
	"smart-contract-verify/service"
	"smart-contract-verify/util"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	InstantiateMsg string = "instantiate_msg.json"
	QueryMsg              = "query_msg.json"
	ExecuteMsg            = "execute_msg.json"
)

type SmartContractRepo struct {
	Db *gorm.DB
}

func New() *SmartContractRepo {
	db := database.InitDb()
	// db.AutoMigrate(&model.SmartContract{})
	return &SmartContractRepo{Db: db}
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
func (repository *SmartContractRepo) CallVerifyContractCode(g *gin.Context) {
	response := model.JsonResponse{}

	// Load config
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Panic("Cannot load config:", err)
	}

	params, err := ioutil.ReadAll(g.Request.Body)
	var request model.VerifyContractRequest
	err = json.Unmarshal(params, &request)
	if err != nil {
		fmt.Println("Can't unmarshal the byte array")
		return
	}

	var contract model.SmartContract
	err = model.GetSmartContract(repository.Db, &contract, request.ContractAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			g.AbortWithStatusJSON(http.StatusInternalServerError, util.CustomResponse(model.CONTRACT_ADDRESS_NOT_FOUND, model.ResponseMessage[model.CONTRACT_ADDRESS_NOT_FOUND]))
			return
		}

		g.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": err})
		return
	}

	verify, dir := service.VerifyContractCode(request.ContractUrl, request.Commit, contract.ContractHash, config.RPC)

	if verify {
		var gitUrl string
		if strings.Contains(request.ContractUrl, ".git") {
			gitUrl = request.ContractUrl[0 : strings.LastIndex(request.ContractUrl, ".")-1]
		} else {
			gitUrl = request.ContractUrl
		}
		gitUrl = gitUrl + "/commit/" + request.Commit
		contract.Url = gitUrl

		var exactContract model.SmartContract
		err = model.GetExactSmartContractByHash(repository.Db, &exactContract, contract.ContractHash)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			contract.ContractVerification = model.EXACT_MATCH
		} else {
			contract.ContractVerification = model.SIMILAR_MATCH
			contract.ContractMatch = exactContract.ContractAddress
		}

		g.BindJSON(&contract)
		err = model.UpdateSmartContract(repository.Db, &contract)
		if err != nil {
			_ = service.RemoveTempDir(dir)
			g.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": err})
			return
		}

		if contract.ContractVerification == model.EXACT_MATCH {
			var unverifiedContract []model.SmartContract
			err = model.GetUnverifiedSmartContractByHash(repository.Db, &unverifiedContract, contract.ContractHash)
			for i := 0; i < len(unverifiedContract); i++ {
				unverifiedContract[i].ContractMatch = contract.ContractAddress
				unverifiedContract[i].ContractVerification = model.SIMILAR_MATCH
				unverifiedContract[i].Url = gitUrl
			}

			g.BindJSON(&unverifiedContract)
			err = model.UpdateMultipleSmartContract(repository.Db, &unverifiedContract)
			if err != nil {
				_ = service.RemoveTempDir(dir)
				g.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": err})
				return
			}
		}

		response = util.CustomResponse(model.SUCCESSFUL, model.ResponseMessage[model.SUCCESSFUL])
	} else {
		response = util.CustomResponse(model.FAILED, model.ResponseMessage[model.FAILED])
	}

	err = service.RemoveTempDir(dir)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusInternalServerError, util.CustomResponse(model.CANT_REMOVE_CODE, model.ResponseMessage[model.CANT_REMOVE_CODE]))
		return
	}

	g.IndentedJSON(http.StatusOK, response)
}

// @BasePath /api/v1
// CallGetContractHash godoc
// @Summary Get the hash of a deployed contract
// @Description Return the hash of a contract provided its code Id
// @Tags smart-contract
// @Accept  json
// @Produce  json
// @Param contractId path string true "Get contract hash"
// @Success 200 {object} model.JsonResponse
// @Router /smart-contract/get-hash/{contractId} [get]
func (repository *SmartContractRepo) CallGetContractHash(g *gin.Context) {
	response := model.JsonResponse{}

	// Load config
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Panic("Cannot load config:", err)
	}

	contractId := g.Param("contractId")

	hash, dir := service.GetContractHash(contractId, config.RPC)
	if hash == "" {
		response = util.CustomResponse(model.ERROR_GET_HASH, model.ResponseMessage[model.ERROR_GET_HASH])
	} else {
		response = util.CustomResponse(model.SUCCESSFUL, hash)
	}

	err = service.RemoveTempDir(dir)
	if err != nil {
		g.AbortWithStatusJSON(http.StatusInternalServerError, util.CustomResponse(model.CANT_REMOVE_CODE, model.ResponseMessage[model.CANT_REMOVE_CODE]))
		return
	}

	g.IndentedJSON(http.StatusOK, response)
}

// @BasePath /api/v1
// TestQueryGetAll godoc
// @Summary Test get unverified contract
// @Description Return all unverified contract with provided hash
// @Tags smart-contract
// @Accept  json
// @Produce  json
// @Param contractHash path string true "Get list unverified contract"
// @Success 200 {object} model.JsonResponse
// @Router /smart-contract/get-unverified-contract/{contractHash} [get]
func (repository *SmartContractRepo) TestQueryGetAll(g *gin.Context) {
	response := model.JsonResponse{}

	contractHash := g.Param("contractHash")

	var unverifiedContract []model.SmartContract
	err := model.GetUnverifiedSmartContractByHash(repository.Db, &unverifiedContract, contractHash)
	if err != nil {
		response = util.CustomResponse(model.FAILED, err.Error())
	} else {
		log.Println(len(unverifiedContract))
		res, _ := json.Marshal(unverifiedContract)
		response = util.CustomResponse(model.SUCCESSFUL, string(res))
	}

	g.IndentedJSON(http.StatusOK, response)
}
