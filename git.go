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

func GetDiffFiles(currentTempGitDir string, onlineHash string) ([]string, error) {
	var result []string

	currentTempGitDirConf, _ := filepath.Abs(currentTempGitDir + "/.git")
	//查看差异文件
	diffFile, err := Cmd("git", "--no-pager", `--git-dir=`+currentTempGitDirConf, `--work-tree=`+currentTempGitDir, "diff", "--name-only", onlineHash)
	if err != nil {
		return nil, err
	}

	diffFileList := strings.Split(diffFile, "\n")
	for k := range diffFileList {
		diffFileList[k] = strings.TrimSpace(diffFileList[k])

		if diffFileList[k] != "" {
			absPath, _ := filepath.Abs(currentTempGitDir + "/" + diffFileList[k])
			if IsExist(absPath) {
				result = append(result, diffFileList[k])
				//log.Println("差异文件:", absPath, "-->", diffFileList[k])
			}
		}
	}
	return result, nil
}

func GitNowHash(currentTempGitDir string) (string, error) {
	currentTempGitDirConf, _ := filepath.Abs(currentTempGitDir + "/.git")
	//查看差异文件
	hash, err := Cmd("git", "--no-pager", `--git-dir=`+currentTempGitDirConf, `--work-tree=`+currentTempGitDir, "rev-parse", "HEAD")
	if err != nil {
		return "", err
	}

	return hash, nil
}
