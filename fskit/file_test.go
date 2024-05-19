package fskit_test

import (
	"testing"

	"github.com/fengjx/go-halo/fskit"
)

func TestLookup(t *testing.T) {
	absPath, err := fskit.Lookup("fs/file.go", 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(absPath)
}
