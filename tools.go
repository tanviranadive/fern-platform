//go:build tools
// +build tools

// Package tools imports various code generation tools
// required by the project but not used in the actual code.
package tools

import (
	_ "github.com/99designs/gqlgen"
	_ "github.com/99designs/gqlgen/graphql/introspection"
)
