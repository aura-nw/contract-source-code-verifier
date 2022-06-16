package repository

import (
	"context"
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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

const (
	InstantiateMsg string = "instantiate_msg.json"
	QueryMsg              = "query_msg.json"
	ExecuteMsg            = "execute_msg.json"
	CW20ExecuteMsg        = "cw20_execute_msg.json"
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

	params, err := ioutil.ReadAll(g.Request.Body)
	var request model.VerifyContractRequest
	err = json.Unmarshal(params, &request)
	if err != nil {
		fmt.Println("Can't unmarshal the byte array")
		return
	}

	response = util.CustomResponse(model.SUCCESSFUL, "")
	g.JSON(http.StatusOK, response)

	go InstantResponse(repository, g, request)
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

	g.JSON(http.StatusOK, response)
}

func InstantResponse(repository *SmartContractRepo, g *gin.Context, request model.VerifyContractRequest) {
	// Load config
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Panic("Cannot load config:", err)
	}

	// Create a new Redis Client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.REDIS_HOST + ":" + config.REDIS_PORT, // We connect to host redis, thats what the hostname of the redis service is set to in the docker-compose
		Password: "",                                          // The password IF set in the redis Config file
		DB:       0,
	})
	// Ping the Redis server and check if any errors occured
	err = redisClient.Ping(context.Background()).Err()
	if err != nil {
		// Sleep for 3 seconds and wait for Redis to initialize
		time.Sleep(3 * time.Second)
		err := redisClient.Ping(context.Background()).Err()
		if err != nil {
			log.Println("Error ping redis: " + err.Error())
		}
	}
	// Generate a new background context that  we will use
	ctx := context.Background()

	err = redisClient.Set(ctx, request.ContractAddress, "Verifying", 0).Err()
	if err != nil {
		log.Println("Error set verifying process key value to redis: " + err.Error())
		return
	}

	var contract model.SmartContract
	err = model.GetSmartContract(repository.Db, &contract, request.ContractAddress)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("Contract address not found")
			return
		}

		log.Println("Error get smart contract data: " + err.Error())
		return
	}

	var contractHash string
	if contract.ContractHash != "" {
		contractHash = contract.ContractHash
	} else {
		contractId := service.GetContractId(contract.ContractAddress, config.RPC)
		if contractId == "" {
			log.Println("Error get contract Id: " + err.Error())
			return
		}

		hash, dir := service.GetContractHash(contractId, config.RPC)
		if hash == "" {
			return
		}
		_ = service.RemoveTempDir(dir)
	}

	fmt.Println("Start verifying smart contract source code")
	verify, dir, contractFolder := service.VerifyContractCode(request.ContractUrl, request.Commit, contractHash, request.CompilerVersion, config.RPC)

	if verify {
		fmt.Println("Verify smart contract successful")
		files, err := ioutil.ReadDir(dir + "/" + contractFolder + config.SCHEMA_DIR)
		if err != nil {
			log.Println("Error read schema dir: " + err.Error())
			return
		}

		var instantiateSchema string
		var querySchema string
		var executeSchema string
		for _, file := range files {
			log.Println(dir + "/" + contractFolder + config.SCHEMA_DIR + file.Name())
			data, err := ioutil.ReadFile(dir + "/" + contractFolder + config.SCHEMA_DIR + file.Name())
			if err != nil {
				_ = service.RemoveTempDir(dir)
				log.Println("Error read schema file: " + err.Error())
				return
			}

			if file.Name() == InstantiateMsg {
				instantiateSchema = string(data)
			} else if file.Name() == QueryMsg {
				querySchema = string(data)
			} else if file.Name() == ExecuteMsg || file.Name() == CW20ExecuteMsg {
				executeSchema = string(data)
			}
		}

		var gitUrl string
		if strings.Contains(request.ContractUrl, ".git") {
			gitUrl = request.ContractUrl[0 : strings.LastIndex(request.ContractUrl, ".")-1]
		} else {
			gitUrl = request.ContractUrl
		}
		gitUrl = gitUrl + "/commit/" + request.Commit
		contract.Url = gitUrl
		contract.CompilerVersion = request.CompilerVersion
		contract.InstantiateMsgSchema = instantiateSchema
		contract.QueryMsgSchema = querySchema
		contract.ExecuteMsgSchema = executeSchema

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
			log.Println("Error update smart contract: " + err.Error())
			return
		}

		if contract.ContractVerification == model.EXACT_MATCH {
			var unverifiedContract []model.SmartContract
			err = model.GetUnverifiedSmartContractByHash(repository.Db, &unverifiedContract, contract.ContractHash)
			for i := 0; i < len(unverifiedContract); i++ {
				unverifiedContract[i].ContractMatch = contract.ContractAddress
				unverifiedContract[i].ContractVerification = model.SIMILAR_MATCH
				unverifiedContract[i].Url = gitUrl
				unverifiedContract[i].CompilerVersion = request.CompilerVersion
				unverifiedContract[i].InstantiateMsgSchema = contract.InstantiateMsgSchema
				unverifiedContract[i].QueryMsgSchema = contract.QueryMsgSchema
				unverifiedContract[i].ExecuteMsgSchema = contract.ExecuteMsgSchema
			}

			g.BindJSON(&unverifiedContract)
			err = model.UpdateMultipleSmartContract(repository.Db, &unverifiedContract)
			if err != nil {
				_ = service.RemoveTempDir(dir)
				log.Println("Error update similar contract: " + err.Error())
				return
			}
		}
		result := model.RedisResponse{
			ContractAddress: request.ContractAddress,
			Verified:        true,
		}
		res, _ := json.Marshal(result)

		err = redisClient.Publish(ctx, config.REDIS_CHANNEL, string(res)).Err()
		if err != nil {
			_ = service.RemoveTempDir(dir)
			log.Println("Error publish to redis: " + err.Error())
			return
		}
	} else {
		result := model.RedisResponse{
			ContractAddress: request.ContractAddress,
			Verified:        false,
		}
		res, _ := json.Marshal(result)

		err = redisClient.Publish(ctx, config.REDIS_CHANNEL, string(res)).Err()
		if err != nil {
			_ = service.RemoveTempDir(dir)
			log.Println("Error publish to redis: " + err.Error())
			return
		}
	}
	redisClient.Close()

	err = service.RemoveTempDir(dir)
	if err != nil {
		log.Println("Error remove temp dir: " + err.Error())
		return
	}
}
