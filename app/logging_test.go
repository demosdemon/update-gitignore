package app

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	defer InitLogging()()
	os.Exit(m.Run())
}
