package model

type VerifyContractRequest struct {
	ContractUrl     string `json:"contractUrl"`
	Image           string `json:"image"`
	ContractAddress string `json:"contractAddress"`
	IsGithubUrl     bool   `json:"isGithubUrl"`
}
