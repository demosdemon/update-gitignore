package app

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	shutdown := InitLogging()
	exit := m.Run()
	if err := shutdown(); err != nil {
		panic(err)
	}
	time.Sleep(time.Second)
	os.Exit(exit)
}
