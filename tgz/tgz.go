package tgz

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/gommon/log"
)

func Compress(tgzName string, files ...string) error {
	created, err := os.Create(tgzName)
	if err != nil {
		log.Errorf("Error creating archive: %v", err)
		return err
	}
	defer func(created *os.File) {
		err := created.Close()
		if err != nil {
			log.Errorf("Error closing archive: %v", err)
		}
	}(created)
	gzipWriter := gzip.NewWriter(created)
	defer func(gzipWriter *gzip.Writer) {
		err := gzipWriter.Close()
		if err != nil {
			log.Errorf("Error closing gzip: %v", err)
		}
	}(gzipWriter)
	tarWriter := tar.NewWriter(gzipWriter)
	defer func(tarWriter *tar.Writer) {
		err := tarWriter.Close()
		if err != nil {
			log.Errorf("Error closing tar: %v", err)
		}
	}(tarWriter)

	for _, src := range files {
		info, err := os.Stat(src)
		if err != nil {
			log.Errorf("Error stating files: %v", err)
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
					log.Errorf("Error walking path: %v", err)
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
						log.Errorf("Error reading symlink: %v", err)
						return err
					}
					header.Linkname = link
					header.Typeflag = tar.TypeSymlink
				} else if info.Mode().IsRegular() {
					header.Typeflag = tar.TypeReg
				} else {
					log.Errorf("Error unsupported type: %c for %s", info.Mode().Type(), path)
					return fmt.Errorf("unsupported type: %c in %s", info.Mode().Type(), path)
				}
				//fmt.Println("baseDir:", baseDir)
				if err := tarWriter.WriteHeader(header); err != nil {
					log.Errorf("Error writing header: %v", err)
					return err
				}

				if info.IsDir() || info.Mode().Type()&os.ModeSymlink != 0 {
					return nil
				}

				file, err := os.Open(path)
				if err != nil {
					log.Errorf("Error opening file: %v", err)
					return err
				}
				defer func(file *os.File) {
					err := file.Close()
					if err != nil {
						log.Errorf("Error closing file: %v", err)
					}
				}(file)

				_, err = io.Copy(tarWriter, file)
				if err != nil {
					log.Errorf("Error copying file data: %v %s", err, path)
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

func Extract(name, dest string) error {
	file, err := os.Open(name)
	if err != nil {
		log.Errorf("Error opening file: %v", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf("Error closing file: %v", err)
		}
	}(file)
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		log.Errorf("Error reading gzip: %v", err)
		return err
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			log.Errorf("Error closing gzip: %v", err)
		}
	}(gzipReader)
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("Error reading tar: %v", err)
			return err
		}
		path := filepath.Join(dest, header.Name)
		info := header.FileInfo()
		switch header.Typeflag {
		case tar.TypeReg:
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
			if err != nil {
				log.Errorf("Error opening file: %v, %s", err, path)
				return err
			}
			_, err = io.Copy(file, tarReader)
			if err != nil {
				log.Errorf("Error copying file: %v", err)
				return err
			}
			err = file.Close()
			if err != nil {
				log.Errorf("Error closing file: %v", err)
				return err
			}
		case tar.TypeDir:
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				log.Errorf("Error creating directory: %v", err)
				return err
			}
		case tar.TypeSymlink:
			if err = os.Symlink(header.Linkname, path); err != nil {
				log.Errorf("Error creating symlink: %v", err)
				// return err
			}
		case tar.TypeXGlobalHeader:
			log.Debugf("Skipping %s of PAX records: %s", header.Name, header.PAXRecords)
		case tar.TypeLink:
			// Handle symbolic link
			targetPath := header.Linkname
			linkPath := header.Name

			log.Printf("Extracting symlink: %s -> %s\n", linkPath,
				targetPath)

			// Create the directory if it does not exist.
			dirPath := linkPath
			for i := len(linkPath) - 1; i > 0 && os.PathSeparator !=
				linkPath[i]; i-- {
				dirPath = linkPath[:i]
				_, err := os.Stat(dirPath)
				if os.IsNotExist(err) {
					os.MkdirAll(dirPath, 0755)
				}
			}

			// Create the symbolic link.
			err = os.Symlink(targetPath, linkPath)
			if err != nil {
				return fmt.Errorf("failed to create symlink %s: %w",
					linkPath, err)
			}

		default:
			log.Errorf("Error reading tar: unsupported type: %c in %s", header.Typeflag, header.Name)
			// return fmt.Errorf("unsupported type: %c in %s", header.Typeflag, header.Name)
		}
	}
	return nil
}

func FileIn(filename, tgzName string) bool {
	file, err := os.Open(tgzName)
	if err != nil {
		log.Errorf("Error opening file: %v", err)
		return false
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf("Error closing file: %v", err)
		}
	}(file)
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		log.Errorf("Error reading gzip: %v", err)
		return false
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			log.Errorf("Error closing gzip: %v", err)
		}
	}(gzipReader)
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("Error reading tar: %v", err)
			return false
		}
		if strings.HasSuffix(header.Name, filename) {
			return true
		}
	}
	return false
}

func List(tgzName string) ([]string, error) {
	result := make([]string, 8)
	file, err := os.Open(tgzName)
	if err != nil {
		log.Errorf("Error opening file: %v", err)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf("Error closing file: %v", err)
		}
	}(file)
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		log.Errorf("Error reading gzip: %v", err)
		return nil, err
	}
	defer func(gzipReader *gzip.Reader) {
		err := gzipReader.Close()
		if err != nil {
			log.Errorf("Error closing gzip: %v", err)
		}
	}(gzipReader)
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("Error reading tar: %v", err)
			return nil, err
		}
		switch header.Typeflag {
		case tar.TypeReg:
			result = append(result, fmt.Sprintf("File: %s", header.Name))
		case tar.TypeDir:
			result = append(result, fmt.Sprintf("Dir: %s", header.Name))
		case tar.TypeSymlink:
			result = append(result, fmt.Sprintf("Symlink: %s -> %s", header.Name, header.Linkname))
		case tar.TypeXGlobalHeader:
			log.Debugf("Skipping %s of PAX records: %s", header.Name, header.PAXRecords)

		default:
			log.Errorf("Error reading tar: unsupported type: %c in %s", header.Typeflag, header.Name)
			return nil, fmt.Errorf("unsupported type: %c in %s", header.Typeflag, header.Name)
		}
	}
	return result, nil
}
