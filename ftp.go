package git2ftp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/secsy/goftp"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
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
		for strings.TrimSpace(ftpc.Host) == "" {
			fmt.Println("ftp域名或IP:")
			fmt.Scanln(&ftpc.Host)
		}
		fmt.Println("端口号[默认21]:")
		fmt.Scanln(&ftpc.Port)
		if strings.TrimSpace(ftpc.Port) == "" {
			ftpc.Port = "21"
		}

		fmt.Println("ftp账号:")
		fmt.Scanln(&ftpc.User)

		fmt.Println("ftp密码:")
		fmt.Scanln(&ftpc.Password)

		fmt.Println("ftp部署路径[默认为'/']:")
		fmt.Scanln(&ftpc.Path)
		if strings.TrimSpace(ftpc.Path) == "" {
			ftpc.Path = "/"
		}

		git2ftpc.Ftp = append(git2ftpc.Ftp, ftpc)
		result, _ := json.Marshal(git2ftpc)
		ioutil.WriteFile(confPath, result, 644)

		log.Println("ftp初始化完成,请重新执行命令.")
		os.Exit(0)
	}

	jsonText, _ := ioutil.ReadFile(confPath)
	json.Unmarshal(jsonText, &git2ftpc)
	return git2ftpc
}

func FtpAbs(path string) string {
	return strings.Replace(path, "//", "/", -1)
}

func FtpRead(client *goftp.Client, path string) (string, error) {
	var buf bytes.Buffer
	err := client.Retrieve(path, &buf)
	return buf.String(), err
}

func FtpIsExist(client *goftp.Client, path string) bool {
	path = FtpAbs(path)
	_, err := client.Stat(path)
	if err != nil {
		return false
	}
	return true
}
func FtpAutoMkdir(client *goftp.Client, remotePath string) error {
	dir, err := client.Stat(remotePath)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return errors.New("该路径存在非文件夹节点")
	}

	// Slow path: make sure parent exists and then call Mkdir for path.
	i := len(remotePath)
	for i > 0 && remotePath[i-1] == '/' { // Skip trailing path separator.
		i--
	}

	j := i
	for j > 0 && remotePath[j-1] != '/' { // Scan backward over element.
		j--
	}

	if j > 1 {
		// Create parent.
		err = FtpAutoMkdir(client, FtpAbs(remotePath[:j-1]))
		if err != nil {
			return err
		}
	}

	_, err = client.Mkdir(remotePath)
	if err != nil {
		return err
	}

	return nil
}
func FtpWriteByFile(client *goftp.Client, localPath string, remotePath string) error {

	remoteDir := path.Dir(remotePath)
	if !FtpIsExist(client, remoteDir) {
		err := FtpAutoMkdir(client, remoteDir)
		if err != nil {
			return err
		}
	}

	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	err = client.Store(remotePath, f)
	f.Close()
	if err != nil {
		return err
	}
	return nil
}

func FtpWrite(client *goftp.Client, remotePath string, data []byte) error {

	remoteDir := path.Dir(remotePath)
	if !FtpIsExist(client, remoteDir) {
		err := FtpAutoMkdir(client, remoteDir)
		if err != nil {
			return err
		}
	}

	var buf bytes.Buffer
	buf.Write(data)
	client.Store(remotePath, &buf)
	return nil
}

func GetHashByFtp(client *goftp.Client, item ftpConfig, currentTempGitDir string) string {
	currentTempGitDirConf, _ := filepath.Abs(currentTempGitDir + "/.git")
	_, err := client.Stat(FtpAbs(item.Path + "/git2ftp.hash"))
	if err != nil {
		log.Println(item.Host, ":", "hash文件不存在")

		logs, _ := Cmd("git", "--no-pager", `--git-dir=`+currentTempGitDirConf, `--work-tree=`+currentTempGitDir, "log", `--pretty=format:%H|%s`, "-30")
		logsList := strings.Split(logs, "\n")
		for i := range logsList {
			fmt.Println(fmt.Sprintf("[%d]:%s", i, logsList[i]))
		}
		i := -1
		fmt.Println(item.Host, ":", "选择当前FTP所存在的对应版本[输入编号]:")

		fmt.Scanln(&i)
		if i >= 0 {
			if len(strings.Split(logsList[i], "|")) > 0 {
				hashs := strings.Split(logsList[i], "|")[0]
				FtpWrite(client, FtpAbs(item.Path+"/git2ftp.hash"), []byte(hashs))
			}
		}
	}

	//获取线上版本号
	onlineHash, err := FtpRead(client, FtpAbs(item.Path+"/git2ftp.hash"))
	if err != nil {
		log.Fatalln(err.Error())
	}
	return onlineHash
}
