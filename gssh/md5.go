package gssh

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// Md5Sum 计算文件哈希值
func Md5Sum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	m := md5.New()
	buf := make([]byte, 1<<20) // 1M

	for {
		n, err := io.ReadFull(f, buf)
		if err == nil || err == io.ErrUnexpectedEOF {
			_, err = m.Write(buf[0:n])
			if err != nil {
				return "", err
			}

		} else if err == io.EOF {
			break
		} else {
			return "", err
		}
	}

	return hex.EncodeToString(m.Sum(nil)), nil
}

// GenMd5File 在相同目录下生成标准md5文件，文件名：原文件名.md5
func GenMd5File(file string) (string, error) {
	md5Str, err := Md5Sum(file)
	if err != nil {
		return "", fmt.Errorf("%v, file=%s", err, file)
	}

	data := md5Str + " " + filepath.Base(file)
	md5File := strings.TrimRight(file, " ") + ".md5"
	err = ioutil.WriteFile(md5File, []byte(data), 0666)
	if err != nil {
		return "", fmt.Errorf("%v, file=%s", err, file)
	}
	return md5File, nil
}
