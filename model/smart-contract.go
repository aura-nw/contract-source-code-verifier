package model

import (
	"time"

	"gorm.io/gorm"
)

type SmartContract struct {
	Id                   int       `json:"id" gorm:"primary_key"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	Height               int       `json:"height"`
	CodeId               int       `json:"code_id"`
	ContractName         string    `json:"contract_name"`
	ContractAddress      string    `json:"contract_address"`
	CreatorAddress       string    `json:"creator_address"`
	ContractHash         string    `json:"contract_hash"`
	Url                  string    `json:"url"`
	InstantiateMsgSchema string    `json:"instantiate_msg_schema"`
	QueryMsgSchema       string    `json:"query_msg_schema"`
	ExecuteMsgSchema     string    `json:"execute_msg_schema"`
	ContractMatch        string    `json:"contract_match"`
	ContractVerification string    `json:"contract_verification"`
	CompilerVersion      string    `json:"compiler_version"`
	S3Location           string    `json:"s3_location"`
	ReferenceCodeId      int       `json:"reference_code_id"`
	MainnetUploadStatus  string    `json:"mainnet_upload_status"`
	TokenName            string    `json:"token_name"`
	TokenSymbol          string    `json:"token_symbol"`
	NumTokens            int       `json:"num_tokens"`
	VerifiedAt           time.Time `json:"verified_at"`
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
	err = db.Where("contract_hash = ? AND contract_verification = ?", hash, verification).First(SmartContract).Error
	if err != nil {
		return err
	}
	return nil
}

// Find smart contracts by hash
func GetSmartContractsByHash(db *gorm.DB, SmartContracts *[]SmartContract, hash string, verification string) (err error) {
	err = db.Where("contract_hash = ? AND contract_verification = ?", hash, verification).Find(&SmartContracts).Error
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
