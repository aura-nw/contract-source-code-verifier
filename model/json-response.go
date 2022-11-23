package model

type JsonResponse struct {
	Code    string
	Message string
}

type RedisResponse struct {
	Code            string
	Message         string
	ContractAddress string
	Verified        bool
}

const (
	SUCCESSFUL                string = "SUCCESSFUL"
	SOURCE_CODE_INCORRECT            = "E001"
	WASM_FILE_INCORRECT              = "E002"
	INTERNAL_ERROR                   = "E003"
	CANT_GENERATE_SCHEMA             = "E004"
	CANT_CREATE_ZIP                  = "E005"
	SIMILAR_CONTRACT_DETECTED        = "E006"
	VERIFY_DUPLICATE_CODE            = "E007"
	GITHUB_404                       = "E008"
	WRONG_COMMIT                     = "E009"
	MISSING_CARGO_LOCK               = "E010"
)

var ResponseMessage = map[string]string{
	SUCCESSFUL:                "Smart contract verify successful",
	CANT_GENERATE_SCHEMA:      "Error generate schema files",
	CANT_CREATE_ZIP:           "Error zip contract source code",
	WASM_FILE_INCORRECT:       "Provided wasm file is incorrect",
	SOURCE_CODE_INCORRECT:     "Smart contract source code or compiler version is incorrect",
	INTERNAL_ERROR:            "Internal error",
	SIMILAR_CONTRACT_DETECTED: "This is a similar match to contract ",
	VERIFY_DUPLICATE_CODE:     "The code id of this contract is already being verified",
	GITHUB_404:                "Contract git URL not found",
	WRONG_COMMIT:              "Commit not found",
	MISSING_CARGO_LOCK:        "Missing Cargo.lock file",
}

const (
	EXACT_MATCH   string = "EXACT MATCH"
	SIMILAR_MATCH        = "SIMILAR MATCH"
	UNVERIFIED           = "UNVERIFIED"
)

const (
	STATUS_UNVERIFIED     string = "Unverified"
	STATUS_NOT_REGISTERED        = "Not registered"
	STATUS_TBD                   = "TBD"
	STATUS_DEPLOYED              = "Deployed"
	STATUS_REJECTED              = "Rejected"
)
