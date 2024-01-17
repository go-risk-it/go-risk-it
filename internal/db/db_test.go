package db

import (
	"testing"
)

func TestDummyDb(t *testing.T) {
	_, err := createContainerConnectionPool()

	if err != nil {
		t.Error("Failed to start test container")
	}

	//q := New(pool) non va
}
