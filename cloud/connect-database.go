package cloud

import (
	"smart-contract-verify/util"

	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB

func InitDb() *gorm.DB {
	Db = connectDB()
	return Db
}

func connectDB() *gorm.DB {
	// Load config
	config, _ := util.LoadConfig(".")

	dsn := config.DB_USER + ":" + config.DB_PASS + "@tcp" + "(" + config.DB_HOST + ":" + config.DB_PORT + ")/" + config.DB_NAME + "?" + "parseTime=true&loc=UTC"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		fmt.Printf("Error connecting to database : %v\n", err)
		return nil
	}

	return db
}
