package config

import (
	"os"
)

type TempFolderManager struct {
	folderPath string
}

func NewTempFolderManager() (TempFolderManager, error) {
	folderPath, err := os.MkdirTemp("", "")
	if err != nil {
		return TempFolderManager{}, err
	}
	return TempFolderManager{folderPath: folderPath}, nil
}

func (m TempFolderManager) WriteTempFile(contents string) (string, error) {
	file, err := os.CreateTemp(m.folderPath, "")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(file.Name(), []byte(contents), 0777)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func (m TempFolderManager) Cleanup() error {
	return os.RemoveAll(m.folderPath)
}
