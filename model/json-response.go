package model

type JsonResponse struct {
	Code    string
	Message string
}

type RedisResponse struct {
	ContractAddress string
	Verified        bool
}

const (
	SUCCESSFUL                 string = "SUCCESSFUL"
	FAILED                            = "FAILED"
	DIR_NOT_FOUND                     = "E001"
	READ_FILE_ERROR                   = "E002"
	CANT_REMOVE_CODE                  = "E003"
	CONTRACT_ADDRESS_NOT_FOUND        = "E004"
	CONTRACT_ID_NOT_FOUND             = "E005"
	ERROR_GET_HASH                    = "E006"
)

var ResponseMessage = map[string]string{
	SUCCESSFUL:                 "Smart contract verify successful",
	FAILED:                     "Smart contract verify failed",
	DIR_NOT_FOUND:              "Schema dir not found",
	READ_FILE_ERROR:            "Cannot read file in current directory",
	CANT_REMOVE_CODE:           "Cannot remove downloaded source code",
	CONTRACT_ADDRESS_NOT_FOUND: "Cannot find the provided contract address",
	CONTRACT_ID_NOT_FOUND:      "Cannot find the ID of provided contract address",
	ERROR_GET_HASH:             "Cannot get hash of contract",
}

const (
	EXACT_MATCH   string = "EXACT MATCH"
	SIMILAR_MATCH        = "SIMILAR MATCH"
	UNVERIFIED           = "UNVERIFIED"
)
