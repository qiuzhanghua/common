package tzst

import (
	"fmt"
	"sync"

	"archive/tar"
	"github.com/klauspost/compress/zstd"
	"github.com/labstack/gommon/log"
	"github.com/qiuzhanghua/common/util"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Compress(tarZstName string, files ...string) error {
	created, err := os.Create(tarZstName)
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

	// Create Zstandard writer
	zstdWriter, err := zstd.NewWriter(created)
	if err != nil {
		log.Errorf("Error creating zstd writer: %v", err)
		return err
	}
	defer func(zstdWriter *zstd.Encoder) {
		err := zstdWriter.Close()
		if err != nil {
			log.Errorf("Error closing zstd: %v", err)
		}
	}(zstdWriter)

	tarWriter := tar.NewWriter(zstdWriter)
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

		err = filepath.Walk(src,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					log.Errorf("Error walking path: %v", err)
					return err
				}

				header, err := tar.FileInfoHeader(info, "")
				if err != nil {
					log.Errorf("Error creating tar header: %v", err)
					return err
				}

				// Set header name
				header.Name = path
				if baseDir != "" {
					header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, src))
				}

				// Clean up header name for cross-platform compatibility
				header.Name = filepath.ToSlash(header.Name)

				// Handle symlinks
				if info.Mode()&os.ModeSymlink != 0 {
					link, err := os.Readlink(path)
					if err != nil {
						log.Errorf("Error reading symlink: %v", err)
						return err
					}
					header.Linkname = link
				}

				if err := tarWriter.WriteHeader(header); err != nil {
					log.Errorf("Error writing header: %v", err)
					return err
				}

				// Don't write file content for directories or symlinks
				if info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
					return nil
				}

				// Write regular file content
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

	// Ensure all data is flushed
	if err := tarWriter.Flush(); err != nil {
		log.Errorf("Error flushing tar writer: %v", err)
		return err
	}

	return nil
}

func Extract(name, dest string) error {
	dest, err := util.ExpandHome(dest)
	if err != nil {
		log.Errorf("Error expanding home dir: %v", err)
		return err
	}
	dest, err = util.AbsPath(dest)
	if err != nil {
		log.Errorf("Error getting absolute path: %v", err)
		return err
	}

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

	// Create Zstandard reader
	zstdReader, err := zstd.NewReader(file)
	if err != nil {
		log.Errorf("Error creating zstd reader: %v", err)
		return err
	}
	defer zstdReader.Close()

	tarReader := tar.NewReader(zstdReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("Error reading tar: %v", err)
			return err
		}

		// Clean and secure the target path
		targetPath := filepath.Join(dest, header.Name)
		targetPath = filepath.Clean(targetPath)

		// Security check: ensure the target path is within destination directory
		if !strings.HasPrefix(targetPath, filepath.Clean(dest)+string(os.PathSeparator)) &&
			targetPath != filepath.Clean(dest) {
			log.Errorf("Security violation: trying to write outside destination directory: %s", header.Name)
			return fmt.Errorf("security violation: path traversal attempt")
		}

		info := header.FileInfo()

		switch header.Typeflag {
		case tar.TypeReg:
			// Ensure parent directory exists
			parentDir := filepath.Dir(targetPath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				log.Errorf("Error creating parent directory: %v", err)
				return err
			}

			// Create the file
			file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
			if err != nil {
				log.Errorf("Error opening file: %v, %s", err, targetPath)
				return err
			}

			// Copy file content
			_, err = io.Copy(file, tarReader)
			if err != nil {
				file.Close()
				log.Errorf("Error copying file: %v", err)
				return err
			}

			// Set file modification time
			if err := os.Chtimes(targetPath, header.AccessTime, header.ModTime); err != nil {
				log.Warnf("Could not set file times: %v", err)
			}

			err = file.Close()
			if err != nil {
				log.Errorf("Error closing file: %v", err)
				return err
			}

		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, info.Mode().Perm()); err != nil {
				log.Errorf("Error creating directory: %v", err)
				return err
			}

			// Set directory modification time
			if err := os.Chtimes(targetPath, header.AccessTime, header.ModTime); err != nil {
				log.Warnf("Could not set directory times: %v", err)
			}

		case tar.TypeSymlink:
			// Ensure parent directory exists
			parentDir := filepath.Dir(targetPath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				log.Errorf("Error creating parent directory: %v", err)
				return err
			}

			// Remove existing file/symlink
			if _, err := os.Lstat(targetPath); err == nil {
				if err := os.Remove(targetPath); err != nil {
					log.Errorf("Error removing existing file: %v", err)
					return err
				}
			}

			// Create the symlink
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				log.Errorf("Error creating symlink: %v", err)
				// Continue extracting other files
			}

		case tar.TypeLink:
			// Hard link handling
			targetLinkPath := filepath.Join(dest, header.Linkname)
			targetLinkPath = filepath.Clean(targetLinkPath)

			// Security check for hard link target
			if !strings.HasPrefix(targetLinkPath, filepath.Clean(dest)+string(os.PathSeparator)) &&
				targetLinkPath != filepath.Clean(dest) {
				log.Errorf("Security violation: hard link points outside destination: %s", header.Linkname)
				break // Skip this entry but continue extraction
			}

			// Ensure parent directory exists
			parentDir := filepath.Dir(targetPath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				log.Errorf("Error creating directory: %v", err)
				break
			}

			// Check if target exists (hard link source must exist)
			if _, err := os.Stat(targetLinkPath); os.IsNotExist(err) {
				log.Errorf("Hard link target does not exist: %s", header.Linkname)
				break
			}

			// Create hard link
			if err := os.Link(targetLinkPath, targetPath); err != nil {
				log.Errorf("Error creating hard link: %v", err)
			}

		case tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
			// Special files - usually skipped in most implementations
			log.Debugf("Skipping special file: %s (type: %c)", header.Name, header.Typeflag)

		case tar.TypeXGlobalHeader, tar.TypeXHeader:
			// PAX extended headers - skip but preserve information
			log.Debugf("Skipping PAX header: %s", header.Name)

		default:
			log.Errorf("Unsupported tar entry type: %c in %s", header.Typeflag, header.Name)
			// Skip unsupported types but continue extraction
		}
	}

	return nil
}

func FileIn(filename, tarZstName string) bool {
	file, err := os.Open(tarZstName)
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

	// Create Zstandard reader
	zstdReader, err := zstd.NewReader(file)
	if err != nil {
		log.Errorf("Error creating zstd reader: %v", err)
		return false
	}
	defer zstdReader.Close()

	tarReader := tar.NewReader(zstdReader)

	// Clean and prepare the filename for comparison
	searchFilename := filepath.Clean(filename)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("Error reading tar: %v", err)
			return false
		}

		// Clean the archive entry name
		entryName := filepath.Clean(header.Name)

		// Multiple matching strategies
		// 1. Exact match
		if entryName == searchFilename {
			return true
		}

		// 2. Check if filename is at the end of entry name (your original logic)
		if strings.HasSuffix(entryName, searchFilename) {
			// Additional check to ensure it's not a partial match
			// For example, "file.txt" should match "dir/file.txt" but not "myfile.txt"
			if entryName == searchFilename ||
				strings.HasSuffix(entryName, string(filepath.Separator)+searchFilename) {
				return true
			}
		}

		// 3. Check basename match
		if filepath.Base(entryName) == filepath.Base(searchFilename) {
			return true
		}

		// 4. Handle cases with trailing slashes (for directories)
		if searchFilename[len(searchFilename)-1] == filepath.Separator {
			cleanSearch := strings.TrimSuffix(searchFilename, string(filepath.Separator))
			if entryName == cleanSearch ||
				strings.HasPrefix(entryName, cleanSearch+string(filepath.Separator)) {
				return true
			}
		}
	}

	return false
}

func List(tarZstName string) ([]string, error) {
	file, err := os.Open(tarZstName)
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

	// Create Zstandard reader
	zstdReader, err := zstd.NewReader(file)
	if err != nil {
		log.Errorf("Error creating zstd reader: %v", err)
		return nil, err
	}
	defer zstdReader.Close()

	tarReader := tar.NewReader(zstdReader)
	result := make([]string, 0, 8) // Initialize with 0 length, capacity 8

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorf("Error reading tar: %v", err)
			return result, err // Return partial results on error
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

var zstdEncoderPool = sync.Pool{
	New: func() interface{} {
		enc, err := zstd.NewWriter(nil)
		if err != nil {
			log.Fatalf("Failed to create new Zstd Encoder: %v", err)
		}
		return enc
	},
}

var zstdDecoderPool = sync.Pool{
	New: func() interface{} {
		dec, err := zstd.NewReader(nil)
		if err != nil {
			log.Fatalf("Failed to create new Zstd Decoder: %v", err)
		}
		return dec
	},
}

// func Compress(tgzName string, files ...string) error {
// 	enc := zstdEncoderPool.Get().(*zstd.Encoder)
// 	defer zstdEncoderPool.Put(enc)
// 	return nil
// }

// func Extract(name, dest string) error {
// 	dec := zstdDecoderPool.Get().(*zstd.Decoder)
// 	defer zstdDecoderPool.Put(dec)
// 	return nil
// }

// func FileIn(filename, tgzName string) bool {
// 	dec := zstdDecoderPool.Get().(*zstd.Decoder)
// 	defer zstdDecoderPool.Put(dec)
// 	return false
// }

// func List(tgzName string) ([]string, error) {
// 	result := make([]string, 8)
// 	dec := zstdDecoderPool.Get().(*zstd.Decoder)
// 	defer zstdDecoderPool.Put(dec)
// 	return result, nil
// }
