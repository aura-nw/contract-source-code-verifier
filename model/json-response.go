package model

type JsonResponse struct {
	Code    string
	Message string
}

const (
	SUCCESSFUL                 string = "SUCCESSFUL"
	FAILED                            = "FAILED"
	DIR_NOT_FOUND                     = "E001"
	READ_FILE_ERROR                   = "E002"
	CANT_REMOVE_CODE                  = "E003"
	CONTRACT_ADDRESS_NOT_FOUND        = "E004"
)

var ResponseMessage = map[string]string{
	SUCCESSFUL:                 "Smart contract verify successful",
	FAILED:                     "Smart contract verify failed",
	DIR_NOT_FOUND:              "Schema dir not found",
	READ_FILE_ERROR:            "Cannot read file in current directory",
	CANT_REMOVE_CODE:           "Cannot remove downloaded source code",
	CONTRACT_ADDRESS_NOT_FOUND: "Cannot find the provided contract address",
}
