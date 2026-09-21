// Command check-task-plugins validates embedded registration and SystemOne hooks.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/QuantumNous/new-api/pkg/jsplugin"
	"github.com/QuantumNous/new-api/plugins"
)

func main() {
	source, err := plugins.Source("typesafe")
	if err != nil {
		panic(err)
	}
	fixture, err := os.ReadFile("plugins/tasks/typesafe/fixture.json")
	if err != nil {
		panic(err)
	}
	report, err := jsplugin.ReplayFixture(context.Background(), source, fixture)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Embedded plugins loaded; TypeSafe fixtures passed: %d/%d\n", report.Passed, report.Total)
}
