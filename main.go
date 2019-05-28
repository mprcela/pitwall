package main

import (
	// disable looking for Consul in svckit/dcy package
	_ "github.com/minus5/svckit/dcy/lazy" // disable looking for Consul in svckit/dcy package

	"github.com/minus5/pitwall/cmd"
)

func main() {
	cmd.Execute()
}
