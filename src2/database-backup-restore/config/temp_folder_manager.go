package config

import (
	"io/ioutil"
	"os"
)

type TempFolderManager struct {
	folderPath string
}

func NewTempFolderManager() (TempFolderManager, error) {
	folderPath, err := ioutil.TempDir("", "")
	if err != nil {
		return TempFolderManager{}, err
	}
	return TempFolderManager{folderPath: folderPath}, nil
}

func (m TempFolderManager) WriteTempFile(contents string) (string, error) {
	file, err := ioutil.TempFile(m.folderPath, "")
	if err != nil {
		return "", err
	}

	err = ioutil.WriteFile(file.Name(), []byte(contents), 0777)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func (m TempFolderManager) Cleanup() error {
	return os.RemoveAll(m.folderPath)
}
