package model

type JsonResponse struct {
	Code    string
	Message string
}

type RedisResponse struct {
	ContractAddress string
	Verified        bool
	Message         string
}

const (
	SUCCESSFUL                 string = "SUCCESSFUL"
	FAILED                            = "FAILED"
	READ_SCHEMA_ERROR                 = "E001"
	READ_FILE_ERROR                   = "E002"
	CANT_REMOVE_CODE                  = "E003"
	CONTRACT_ADDRESS_NOT_FOUND        = "E004"
	CONTRACT_ID_NOT_FOUND             = "E005"
	ERROR_GET_HASH                    = "E006"
	CANT_READ_ZIP                     = "E007"
	UPLOAD_S3_FAILED                  = "E008"
	READ_SCHEMA_FILE_ERROR            = "E009"
)

var ResponseMessage = map[string]string{
	SUCCESSFUL:                 "Smart contract verify successful",
	FAILED:                     "Smart contract verify failed",
	READ_SCHEMA_ERROR:          "Error read schema directory",
	READ_FILE_ERROR:            "Cannot read file in current directory",
	CANT_REMOVE_CODE:           "Cannot remove downloaded source code",
	CONTRACT_ADDRESS_NOT_FOUND: "Cannot find the provided contract address",
	CONTRACT_ID_NOT_FOUND:      "Cannot find the ID of provided contract address",
	ERROR_GET_HASH:             "Cannot get hash of contract",
	CANT_READ_ZIP:              "Cannot read zip file",
	UPLOAD_S3_FAILED:           "Cannot upload contract code to S3",
	READ_SCHEMA_FILE_ERROR:     "Error read schema file",
}

const (
	EXACT_MATCH   string = "EXACT MATCH"
	SIMILAR_MATCH        = "SIMILAR MATCH"
	UNVERIFIED           = "UNVERIFIED"
)
