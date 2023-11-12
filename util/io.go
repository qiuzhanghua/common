package util

import (
	"bufio"
	"fmt"
	"github.com/labstack/gommon/log"
	"io"
	"os"
)

func CopyFile(from, to string) error {
	buf := make([]byte, 1024)
	fin, err := os.Open(from)
	if err != nil {
		log.Errorf("Error opening file: %v", err)
		return err
	}
	defer func(fin *os.File) {
		err := fin.Close()
		if err != nil {
			log.Errorf("Error closing file: %v", err)
		}
	}(fin)
	fout, err := os.Create(to)
	if err != nil {
		log.Errorf("Error creating file: %v", err)
		return err
	}
	defer func(fout *os.File) {
		err := fout.Close()
		if err != nil {
			log.Errorf("Error closing file: %v", err)
		}
	}(fout)

	for {
		n, err := fin.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		if _, err := fout.Write(buf[:n]); err != nil {
			log.Errorf("Error writing file: %v", err)
			return err
		}
	}
	return nil
}

// ReadLines reads a whole file into memory
// and returns a slice of its lines.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
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

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// WriteLines writes the lines to the given file.
func WriteLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		log.Errorf("Error creating file: %v", err)
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Errorf("Error closing file: %v", err)
		}
	}(file)

	w := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := fmt.Fprintln(w, line)
		if err != nil {
			log.Errorf("Error writing file: %v", err)
			return err
		}
	}
	return w.Flush()
}
