package cloud

import (
	"context"
	"log"
	"smart-contract-verify/util"
	"time"

	"github.com/go-redis/redis/v8"
)

func ConnectRedis() (*redis.Client, context.Context) {
	// Load config
	config, _ := util.LoadConfig(".")

	// Create a new Redis Client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.REDIS_HOST + ":" + config.REDIS_PORT, // We connect to host redis, thats what the hostname of the redis service is set to in the docker-compose
		Password: "",                                          // The password IF set in the redis Config file
		DB:       0,
	})
	// Ping the Redis server and check if any errors occured
	err := redisClient.Ping(context.Background()).Err()
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

	return redisClient, ctx
}
