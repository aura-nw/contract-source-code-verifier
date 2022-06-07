package model

type VerifyContractRequest struct {
	ContractUrl     string `json:"contractUrl"`
	Commit          string `json:"commit"`
	ContractAddress string `json:"contractAddress"`
	CompilerVersion string `json:"compilerVersion"`
	// CompilerImage   string `json:"compilerImage"`
}
