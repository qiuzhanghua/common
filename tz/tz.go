package tz

import (
	"archive/zip"
	"bytes"
	"github.com/rs/zerolog/log"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func FileIn(filename, zipName string) bool {
	archive, err := zip.OpenReader(zipName)

	if err != nil {
		log.Error().Msgf("Error opening archive: %v", err)
		return false
	}
	defer func(archive *zip.ReadCloser) {
		err := archive.Close()
		if err != nil {
			log.Error().Msgf("Error closing archive: %v", err)
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
		log.Error().Msgf("Error opening archive: %v", err)
		return err
	}
	defer func(archive *zip.ReadCloser) {
		err := archive.Close()
		if err != nil {
			log.Error().Msgf("Error closing archive: %v", err)
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
				log.Error().Msgf("Error copying file: %v", err)
				return err
			}
			linkMap[f.Name] = buf.String()
			continue
		}

		destFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			log.Error().Msgf("Error opening file: %v", err)
			return err
		}
		if _, err := io.Copy(destFile, fileInArchive); err != nil {
			log.Error().Msgf("Error copying file: %v", err)
			return err
		}
		_ = destFile.Close()
		_ = fileInArchive.Close()
	}
	wd, err := os.Getwd()
	err = os.Chdir(dest)
	if err != nil {
		log.Error().Msgf("Error changing directory: %v", err)
		return err
	}
	for k, v := range linkMap {
		err = os.Symlink(v, k)
		if err != nil {
			log.Error().Msgf("Error creating symlink: %v", err)
			return err
		}
	}
	_ = os.Chdir(wd)
	return nil
}
