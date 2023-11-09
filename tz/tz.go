package tz

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"github.com/labstack/gommon/log"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func FileIn(filename, zipName string) bool {
	archive, err := zip.OpenReader(zipName)

	if err != nil {
		log.Errorf("Error opening archive: %v", err)
		return false
	}
	defer func(archive *zip.ReadCloser) {
		err := archive.Close()
		if err != nil {
			log.Errorf("Error closing archive: %v", err)
		}
	}(archive)

	for _, f := range archive.File {
		if strings.HasSuffix(f.Name, filename) {
			return true
		}
	}
	return false
}

func Extract(name, dest string) error {
	archive, err := zip.OpenReader(name)
	if err != nil {
		log.Errorf("Error opening archive: %v", err)
		return err
	}
	defer func(archive *zip.ReadCloser) {
		err := archive.Close()
		if err != nil {
			log.Errorf("Error closing archive: %v", err)
		}
	}(archive)
	linkMap := make(map[string]string)

	for _, f := range archive.File {
		filePath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		dir := filepath.Dir(filePath)
		_ = os.MkdirAll(dir, os.ModePerm)

		fileInArchive, err := f.Open()
		if f.Mode()&fs.ModeSymlink > 0 {
			buf := new(bytes.Buffer)
			_, err := io.Copy(buf, fileInArchive)
			if err != nil {
				log.Errorf("Error copying file: %v", err)
				return err
			}
			linkMap[f.Name] = buf.String()
			continue
		}

		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			log.Errorf("Error opening file: %v", err)
			return err
		}
		if _, err := io.Copy(destFile, fileInArchive); err != nil {
			log.Errorf("Error copying file: %v", err)
			return err
		}
		_ = destFile.Close()
		_ = fileInArchive.Close()
	}
	wd, err := os.Getwd()
	err = os.Chdir(dest)
	if err != nil {
		log.Errorf("Error changing directory: %v", err)
		return err
	}
	for k, v := range linkMap {
		err = os.Symlink(v, k)
		if err != nil {
			log.Errorf("Error creating symlink: %v", err)
			return err
		}
	}
	_ = os.Chdir(wd)
	return nil
}

func Compress(zipFile string, files ...string) error {
	f, err := os.Create(zipFile)
	if err != nil {
		log.Errorf("Error creating file: %v", err)
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Errorf("Error closing file: %v", err)
		}
	}(f)
	writer := zip.NewWriter(f)
	defer func(writer *zip.Writer) {
		err := writer.Close()
		if err != nil {
			log.Errorf("Error closing writer: %v", err)
		}
	}(writer)

	for _, file := range files {
		stat, err := os.Stat(file)
		if err != nil {
			return err
		}
		if stat.IsDir() {
			err = addDirToZip(writer, file)
			if err != nil {
				log.Errorf("Error adding dir to zip: %v", err)
				return err
			}
			continue
		} else if stat.Mode().IsRegular() {
			err := addFileToZip(writer, file)
			if err != nil {
				log.Errorf("Error adding file to zip: %v", err)
				return err
			}
		} else {
			return errors.New("unsupported file type for " + file)
		}
	}
	return nil
}

func List(zipFile string) ([]string, error) {
	result := make([]string, 8)

	archive, err := zip.OpenReader(zipFile)
	if err != nil {
		log.Errorf("Error opening archive: %v", err)
		return nil, err
	}
	defer func(archive *zip.ReadCloser) {
		err := archive.Close()
		if err != nil {
			log.Errorf("Error closing archive: %v", err)
		}
	}(archive)

	for _, f := range archive.File {
		info := f.FileInfo()
		if info.IsDir() {
			result = append(result, fmt.Sprintf("Dir: %s", f.Name))
			continue
		} else if info.Mode().Type() == fs.ModeSymlink {
			buf := new(bytes.Buffer)
			reader, err := f.Open()
			if err != nil {
				log.Errorf("Error opening Symlink: %v", err)
				return nil, err
			}
			_, err = io.Copy(buf, reader)
			if err != nil {
				log.Errorf("Error copying Symlink: %v", err)
				return nil, err
			}
			link := buf.String()
			result = append(result, fmt.Sprintf("Symlink: %s -> %s", f.Name, link))
			continue
		} else if info.Mode().IsRegular() {
			result = append(result, fmt.Sprintf("File: %s", f.Name))
			continue
		} else {
			return nil, errors.New("unknown file type for " + f.Name + " in zip")
		}
	}
	return result, nil
}

func addFileToZip(writer *zip.Writer, file string) error {
	info, err := os.Stat(file)
	if err != nil {
		log.Errorf("Error getting file info: %v", err)
		return err
	}
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		log.Errorf("Error creating header: %v", err)
		return err
	}
	header.Method = zip.Deflate
	header.Name = file
	headerWriter, err := writer.CreateHeader(header)
	if err != nil {
		log.Errorf("Error creating header: %v", err)
		return err
	}
	f, err := os.Open(file)
	if err != nil {
		log.Errorf("Error opening file: %v", err)
		return err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Errorf("Error closing file: %v", err)
		}
	}(f)
	_, err = io.Copy(headerWriter, f)
	return err
}

func addDirToZip(writer *zip.Writer, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Errorf("Error walking path: %v", err)
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			log.Errorf("Error creating header: %v", err)
			return err
		}
		header.Method = zip.Deflate
		header.Name, err = filepath.Rel(filepath.Dir(dir), path)
		if err != nil {
			log.Errorf("Error getting relative path: %v", err)
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}
		headerWriter, err := writer.CreateHeader(header)
		if err != nil {
			log.Errorf("Error creating header: %v", err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Mode().Type() == fs.ModeSymlink {
			link, err := os.Readlink(path)
			if err != nil {
				log.Errorf("Error reading symlink: %v", err)
				return err
			}
			_, err = headerWriter.Write([]byte(link))
			if err != nil {
				log.Errorf("Error writing symlink: %v", err)
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			log.Errorf("Skipping non regular file: %s", path)
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			log.Errorf("Error opening file: %v", err)
			return err
		}
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				log.Errorf("Error closing file: %v", err)
			}
		}(f)
		_, err = io.Copy(headerWriter, f)
		return err
	})
}
