package tests

import (
	"log"
	"os"
	"testing"

	"ccards/tests/setup"
)

func TestMain(m *testing.M) {
	_, err := setup.SetupTestEnvironment()
	if err != nil {
		log.Fatalf("Failed to setup test environment: %v", err)
	}

	code := m.Run()

	setup.CleanupTestEnvironment()

	os.Exit(code)
}
