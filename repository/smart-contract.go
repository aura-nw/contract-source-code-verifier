package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"smart-contract-verify/cloud"
	"smart-contract-verify/model"
	"smart-contract-verify/service"
	"smart-contract-verify/util"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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
	db := cloud.InitDb()
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
	config, _ := util.LoadConfig(".")

	params, err := ioutil.ReadAll(g.Request.Body)
	var request model.VerifyContractRequest
	err = json.Unmarshal(params, &request)
	if err != nil {
		fmt.Println("Can't unmarshal the byte array")
		return
	}
	log.Println("Verify contract request: ", request)

	var contract model.SmartContract
	if err := model.GetSmartContract(repository.Db, &contract, request.ContractAddress); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Println("Contract address not found")
			return
		}

		log.Println("Error get smart contract data: " + err.Error())
		return
	}

	// Initialize redis	client
	redisClient, ctx := cloud.ConnectRedis()

	keyCode := redisClient.Get(ctx, config.ZIP_PREFIX+strconv.Itoa(contract.CodeId)).Val()
	if keyCode != "" {
		log.Printf("Code Id %d is already been verified", contract.CodeId)
		response = util.CustomResponse(model.VERIFY_DUPLICATE_CODE, model.ResponseMessage[model.VERIFY_DUPLICATE_CODE])
		g.JSON(http.StatusOK, response)
		return
	}

	// Set verify status for current contract
	_ = redisClient.Set(ctx, request.ContractAddress, "Verifying", 0).Err()
	// Set key for a code id to specify as being verified
	_ = redisClient.Set(ctx, config.ZIP_PREFIX+strconv.Itoa(contract.CodeId), "Verifying", 0).Err()

	var contractHash string
	if contract.ContractHash != "" {
		contractHash = contract.ContractHash
	} else {
		hash, dir := service.GetContractHash(strconv.Itoa(contract.CodeId), config.RPC)
		if hash == "" {
			log.Println("Cannot get contract hash")
			util.PublishRedisMessage(
				ctx,
				redisClient,
				request.ContractAddress,
				config.REDIS_CHANNEL,
				"",
				false,
				model.WASM_FILE_INCORRECT,
				model.ResponseMessage[model.WASM_FILE_INCORRECT])
			return
		}
		log.Println("Result get contract hash: ", hash)
		contractHash = hash
		_ = util.RemoveTempDir(dir)

		var exactContract model.SmartContract
		if err := model.GetSmartContractByHash(repository.Db, &exactContract, contractHash, model.EXACT_MATCH); err == nil {
			contract.ContractVerification = model.SIMILAR_MATCH
			contract.ContractMatch = exactContract.ContractAddress
			contract.ContractHash = exactContract.ContractHash
			contract.Url = exactContract.Url
			contract.InstantiateMsgSchema = exactContract.InstantiateMsgSchema
			contract.QueryMsgSchema = exactContract.QueryMsgSchema
			contract.ExecuteMsgSchema = exactContract.ExecuteMsgSchema
			contract.CompilerVersion = exactContract.CompilerVersion
			contract.S3Location = exactContract.S3Location

			log.Println("Contract updated as similar: ", contract)
			g.BindJSON(&contract)
			if err = model.UpdateSmartContract(repository.Db, &contract); err != nil {
				_ = util.RemoveTempDir(dir)
				log.Println("Error update smart contract: " + err.Error())
				return
			}
			util.PublishRedisMessage(
				ctx,
				redisClient,
				request.ContractAddress,
				config.REDIS_CHANNEL,
				dir,
				true,
				model.SIMILAR_CONTRACT_DETECTED,
				model.ResponseMessage[model.SIMILAR_CONTRACT_DETECTED]+contract.ContractMatch)
			return
		}
		log.Println("Result get exact contract by hash: ", exactContract)
	}

	response = util.CustomResponse(model.SUCCESSFUL, "")
	redisClient.Close()

	g.JSON(http.StatusOK, response)

	go InstantResponse(repository, g, request, contract, contractHash)
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
	config, _ := util.LoadConfig(".")

	contractId := g.Param("contractId")

	hash, dir := service.GetContractHash(contractId, config.RPC)
	if hash == "" {
		response = util.CustomResponse(model.WASM_FILE_INCORRECT, model.ResponseMessage[model.WASM_FILE_INCORRECT])
	} else {
		response = util.CustomResponse(model.SUCCESSFUL, hash)
	}

	_ = util.RemoveTempDir(dir)

	g.JSON(http.StatusOK, response)
}

func InstantResponse(repository *SmartContractRepo, g *gin.Context, request model.VerifyContractRequest, contract model.SmartContract, contractHash string) {
	// Load config
	config, _ := util.LoadConfig(".")

	// Initialize redis	client
	redisClient, ctx := cloud.ConnectRedis()

	log.Println("Start verifying smart contract source code")
	var contractDir string
	if match, _ := regexp.MatchString(config.WORKSPACE_REGEX, request.CompilerVersion); match {
		exactContractFolder := strings.ReplaceAll(strings.Split(request.WasmFile, ".")[0], "_", "-")
		contractDir = config.WORKSPACE_DIR + exactContractFolder
	} else {
		contractDir = ""
	}
	verify, dir, contractFolder := service.VerifyContractCode(request, contractHash, contractDir, strconv.Itoa(contract.CodeId))

	switch verify {
	case model.SOURCE_CODE_INCORRECT:
		log.Println("Verify smart contract failed")
		util.PublishRedisMessage(
			ctx,
			redisClient,
			request.ContractAddress,
			config.REDIS_CHANNEL,
			dir,
			false,
			model.SOURCE_CODE_INCORRECT,
			model.ResponseMessage[model.SOURCE_CODE_INCORRECT])
	case model.WASM_FILE_INCORRECT:
		log.Println("Verify smart contract failed")
		util.PublishRedisMessage(
			ctx,
			redisClient,
			request.ContractAddress,
			config.REDIS_CHANNEL,
			dir,
			false,
			model.WASM_FILE_INCORRECT,
			model.ResponseMessage[model.WASM_FILE_INCORRECT])
	case model.SUCCESSFUL:
		log.Println("Verify smart contract successful")

		//upload to the s3 bucket
		s3Location := util.UploadContractToS3(g, contract, ctx, redisClient, dir, request.ContractAddress)
		if s3Location == "" {
			break
		}

		schemaDir := dir + "/" + contractFolder
		if match, _ := regexp.MatchString(config.WORKSPACE_REGEX, request.CompilerVersion); match {
			exactContractFolder := strings.ReplaceAll(strings.Split(request.WasmFile, ".")[0], "_", "-")
			schemaDir = schemaDir + "/" + config.WORKSPACE_DIR + exactContractFolder
		}
		schemaDir = schemaDir + config.SCHEMA_DIR
		files, err := ioutil.ReadDir(schemaDir)
		if err != nil {
			_ = util.RemoveTempDir(dir)
			log.Println("Error read schema dir: " + err.Error())
			util.PublishRedisMessage(
				ctx,
				redisClient,
				request.ContractAddress,
				config.REDIS_CHANNEL,
				dir,
				false,
				model.INTERNAL_ERROR,
				model.ResponseMessage[model.INTERNAL_ERROR])
			break
		}

		var instantiateSchema string
		var querySchema string
		var executeSchema string
		for _, file := range files {
			schemaFile := schemaDir + file.Name()
			data, err := ioutil.ReadFile(schemaFile)
			if err != nil {
				_ = util.RemoveTempDir(dir)
				log.Println("Error read schema file: " + err.Error())
				util.PublishRedisMessage(
					ctx,
					redisClient,
					request.ContractAddress,
					config.REDIS_CHANNEL,
					dir,
					false,
					model.INTERNAL_ERROR,
					model.ResponseMessage[model.INTERNAL_ERROR])
				break
			}

			switch file.Name() {
			case InstantiateMsg:
				instantiateSchema = string(data)
			case QueryMsg:
				querySchema = string(data)
			case ExecuteMsg, CW20ExecuteMsg:
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
		contract.ContractVerification = model.EXACT_MATCH
		contract.ContractHash = contractHash
		contract.Url = gitUrl
		contract.CompilerVersion = request.CompilerVersion
		contract.InstantiateMsgSchema = instantiateSchema
		contract.QueryMsgSchema = querySchema
		contract.ExecuteMsgSchema = executeSchema
		contract.S3Location = s3Location

		log.Println("Contract updated after verifying: ", contract)
		g.BindJSON(&contract)
		if err = model.UpdateSmartContract(repository.Db, &contract); err != nil {
			_ = util.RemoveTempDir(dir)
			log.Println("Error update smart contract: " + err.Error())
			break
		}

		if contract.ContractVerification == model.EXACT_MATCH {
			var unverifiedContract []model.SmartContract
			err = model.GetSmartContractByHash(repository.Db, &unverifiedContract, contract.ContractHash, model.UNVERIFIED)
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

			if err = model.UpdateSmartContract(repository.Db, &unverifiedContract); err != nil {
				_ = util.RemoveTempDir(dir)
				log.Println("Error update similar contract: " + err.Error())
				break
			}
		}
		util.PublishRedisMessage(
			ctx,
			redisClient,
			request.ContractAddress,
			config.REDIS_CHANNEL,
			dir,
			true,
			model.SUCCESSFUL,
			model.ResponseMessage[model.SUCCESSFUL])
	default:
		log.Println("Verify smart contract failed")
		util.PublishRedisMessage(
			ctx,
			redisClient,
			request.ContractAddress,
			config.REDIS_CHANNEL,
			dir,
			false,
			model.INTERNAL_ERROR,
			model.ResponseMessage[model.INTERNAL_ERROR])
	}
	redisClient.Del(ctx, config.ZIP_PREFIX+strconv.Itoa(contract.CodeId))
	redisClient.Close()

	_ = util.RemoveTempDir(dir)
}
