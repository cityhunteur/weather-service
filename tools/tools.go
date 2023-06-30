//go:build tools
// +build tools

// Package tools imports the tools used in this project.
package tools

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/vektra/mockery/v2"
)
