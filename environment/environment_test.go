package environment_test

import (
	"testing"

	"github.com/flowfunction/cddl/ast"
	env "github.com/flowfunction/cddl/environment"
)

type EnvInitializer func() *env.Environment

func TestEnvAdd(t *testing.T) {
	sharedEnv := env.NewEnvironment()
	shared := func() *env.Environment { return sharedEnv }

	tests := []struct {
		ident   string
		value   ast.Node
		envInit EnvInitializer
		err     error
	}{
		// test in new environment
		{"name", &ast.TstrType{}, env.NewEnvironment, nil},

		// test shared environment
		{"age", &ast.IntegerLiteral{}, shared, nil},
		{"age", &ast.IntegerLiteral{}, shared, env.ErrSymbolExists},
	}

	for _, test := range tests {
		environ := test.envInit()
		err := environ.Add(test.ident, &test.value)
		if err != test.err {
			t.Fatal(err)
		}
	}
}
