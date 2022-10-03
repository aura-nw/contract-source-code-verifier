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
	S3Location           string
	ReferenceCodeId      int
	MainnetUploadStatus  string
	TokenName            string
	TokenSymbol          string
	NumTokens            int
	VerifiedAt           time.Time
}

// Find a smart contract
func GetSmartContract(db *gorm.DB, SmartContract *SmartContract, address string) (err error) {
	err = db.Where("contract_address = ?", address).First(SmartContract).Error
	if err != nil {
		return err
	}
	return nil
}

// Find smart contract by hash
func GetSmartContractByHash(db *gorm.DB, SmartContract interface{}, hash string, verification string) (err error) {
	err = db.Where("contract_hash = ?", hash).Where("contract_verification = ?", verification).First(SmartContract).Error
	if err != nil {
		return err
	}
	return nil
}

// Find smart contracts by hash
func GetSmartContractsByHash(db *gorm.DB, SmartContracts interface{}, hash string, verification string) (err error) {
	err = db.Where("contract_hash = ?", hash).Where("contract_verification = ?", verification).Find(&SmartContracts).Error
	if err != nil {
		return err
	}
	return nil
}

// Update smart contract(s)
func UpdateSmartContract(db *gorm.DB, SmartContract interface{}) (err error) {
	db.Save(SmartContract)
	return nil
}
