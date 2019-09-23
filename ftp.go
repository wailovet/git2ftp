package git2ftp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/secsy/goftp"
	"io/ioutil"
	"path/filepath"
)

type ftpConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Path     string `json:"path"`
}

type git2ftpConfig struct {
	Ftp []ftpConfig `json:"ftp"`
}

func InitGit2ftpConfig() git2ftpConfig {
	current, _ := GetCurrentPath()
	confPath, _ := filepath.Abs(current + "/git2ftp.json")
	ftpc := ftpConfig{}
	git2ftpc := git2ftpConfig{}
	if !IsExist(confPath) {
		fmt.Println("git2ftp未被初始化!")
		fmt.Println("输入要上传的ftp主机:")
		fmt.Scanln(&ftpc.Host)
		fmt.Println("输入ftp账号:")
		fmt.Scanln(&ftpc.User)
		fmt.Println("输入ftp密码:")
		fmt.Scanln(&ftpc.Password)
		ftpc.Port = "21"
		git2ftpc.Ftp = append(git2ftpc.Ftp, ftpc)
		result, _ := json.Marshal(git2ftpc)
		ioutil.WriteFile(confPath, result, 644)
	}

	jsonText, _ := ioutil.ReadFile(confPath)
	json.Unmarshal(jsonText, &git2ftpc)
	return git2ftpc
}

func FtpRead(client *goftp.Client, path string) (string, error) {
	var buf bytes.Buffer
	err := client.Retrieve(path, &buf)
	return buf.String(), err
}
