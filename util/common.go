package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os/exec"
	"smart-contract-verify/model"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func RemoveTempDir(dir string) error {
	_, err := exec.Command("rm", "-rf", dir).CombinedOutput()
	if err != nil {
		return err
	}
	return nil
}

func MakeTempDir() (string, []byte) {
	dir := "temp/tempdir" + fmt.Sprint(time.Now().Unix()) + fmt.Sprint(rand.Int())
	out, _ := exec.Command("mkdir", dir).CombinedOutput()
	return dir, out
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
			_ = RemoveTempDir(dir)
		}
		log.Println("Error publish to redis: " + err.Error())
		return
	}
}

func UploadContractToS3(g *gin.Context, contract model.SmartContract, ctx context.Context, redisClient *redis.Client, dir string, contractAddress string) string {
	// Load config
	config, _ := LoadConfig(".")

	session := g.MustGet("session").(*session.Session)
	uploader := s3manager.NewUploader(session)

	fileName := "code_id_" + strconv.Itoa(contract.CodeId) + ".zip"
	file, err := ioutil.ReadFile(dir + "/" + fileName)
	if err != nil {
		_ = RemoveTempDir(dir)
		log.Println("Error read source code zip file: " + err.Error())
		PublishRedisMessage(ctx, redisClient, contractAddress, config.REDIS_CHANNEL, "", false)
		return ""
	}

	up, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(config.BUCKET_NAME),
		Key:    aws.String(config.AWS_FOLDER + fileName),
		Body:   bytes.NewBuffer(file),
	})

	if err != nil {
		_ = RemoveTempDir(dir)
		log.Println("Error upload contract code to S3: " + err.Error())
		PublishRedisMessage(ctx, redisClient, contractAddress, config.REDIS_CHANNEL, "", false)
		return ""
	}
	log.Println("Upload contract code to S3 successful: ", up)

	return up.Location
}
