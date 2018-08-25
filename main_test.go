package main

import (
	"testing"

	"ttp.sh/frost/internal/manager"
)

func TestComposer(t *testing.T) {
	manager.Run("./tests")
}
