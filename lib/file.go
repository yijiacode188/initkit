package lib

import (
	"os"
	"path/filepath"
)

func ListFiles(dir string) (error, []string) {
	var paths []string
	// 使用 filepath 包来遍历目录
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只列出文件
		if !info.IsDir() {
			paths = append(paths, path)

		}
		return nil
	})
	return err, paths
}
