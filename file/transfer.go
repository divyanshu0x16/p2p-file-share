package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const ChunkSize = 4096

func SaveFile(data []byte, filePath string, shareDir string) error {
	fullPath := filepath.Join(shareDir, filePath)

	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("couldn't create directory: %v", err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("couldn't create file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("couldn't write to file: %v", err)
	}

	return nil
}

func ReadFile(filePath string, shareDir string) ( []byte, error ){
	fullPath := filepath.Join(shareDir, filePath)

	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't access file: %v", err)
	}

	if info.IsDir(){
		return nil, fmt.Errorf("cannot transfer directories")
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("couldn't read file: %v", err)
	}

	return data, nil
}

// ChunkedCopy copies data from reader to writer in chunks, reporting progress
func ChunkedCopy(writer io.Writer, reader io.Reader, totalSize int64) error {
	buffer := make([]byte, ChunkSize)
	var totalRead int64

	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		if _, err := writer.Write(buffer[:n]); err != nil {
			return err
		}

		totalRead += int64(n)
		progress := float64(totalRead) / float64(totalSize) * 100
		fmt.Printf("\rProgress: %.1f%% (%d/%d bytes)", progress, totalRead, totalSize)
		
		if err == io.EOF {
			break
		}
	}
	fmt.Println()
	return nil
}