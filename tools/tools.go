// +build tools

// Do not compile this; it will cause an error
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

// To install a tool: go install -mod=vendor <tool path>
// Ex. go install -mod=vendor github.com/golang/mock/mockgen

package tools

import (
	_ "github.com/go-delve/delve/cmd/dlv"
	_ "github.com/golang/mock/mockgen"
	_ "github.com/pressly/goose/cmd/goose"
	_ "github.com/securego/gosec/cmd/gosec"
)
