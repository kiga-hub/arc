package utils

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
)

//Zip 压缩
func Zip(srcFile string, destZip string) error {
	zipfile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile)+"/")
		// header.Name = path
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err0 := os.Open(path)
			if err0 != nil {
				return err0
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})
	return err
}

//Unzip 解压
func Unzip(fileBuffer []byte, destDir string) error {
	reader := bytes.NewReader(fileBuffer)
	zipReader, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			err = os.MkdirAll(fpath, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//UnzipBuffer 解压
func UnzipBuffer(fileBuffer []byte, fileMap map[string]([]byte)) error {
	reader := bytes.NewReader(fileBuffer)
	zipReader, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		if !f.FileInfo().IsDir() {

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(inFile)
			if err != nil {
				return err
			}
			fileMap[f.Name] = buf.Bytes()
		}
	}
	return nil
}
