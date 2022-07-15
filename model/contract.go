package model

type ContractInfo struct {
	CodeId    string `json:"code_id"`
	Creator   string `json:"creator"`
	Admin     string `json:"admin"`
	Label     string `json:"label"`
	Created   string `json:"created"`
	IbcPortId string `json:"ibc_port_id"`
	Extension string `json:"extension"`
}

type Contract struct {
	Address      string       `json:"address"`
	ContractInfo ContractInfo `json:"contract_info"`
}
