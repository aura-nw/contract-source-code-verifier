package util

import "smart-contract-verify/model"

func CustomResponse(code string, message string) model.JsonResponse {
	return model.JsonResponse{
		Code:    code,
		Message: message,
	}
}
