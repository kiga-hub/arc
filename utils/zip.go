package utils

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Zip compress
//
//goland:noinspection GoUnusedExportedFunction
func Zip(srcFile string, destZip string) error {
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer func() {
		if err := zipFile.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	archive := zip.NewWriter(zipFile)
	defer func() {
		if err := archive.Close(); err != nil {
			fmt.Println(err)
		}
	}()

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
			defer func() {
				if err := file.Close(); err != nil {
					fmt.Println(err)
				}
			}()
			_, err = io.Copy(writer, file)
		}
		return err
	})
	return err
}

// Unzip Decompress
//
//goland:noinspection GoUnusedExportedFunction
func Unzip(fileBuffer []byte, destDir string) error {
	reader := bytes.NewReader(fileBuffer)
	zipReader, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		err := func() error {
			fPath := filepath.Join(destDir, f.Name)
			if f.FileInfo().IsDir() {
				err = os.MkdirAll(fPath, os.ModePerm)
				if err != nil {
					return err
				}
			} else {
				if err = os.MkdirAll(filepath.Dir(fPath), os.ModePerm); err != nil {
					return err
				}

				inFile, err := f.Open()
				if err != nil {
					return err
				}
				defer func() {
					if err := inFile.Close(); err != nil {
						fmt.Println(err)
					}
				}()

				outFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					return err
				}
				defer func(outFile *os.File) {
					err := outFile.Close()
					if err != nil {
						fmt.Println(err)
					}
				}(outFile)

				_, err = io.Copy(outFile, inFile)
				if err != nil {
					return err
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

// UnzipBuffer Decompress
//
//goland:noinspection GoUnusedExportedFunction
func UnzipBuffer(fileBuffer []byte, fileMap map[string][]byte) error {
	reader := bytes.NewReader(fileBuffer)
	zipReader, err := zip.NewReader(reader, reader.Size())
	if err != nil {
		return err
	}

	for _, f := range zipReader.File {
		err := func() error {
			if !f.FileInfo().IsDir() {

				inFile, err := f.Open()
				if err != nil {
					return err
				}
				defer func(inFile io.ReadCloser) {
					err := inFile.Close()
					if err != nil {
						fmt.Println(err)
					}
				}(inFile)

				buf := new(bytes.Buffer)
				_, err = buf.ReadFrom(inFile)
				if err != nil {
					return err
				}
				fileMap[f.Name] = buf.Bytes()
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}
