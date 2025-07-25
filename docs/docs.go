// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {
            "name": "FeatureFlags API Support",
            "email": "support@featureflags.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://opensource.org/licenses/MIT"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "Check service health status",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "health"
                ],
                "summary": "Health check",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "service": {
                                    "type": "string"
                                },
                                "status": {
                                    "type": "string"
                                }
                            }
                        }
                    }
                }
            }
        },
        "/api/v1/flags": {
            "get": {
                "description": "Get all feature flags with their dependencies",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "flags"
                ],
                "summary": "List all flags",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "count": {
                                    "type": "integer"
                                },
                                "flags": {
                                    "type": "array",
                                    "items": {
                                        "$ref": "#/definitions/entity.Flag"
                                    }
                                }
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            },
            "post": {
                "description": "Create a new feature flag with optional dependencies",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "flags"
                ],
                "summary": "Create a new flag",
                "parameters": [
                    {
                        "description": "Flag creation request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/validator.FlagCreateRequest"
                        }
                    },
                    {
                        "type": "string",
                        "description": "Actor performing the action",
                        "name": "X-Actor",
                        "in": "header"
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created",
                        "schema": {
                            "$ref": "#/definitions/entity.Flag"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "409": {
                        "description": "Conflict",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/flags/{id}": {
            "get": {
                "description": "Get a specific feature flag by ID",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "flags"
                ],
                "summary": "Get a flag",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Flag ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/entity.Flag"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/flags/{id}/toggle": {
            "post": {
                "description": "Enable or disable a feature flag",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "flags"
                ],
                "summary": "Toggle a flag",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Flag ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "Toggle request",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/validator.FlagToggleRequest"
                        }
                    },
                    {
                        "type": "string",
                        "description": "Actor performing the action",
                        "name": "X-Actor",
                        "in": "header"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "flag_id": {
                                    "type": "integer"
                                },
                                "message": {
                                    "type": "string"
                                },
                                "status": {
                                    "type": "string"
                                }
                            }
                        }
                    },
                    "400": {
                        "description": "Bad Request - Missing dependencies",
                        "schema": {
                            "$ref": "#/definitions/DependencyError"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        },
        "/api/v1/flags/{id}/audit": {
            "get": {
                "description": "Get audit logs for a specific flag",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "audit"
                ],
                "summary": "Get flag audit logs",
                "parameters": [
                    {
                        "type": "integer",
                        "description": "Flag ID",
                        "name": "id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "properties": {
                                "audit_logs": {
                                    "type": "array",
                                    "items": {
                                        "$ref": "#/definitions/entity.AuditLog"
                                    }
                                },
                                "count": {
                                    "type": "integer"
                                }
                            }
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/ErrorResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "entity.AuditLog": {
            "type": "object",
            "properties": {
                "action": {
                    "type": "string"
                },
                "actor": {
                    "type": "string"
                },
                "created_at": {
                    "type": "string"
                },
                "flag_id": {
                    "type": "integer"
                },
                "id": {
                    "type": "integer"
                },
                "reason": {
                    "type": "string"
                }
            }
        },
        "entity.Flag": {
            "type": "object",
            "properties": {
                "created_at": {
                    "type": "string"
                },
                "dependencies": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                },
                "updated_at": {
                    "type": "string"
                }
            }
        },
        "validator.FlagCreateRequest": {
            "type": "object",
            "required": [
                "name"
            ],
            "properties": {
                "dependencies": {
                    "type": "array",
                    "items": {
                        "type": "integer"
                    }
                },
                "name": {
                    "type": "string",
                    "maxLength": 100,
                    "minLength": 3
                }
            }
        },
        "validator.FlagToggleRequest": {
            "type": "object",
            "required": [
                "reason"
            ],
            "properties": {
                "enable": {
                    "type": "boolean"
                },
                "reason": {
                    "type": "string",
                    "maxLength": 500,
                    "minLength": 3
                }
            }
        },
        "ErrorResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                }
            }
        },
        "DependencyError": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                },
                "missing_dependencies": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0.0",
	Host:             "localhost:8080",
	BasePath:         "",
	Schemes:          []string{"http"},
	Title:            "FeatureFlags API",
	Description:      "A robust backend service for managing feature flags with dependency support, audit logging, and circular dependency detection.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
} 