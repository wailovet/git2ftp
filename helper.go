package git2ftp

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func IsExist(path string) bool {
	path, _ = filepath.Abs(path)
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}
func AutoMkdirAll(tmpFileDir string) error {
	tmpFileDir, _ = filepath.Abs(tmpFileDir)
	_, err := os.Stat(tmpFileDir)
	if err != nil {
		_ = os.MkdirAll(tmpFileDir, os.ModeDir)
	}
	return err
}
func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}

func Cmd(name string, arg ...string) (string, error) {
	cmd := exec.Command(name, arg...)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Start()
	if err != nil {
		return "", err
	}

	err = cmd.Wait()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}
