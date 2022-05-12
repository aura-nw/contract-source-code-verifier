package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"smart-contract-verify/database"
	"smart-contract-verify/model"
	"smart-contract-verify/service"
	"smart-contract-verify/util"
	"time"

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
	db.AutoMigrate(&model.SmartContract{})
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

	verify, dir := service.VerifyContractCode(request.ContractUrl, request.Image, request.ContractAddress, request.IsGithubUrl, config.RPC)

	if verify {
		files, err := ioutil.ReadDir(dir + config.DIR)
		if err != nil {
			g.AbortWithStatusJSON(http.StatusInternalServerError, util.CustomResponse(model.DIR_NOT_FOUND, model.ResponseMessage[model.DIR_NOT_FOUND]))
			return
		}

		var schema string
		for _, file := range files {
			data, err := ioutil.ReadFile(dir + config.DIR + file.Name())
			if err != nil {
				g.AbortWithStatusJSON(http.StatusInternalServerError, util.CustomResponse(model.READ_FILE_ERROR, model.ResponseMessage[model.READ_FILE_ERROR]))
				return
			}

			if file.Name() == InstantiateMsg || file.Name() == QueryMsg || file.Name() == ExecuteMsg {
				schema += string(data) + ";"
			}
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
		} else {
			contract.Schema = schema
			contract.UpdatedAt = time.Now()
			g.BindJSON(&contract)
			err = model.UpdateSmartContract(repository.Db, &contract)
			if err != nil {
				g.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error": err})
				return
			}
		}

		response = util.CustomResponse(model.SUCCESSFUL, model.ResponseMessage[model.SUCCESSFUL])
	} else {
		response = util.CustomResponse(model.FAILED, model.ResponseMessage[model.FAILED])
	}

	_, err = exec.Command("rm", "-rf", dir).CombinedOutput()
	if err != nil {
		g.AbortWithStatusJSON(http.StatusInternalServerError, util.CustomResponse(model.CANT_REMOVE_CODE, model.ResponseMessage[model.CANT_REMOVE_CODE]))
		return
	}

	g.IndentedJSON(http.StatusOK, response)
}
