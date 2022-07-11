package util

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"smart-contract-verify/model"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
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

func PublishRedisMessage(ctx context.Context, redisClient *redis.Client, contractAddress string, redisChannel string, dir string, verified bool, code string, message string) {
	result := model.RedisResponse{
		Code:            code,
		Message:         message,
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

	fileName := config.ZIP_PREFIX + strconv.Itoa(contract.CodeId) + "_" + contractAddress + ".zip"
	file, err := ioutil.ReadFile(dir + "/" + fileName)
	if err != nil {
		_ = RemoveTempDir(dir)
		log.Println("Error read source code zip file: " + err.Error())
		PublishRedisMessage(ctx, redisClient, contractAddress, config.REDIS_CHANNEL, dir, false, model.INTERNAL_ERROR, model.ResponseMessage[model.INTERNAL_ERROR])
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
		PublishRedisMessage(ctx, redisClient, contractAddress, config.REDIS_CHANNEL, dir, false, model.INTERNAL_ERROR, model.ResponseMessage[model.INTERNAL_ERROR])
		return ""
	}
	log.Println("Upload contract code to S3 successful: ", up)

	return up.Location
}

func CompileSourceCode(compilerImage string, contractDir string, contractCache string) bool {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println("Error create docker client: " + err.Error())
		return false
	}

	reader, err := cli.ImagePull(ctx, compilerImage, types.ImagePullOptions{})
	if err != nil {
		log.Println("Error pull compiler image: " + err.Error())
		return false
	}

	defer reader.Close()
	io.Copy(os.Stdout, reader)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: compilerImage,
		Tty:   true,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: contractDir,
				Target: "/code",
			},
			{
				Type:   mount.TypeVolume,
				Source: contractCache,
				Target: "/code/target",
			},
			{
				Type:   mount.TypeVolume,
				Source: "registry_cache",
				Target: "/usr/local/cargo/registry",
			},
		},
	}, nil, nil, "")
	if err != nil {
		log.Println("Error create container: " + err.Error())
		return false
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		log.Println("Error start container: " + err.Error())
		return false
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	select {
	case err := <-errCh:
		if err != nil {
			log.Println("Error wait for container to finish running: " + err.Error())
			return false
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		log.Println("Error get container logs: " + err.Error())
		return false
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	newStr := buf.String()
	log.Println("Compile contract log:", newStr)

	if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{}); err != nil {
		log.Println("Error remove container: " + err.Error())
		return false
	}

	return true
}

func CloneAndCheckOutContract(contractDir string, contractUrl string, contractHash string) {
	contract, err := git.PlainClone(contractDir, false, &git.CloneOptions{
		URL:      contractUrl,
		Progress: os.Stdout,
	})
	if err != nil {
		log.Fatal(err)
	}
	hash := plumbing.NewHash(contractHash)
	workTree, _ := contract.Worktree()
	_ = workTree.Checkout(&git.CheckoutOptions{
		Hash: hash,
	})
}
