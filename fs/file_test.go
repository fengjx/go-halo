package fs_test

import (
	"testing"

	"github.com/fengjx/go-halo/fs"
)

func TestLookup(t *testing.T) {
	absPath, err := fs.Lookup("utils", 3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(absPath)
}
