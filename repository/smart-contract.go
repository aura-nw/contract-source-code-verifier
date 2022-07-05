package repository

import (
	"bytes"
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
	"strconv"
	"strings"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
	log.Println("Verify contract request: ", request)

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
		hash, dir := service.GetContractHash(strconv.Itoa(contract.CodeId), config.RPC)
		if hash == "" {
			log.Println("Cannot get contract hash")
			PublishRedisMessage(ctx, redisClient, request.ContractAddress, config.REDIS_CHANNEL, "", false)
			return
		}
		log.Println("Result get contract hash: ", hash)
		contractHash = hash
		_ = service.RemoveTempDir(dir)

		var exactContract model.SmartContract
		err = model.GetExactSmartContractByHash(repository.Db, &exactContract, contractHash)
		log.Println("Result get exact contract by hash: ", exactContract)
		if err == nil {
			contract.ContractVerification = model.SIMILAR_MATCH
			contract.ContractMatch = exactContract.ContractAddress
			contract.ContractHash = exactContract.ContractHash
			contract.Url = exactContract.Url
			contract.InstantiateMsgSchema = exactContract.InstantiateMsgSchema
			contract.QueryMsgSchema = exactContract.QueryMsgSchema
			contract.ExecuteMsgSchema = exactContract.ExecuteMsgSchema
			contract.CompilerVersion = exactContract.CompilerVersion
			contract.S3Location = exactContract.S3Location

			g.BindJSON(&contract)
			err = model.UpdateSmartContract(repository.Db, &contract)
			if err != nil {
				_ = service.RemoveTempDir(dir)
				log.Println("Error update smart contract: " + err.Error())
				return
			}
			PublishRedisMessage(ctx, redisClient, request.ContractAddress, config.REDIS_CHANNEL, dir, true)
			return
		}
	}

	fmt.Println("Start verifying smart contract source code")
	var contractDir string
	if request.ContractDir != "" {
		contractDir = string(request.ContractDir[1:len(request.ContractDir)])
	} else {
		contractDir = request.ContractDir
	}
	verify, dir, contractFolder := service.VerifyContractCode(request.ContractUrl, request.Commit, contractHash, request.CompilerVersion, config.RPC, request.WasmFile, contractDir, strconv.Itoa(contract.CodeId))

	if verify {
		fmt.Println("Verify smart contract successful")
		session := g.MustGet("session").(*session.Session)
		uploader := s3manager.NewUploader(session)

		fileName := "code_id_" + strconv.Itoa(contract.CodeId) + ".zip"
		file, err := ioutil.ReadFile(dir + "/" + fileName)
		if err != nil {
			_ = service.RemoveTempDir(dir)
			log.Println("Error read source code zip file: " + err.Error())
			PublishRedisMessage(ctx, redisClient, request.ContractAddress, config.REDIS_CHANNEL, "", false)
			return
		}

		//upload to the s3 bucket
		up, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(config.BUCKET_NAME),
			Key:    aws.String(config.AWS_FOLDER + fileName),
			Body:   bytes.NewBuffer(file),
		})

		if err != nil {
			_ = service.RemoveTempDir(dir)
			log.Println("Error upload contract code to S3: " + err.Error())
			PublishRedisMessage(ctx, redisClient, request.ContractAddress, config.REDIS_CHANNEL, "", false)
			return
		}
		log.Println("Upload contract code to S3 successful: ", up)

		schemaDir := dir + "/" + contractFolder
		if request.ContractDir != "" {
			schemaDir = schemaDir + request.ContractDir
		}
		schemaDir = schemaDir + config.SCHEMA_DIR
		files, err := ioutil.ReadDir(schemaDir)
		if err != nil {
			_ = service.RemoveTempDir(dir)
			log.Println("Error read schema dir: " + err.Error())
			PublishRedisMessage(ctx, redisClient, request.ContractAddress, config.REDIS_CHANNEL, "", false)
			return
		}

		var instantiateSchema string
		var querySchema string
		var executeSchema string
		for _, file := range files {
			schemaFile := schemaDir + file.Name()
			data, err := ioutil.ReadFile(schemaFile)
			if err != nil {
				_ = service.RemoveTempDir(dir)
				log.Println("Error read schema file: " + err.Error())
				PublishRedisMessage(ctx, redisClient, request.ContractAddress, config.REDIS_CHANNEL, "", false)
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
		contract.ContractVerification = model.EXACT_MATCH
		contract.ContractHash = contractHash
		contract.Url = gitUrl
		contract.CompilerVersion = request.CompilerVersion
		contract.InstantiateMsgSchema = instantiateSchema
		contract.QueryMsgSchema = querySchema
		contract.ExecuteMsgSchema = executeSchema
		contract.S3Location = up.Location

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
		PublishRedisMessage(ctx, redisClient, request.ContractAddress, config.REDIS_CHANNEL, dir, true)
	} else {
		PublishRedisMessage(ctx, redisClient, request.ContractAddress, config.REDIS_CHANNEL, dir, false)
	}
	redisClient.Close()

	err = service.RemoveTempDir(dir)
	if err != nil {
		log.Println("Error remove temp dir: " + err.Error())
		return
	}
}

func PublishRedisMessage(ctx context.Context, redisClient *redis.Client, contractAddress string, redisChannel string, dir string, verified bool) {
	result := model.RedisResponse{
		ContractAddress: contractAddress,
		Verified:        verified,
	}
	res, _ := json.Marshal(result)

	err := redisClient.Publish(ctx, redisChannel, string(res)).Err()
	if err != nil {
		if dir != "" {
			_ = service.RemoveTempDir(dir)
		}
		log.Println("Error publish to redis: " + err.Error())
		return
	}
}

// @BasePath /api/v1
// TestUploadToS3 godoc
// @Summary Test upload S3
// @Description Upload S3
// @Tags smart-contract
// @Accept  json
// @Produce  json
// @Success 200 {object} model.JsonResponse
// @Router /smart-contract/test-upload-s3 [post]
func (repository *SmartContractRepo) TestUploadToS3(g *gin.Context) {
	response := model.JsonResponse{}

	file, _ := ioutil.ReadFile("test-2.zip")

	uploadLocation := service.UploadContractCode(g, "test-2.zip", file)
	response = model.JsonResponse{
		Message: uploadLocation,
	}

	g.JSON(http.StatusOK, response)
}
