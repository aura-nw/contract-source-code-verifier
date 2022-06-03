package model

import (
	"time"

	"gorm.io/gorm"
)

type SmartContract struct {
	Id                   int
	CreatedAt            time.Time
	UpdatedAt            time.Time
	Height               int
	CodeId               int
	ContractName         string
	ContractAddress      string
	CreatorAddress       string
	ContractHash         string
	Url                  string
	InstantiateMsgSchema string
	QueryMsgSchema       string
	ExecuteMsgSchema     string
	ContractMatch        string
	ContractVerification string
	CompilerVersion      string
}

// Find a smart contract
func GetSmartContract(db *gorm.DB, SmartContract *SmartContract, address string) (err error) {
	err = db.Where("contract_address = ?", address).First(SmartContract).Error
	if err != nil {
		return err
	}
	return nil
}

// Find the exact smart contract by hash
func GetExactSmartContractByHash(db *gorm.DB, SmartContract *SmartContract, hash string) (err error) {
	err = db.Where("contract_hash = ?", hash).Where("contract_verification = ?", EXACT_MATCH).First(SmartContract).Error
	if err != nil {
		return err
	}
	return nil
}

// Find the exact smart contract by hash
func GetUnverifiedSmartContractByHash(db *gorm.DB, SmartContract *[]SmartContract, hash string) (err error) {
	err = db.Where("contract_hash = ?", hash).Where("contract_verification = ?", UNVERIFIED).Find(SmartContract).Error
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

// Update smart contract
func UpdateMultipleSmartContract(db *gorm.DB, SmartContract *[]SmartContract) (err error) {
	db.Save(SmartContract)
	return nil
}
