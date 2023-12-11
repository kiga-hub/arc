package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"strings"
	"time"
)

//GetStoreDir  获存储文件夹
func GetStoreDir(rootDir string) (string, error) {
	subStoreDir := time.Now().Local().Format("2006-01-02/15")

	fullDir := fmt.Sprintf("%s/%s", rootDir, subStoreDir)
	if _, err := os.Stat(fullDir); os.IsNotExist(err) {
		// 必须分成两步：先创建文件夹、再修改权限
		if err := os.MkdirAll(fullDir, os.ModePerm); err != nil {
			return "", err
		}
		if err := os.Chmod(fullDir, 0777); err != nil {
			return "", err
		}
	}
	return fullDir, nil
}

//CopyFile 复制文件
func CopyFile(dstName, srcName string) (int64, error) {
	src, err := os.Open(srcName)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	return io.Copy(dst, src)
}

//MoveFile 删除文件
func MoveFile(fileName string, rootDir string) (string, error) {
	index := strings.LastIndex(fileName, "/")
	file := ""
	if -1 != index {
		prefix := []byte(fileName)[index+1:]
		file = string(prefix)
	}

	timeStr := time.Now().Local().Format("20060102150405")
	dstFileName := fmt.Sprintf("%s_%s", timeStr, file)

	storeDir, err := GetStoreDir(rootDir)
	if err != nil {
		return "", err
	}

	storePath := fmt.Sprintf("%s/%s", storeDir, dstFileName)
	if _, err := os.Stat(storePath); os.IsExist(err) {
		return "", err
	}

	err = os.MkdirAll(storeDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	_, err = CopyFile(storePath, fileName)
	if err != nil {
		return "", err
	}

	// TODO this could be improved
	return storePath[len(rootDir)+1:], nil
}

//StoreFile 存储文件
func StoreFile(srcFile multipart.File, rootDir string, fileName string) (string, error) {
	timeStr := time.Now().Local().Format("20060102150405")
	dstFileName := fmt.Sprintf("%s_%s", timeStr, fileName)
	storeDir, err := GetStoreDir(rootDir)
	if err != nil {
		return "", err
	}

	storePath := fmt.Sprintf("%s/%s", storeDir, dstFileName)
	if _, err := os.Stat(storePath); os.IsExist(err) {
		return "", err
	}

	dst, err := os.Create(storePath)
	defer func() {
		err := dst.Close()
		if err != nil {
			fmt.Println("storeFile dst.Close() err:", err)
		}
	}()
	if err != nil {
		return "", err
	}

	if _, err = io.Copy(dst, srcFile); err != nil {
		return "", err
	}

	return storePath[len(rootDir)+1:], nil
}
