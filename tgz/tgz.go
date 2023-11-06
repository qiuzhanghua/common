package tgz

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Compress(tgzName string, files ...string) error {
	created, err := os.Create(tgzName)
	if err != nil {
		log.Error().Msgf("Error creating archive: %v", err)
		return err
	}
	defer func(created *os.File) {
		err := created.Close()
		if err != nil {
			log.Error().Msgf("Error closing archive: %v", err)
		}
	}(created)
	gzipWriter := gzip.NewWriter(created)
	defer func(gzipWriter *gzip.Writer) {
		err := gzipWriter.Close()
		if err != nil {
			log.Error().Msgf("Error closing gzip: %v", err)
		}
	}(gzipWriter)
	tarWriter := tar.NewWriter(gzipWriter)
	defer func(tarWriter *tar.Writer) {
		err := tarWriter.Close()
		if err != nil {
			log.Error().Msgf("Error closing tar: %v", err)
		}
	}(tarWriter)

	for _, src := range files {
		info, err := os.Stat(src)
		if err != nil {
			log.Error().Msgf("Error stating files: %v", err)
			return err
		}
		var baseDir string
		if info.IsDir() {
			baseDir = filepath.Base(src)
		}
		//fmt.Println("baseDir:", baseDir)
		err = filepath.Walk(src,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					log.Error().Msgf("Error walking path: %v", err)
					return err
				}
				header := new(tar.Header)
				header.Name = path
				header.Size = info.Size()
				header.Mode = int64(info.Mode())
				header.ModTime = info.ModTime()
				header.AccessTime = info.ModTime()
				header.ChangeTime = info.ModTime()

				//fmt.Println("path:", path)
				//fmt.Println("info:", info.Name(), info.Mode(), info.Mode().Type()&os.ModeSymlink != 0, info.Sys())

				if baseDir != "" {
					header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, src))
				}

				if info.IsDir() {
					header.Name += "/"
					header.Typeflag = tar.TypeDir
				} else if info.Mode().Type()&os.ModeSymlink != 0 {
					link, err := os.Readlink(path)
					if err != nil {
						log.Error().Msgf("Error reading symlink: %v", err)
						return err
					}
					header.Linkname = link
					header.Typeflag = tar.TypeSymlink
				} else if info.Mode().IsRegular() {
					header.Typeflag = tar.TypeReg
				} else {
					log.Error().Msgf("Error unsupported type: %c for %s", info.Mode().Type(), path)
					return fmt.Errorf("unsupported type: %c in %s", info.Mode().Type(), path)
				}
				//fmt.Println("baseDir:", baseDir)
				if err := tarWriter.WriteHeader(header); err != nil {
					log.Error().Msgf("Error writing header: %v", err)
					return err
				}

				if info.IsDir() || info.Mode().Type()&os.ModeSymlink != 0 {
					return nil
				}

				file, err := os.Open(path)
				if err != nil {
					log.Error().Msgf("Error opening file: %v", err)
					return err
				}
				defer func(file *os.File) {
					err := file.Close()
					if err != nil {
						log.Error().Msgf("Error closing file: %v", err)
					}
				}(file)

				_, err = io.Copy(tarWriter, file)
				if err != nil {
					log.Error().Msgf("Error copying file data: %v %s", err, path)
					return err
				}
				return nil
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func Extract(name string, dest string) error {
	file, err := os.Open(name)
	if err != nil {
		log.Error().Msgf("Error opening file: %v", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error().Msgf("Error closing file: %v", err)
		}
	}(file)
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		log.Error().Msgf("Error reading gzip: %v", err)
		return err
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			log.Error().Msgf("Error closing gzip: %v", err)
		}
	}(gzipReader)
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Error().Msgf("Error reading tar: %v", err)
			return err
		}
		if header.Typeflag != tar.TypeReg && header.Typeflag != tar.TypeDir && header.Typeflag != tar.TypeSymlink {
			log.Error().Msgf("Error reading tar: unsupported type: %c in %s", header.Typeflag, header.Name)
			return fmt.Errorf("unsupported type: %c in %s", header.Typeflag, header.Name)
		}
		path := filepath.Join(dest, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				log.Error().Msgf("Error creating directory: %v", err)
				return err
			}
			continue
		} else if header.Typeflag == tar.TypeSymlink {
			if err = os.Symlink(header.Linkname, path); err != nil {
				log.Error().Msgf("Error creating symlink: %v", err)
				return err
			}
			continue
		} else {
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				log.Error().Msgf("Error opening file: %v, %s", err, path)
				return err
			}
			_, err = io.Copy(file, tarReader)
			if err != nil {
				log.Error().Msgf("Error copying file: %v", err)
				return err
			}
			err = file.Close()
			if err != nil {
				log.Error().Msgf("Error closing file: %v", err)
				return err
			}
		}
	}
	return nil
}

func FileIn(filename, tgzName string) bool {
	file, err := os.Open(tgzName)
	if err != nil {
		log.Error().Msgf("Error opening file: %v", err)
		return false
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error().Msgf("Error closing file: %v", err)
		}
	}(file)
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		log.Error().Msgf("Error reading gzip: %v", err)
		return false
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			log.Error().Msgf("Error closing gzip: %v", err)
		}
	}(gzipReader)
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Error().Msgf("Error reading tar: %v", err)
			return false
		}
		if strings.HasSuffix(header.Name, filename) {
			return true
		}
	}
	return false
}
