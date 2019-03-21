package app

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	shutdown := InitLogging()
	exit := m.Run()
	if err := shutdown(); err != nil {
		panic(err)
	}
	os.Exit(exit)
}
