package git2ftp

import (
	"errors"
	"path/filepath"
	"regexp"
	"strings"
)

func GetGitRemoteUrl() (string, error) {
	result, err := Cmd("git", "remote", "-v")
	if err != nil {
		return "", err
	}

	re, _ := regexp.Compile(`https?:\/\/.* `)

	//查找符合正则的第一个
	all := re.FindAll([]byte(result), -1)
	if len(all) < 1 {
		return "", errors.New(result)
	}
	return strings.TrimSpace(string(all[0])), nil
}

func SyncGit(currentTempGitDir string) error {
	currentTempGitDirConf, _ := filepath.Abs(currentTempGitDir + "/.git")

	_, err := Cmd("git", `--git-dir=`+currentTempGitDirConf, `--work-tree=`+currentTempGitDir, "reset", "--hard")
	if err != nil {
		return err
	}
	_, err = Cmd("git", `--git-dir=`+currentTempGitDirConf, `--work-tree=`+currentTempGitDir, "pull", "origin", "master")
	if err != nil {
		return err
	}

	return nil
}
