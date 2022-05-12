package model

import (
	"time"

	"gorm.io/gorm"
)

type SmartContract struct {
	Id              int
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ContractAddress string
	CreatorAddress  string
	Schema          string
	Url             string
}

// Find a smart contract
func GetSmartContract(db *gorm.DB, SmartContract *SmartContract, address string) (err error) {
	err = db.Where("contract_address = ?", address).First(SmartContract).Error
	if err != nil {
		return err
	}
	return nil
}

// Update smart contract
func UpdateSmartContract(db *gorm.DB, SmartContract *SmartContract) (err error) {
	db.Save(SmartContract)
	return nil
}
