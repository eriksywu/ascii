package filestore

import (
	"fmt"
	"github.com/eriksywu/ascii/pkg/image"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ image.ImageStore = (*FileStore)(nil)

// Simple store to save to local file
type FileStore struct {
	rootPath string
}

func NewStore(path string) (*FileStore, error) {
	path = filepath.Clean(path)
	if fileInfo, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, os.ModeDir); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if !fileInfo.IsDir() {
		return nil, fmt.Errorf("given path %s is not a directory", path)
	}
	return &FileStore{
		rootPath: path,
	}, nil
}

func (f FileStore) PushASCIIImage(asciiImage string, id uuid.UUID) error {
	file, err := os.Create(filepath.Join(f.rootPath, id.String()))
	if err != nil {
		return err
	}
	_, err = file.WriteString(asciiImage)
	if err != nil {
		return err
	}
	return nil
}

func (f FileStore) GetASCIIImage(id uuid.UUID) (bool, string, error) {
	targetFile := filepath.Join(f.rootPath, id.String())
	if _, err := os.Stat(targetFile); os.IsNotExist(err) {
		return false, "", nil
	} else if err != nil {
		return false, "", err
	}
	content, err := ioutil.ReadFile(targetFile)
	if err != nil {
		return false, "", err
	}
	if err != nil {
		return false, "", err
	}
	return true, string(content), nil
}

func (f FileStore) ListASCIIImages() ([]uuid.UUID, error) {
	var images []uuid.UUID

	err := filepath.Walk(f.rootPath, func(path string, file os.FileInfo, err error) error {
		if file.IsDir() {
			return nil
		}
		_, imageFileName := filepath.Split(path)
		if id, err := uuid.Parse(imageFileName); err == nil {
			images = append(images, id)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return images, nil
}
