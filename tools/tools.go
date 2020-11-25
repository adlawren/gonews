// +build tools

// Do not compile this; it will cause an error
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

package tools

import (
	_ "github.com/golang/mock/mockgen"
)
