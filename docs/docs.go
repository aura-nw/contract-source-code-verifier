// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "API Support",
            "email": "soberkoder@gmail.com"
        },
        "license": {
            "name": "Apache 2.0",
            "url": "http://www.apache.org/licenses/LICENSE-2.0.html"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/smart-contract/get-hash/{contractId}": {
            "get": {
                "description": "Return the hash of a contract provided its code Id",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "smart-contract"
                ],
                "summary": "Get the hash of a deployed contract",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Get contract hash",
                        "name": "contractId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.JsonResponse"
                        }
                    }
                }
            }
        },
        "/smart-contract/verify": {
            "post": {
                "description": "Compare if source code truely belongs to deployed smart contract",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "smart-contract"
                ],
                "summary": "Verify a smart contract source code",
                "parameters": [
                    {
                        "description": "Verify smart contract source code",
                        "name": "verify-contract-request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/model.VerifyContractRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/model.JsonResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "model.JsonResponse": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "model.VerifyContractRequest": {
            "type": "object",
            "properties": {
                "commit": {
                    "type": "string"
                },
                "contractAddress": {
                    "type": "string"
                },
                "contractUrl": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "Verify Smart Contract API",
	Description:      "This is a smart contract verify application",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
