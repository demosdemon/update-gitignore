package app

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	shutdown := InitLogging()
	exit := m.Run()
	shutdown()
	os.Exit(exit)
}
