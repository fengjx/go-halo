package fs_test

import (
	"testing"

	"github.com/fengjx/go-halo/fs"
)

func TestLookup(t *testing.T) {
	absPath, err := fs.Lookup("fs/file.go", 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(absPath)
}
