package app

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	InitLogging()
	os.Exit(m.Run())
}
