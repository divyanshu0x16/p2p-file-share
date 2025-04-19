package file

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type FileInfo struct {
	Name string `json:"name"`
	Size int64 `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsDir bool `json:"is_dir"`
	OwnerID string `json:"owner_id"`
}

type FileList struct {
	PeerID string `json:"peer_id"`
	Files []FileInfo `json:"files"`
}

func ScanDirectory(dir string, peerID string) ( *FileList, error ){
	fileList := &FileList{
		PeerID: peerID,
		Files: []FileInfo{},
	}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == dir {
			return nil
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			relPath = path //If we can't get relative path, use full path
		}

		fileInfo := FileInfo{
			Name: relPath,
			Size: info.Size(),
			ModTime: info.ModTime(),
			IsDir: info.IsDir(),
			OwnerID: peerID,
		}

		fileList.Files = append(fileList.Files, fileInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error scanning directory: %v", err)
	}

	return fileList, nil
}

//Convert FileList to JSON
func (fl *FileList) ToJson() ([]byte, error){
	return json.Marshal(fl)
}

//Parse a JSON string to a FileList
func FromJSON(data []byte) (*FileList, error){
	var fileList FileList
	err := json.Unmarshal(data, &fileList)

	if err != nil {
		return nil, err
	}

	return &fileList, nil
}

func PrintFileList(fl *FileList){
	fmt.Printf("Files from %s:\n", fl.PeerID)
	fmt.Println("------------------------------------")
	for _, file := range fl.Files {
		fileType := "File"
		if file.IsDir {
			fileType = "Dir "
		}
		fmt.Printf("[%s] %-30s %8d bytes  %s\n", 
			fileType, 
			file.Name, 
			file.Size, 
			file.ModTime.Format("2006-01-02 15:04:05"))
	}
	fmt.Println("------------------------------------")
}