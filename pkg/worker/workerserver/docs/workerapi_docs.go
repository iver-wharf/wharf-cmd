// Package docs GENERATED BY SWAG; DO NOT EDIT
// This file was generated by swaggo/swag
package docs

import "github.com/swaggo/swag"

const docTemplateworkerapi = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "Iver wharf-cmd support",
            "url": "https://github.com/iver-wharf/wharf-cmd/issues",
            "email": "wharf@iver.se"
        },
        "license": {
            "name": "MIT",
            "url": "https://github.com/iver-wharf/wharf-cmd/blob/master/LICENSE"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/": {
            "get": {
                "description": "Pong.\nAdded in v0.8.0.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "meta"
                ],
                "summary": "Ping",
                "operationId": "ping",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/workerserver.Ping"
                        }
                    }
                }
            }
        },
        "/api/artifact/{artifactId}/download": {
            "post": {
                "description": "Added in v0.8.0.",
                "produces": [
                    "application/octet-stream"
                ],
                "tags": [
                    "worker"
                ],
                "summary": "Download an artifact file.",
                "operationId": "downloadArtifact",
                "parameters": [
                    {
                        "minimum": 0,
                        "type": "integer",
                        "description": "Artifact ID",
                        "name": "artifactId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "file"
                        }
                    },
                    "400": {
                        "description": "Bad request",
                        "schema": {
                            "$ref": "#/definitions/problem.Response"
                        }
                    },
                    "404": {
                        "description": "Cannot find artifact",
                        "schema": {
                            "$ref": "#/definitions/problem.Response"
                        }
                    },
                    "502": {
                        "description": "Canont read artifact file",
                        "schema": {
                            "$ref": "#/definitions/problem.Response"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "problem.Response": {
            "type": "object",
            "properties": {
                "detail": {
                    "description": "Detail is a human-readable explanation specific to this occurrence of the\nproblem.\n\nRecommended to have proper punctuation, and be capitalized,\nlike a sentence. Compared to Title this field may stretch on and be\nlonger.",
                    "type": "string",
                    "example": "Build requires input variable 'myInput' to be of type 'string', but got 'int' instead."
                },
                "errors": {
                    "description": "Error is an extended field for the regular Problem model defined in\nRFC-7807. It contains the string message of the error (if any).",
                    "type": "array",
                    "items": {
                        "type": "string"
                    },
                    "example": [
                        "strconv.ParseUint: parsing \"-1\": invalid syntax"
                    ]
                },
                "instance": {
                    "description": "Instance is a URI reference that identifies the specific occurrence of\nthe problem. It may or may not yield further information if dereferenced.",
                    "type": "string",
                    "example": "/projects/12345/builds/run/6789"
                },
                "status": {
                    "description": "Status is the HTTP status code generated by the origin server for this\noccurrence of the problem.",
                    "type": "integer",
                    "example": 400
                },
                "title": {
                    "description": "Title is a short, human-readable summary of the problem type.\nIt SHOULD NOT change from occurrence to ocurrence of the problem, except\nfor purposes of localization.\n\nRecommended to be kept brief, have proper punctuation, and be\ncapitalized, like a short sentence.",
                    "type": "string",
                    "example": "Invalid input variable for build."
                },
                "type": {
                    "description": "Type is a URI reference that identifies the problem type. The IETF\nRFC-7807 specification encourages that, when dereferenced, it provide\nhuman-readable documentation for the problem type (e.g., using HTML).\nWhen this member is not present, its value is assumed to be\n\"about:blank\".",
                    "type": "string",
                    "example": "https://wharf.iver.com/#/prob/build/run/invalid-input"
                }
            }
        },
        "workerserver.Ping": {
            "type": "object",
            "properties": {
                "message": {
                    "type": "string",
                    "example": "pong"
                }
            }
        }
    }
}`

// SwaggerInfoworkerapi holds exported Swagger Info so clients can modify it
var SwaggerInfoworkerapi = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "Wharf worker API",
	Description:      "REST API for wharf-cmd to access build results.\nPlease refer to the gRPC API for more endpoints.",
	InfoInstanceName: "workerapi",
	SwaggerTemplate:  docTemplateworkerapi,
}

func init() {
	swag.Register(SwaggerInfoworkerapi.InstanceName(), SwaggerInfoworkerapi)
}
