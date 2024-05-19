package fskit

import (
	"errors"
	"os"
	"path/filepath"
)

// Lookup 查找路径，如果不存在则向父路径查找
// tier 查找层级，0 不往父路径查找
func Lookup(filename string, tier int) (path string, err error) {
	for i := 0; i <= tier; i++ {
		if _, err = os.Stat(filename); errors.Is(err, os.ErrNotExist) {
			filename = filepath.Join("../", filename)
			continue
		}
		path, _ = filepath.Abs(filename)
		return path, nil
	}
	return "", os.ErrNotExist
}
