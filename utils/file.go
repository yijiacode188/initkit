package utils

import (
	"crypto/md5"
	"fmt"
	"io"
)

// FileGetMD5 计算文件的MD5值
func FileGetMD5(file io.Reader) (string, error) {
	hasher := md5.New()
	// 将文件内容复制到MD5计算器
	_, err := io.Copy(hasher, file)
	if err != nil {
		return "", err
	}

	// 计算文件的MD5值并返回十六进制字符串
	md5Hash := hasher.Sum(nil)
	return fmt.Sprintf("%x", md5Hash), nil
}
