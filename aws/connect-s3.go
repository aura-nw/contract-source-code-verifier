package aws

import (
	"log"
	"smart-contract-verify/util"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

func ConnectS3() *session.Session {
	// Load config
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Panic("Cannot load config:", err)
	}

	sess, err := session.NewSession(
		&aws.Config{
			Region: aws.String(config.AWS_REGION),
			Credentials: credentials.NewStaticCredentials(
				config.AWS_ACCESS_KEY_ID,
				config.AWS_SECRET_ACCESS_KEY,
				"",
			),
		},
	)
	if err != nil {
		log.Println("Error create session: " + err.Error())
	}
	return sess
}
