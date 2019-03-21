package app

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	shutdown := InitLogging()
	defer shutdown()
	os.Exit(m.Run())
}
