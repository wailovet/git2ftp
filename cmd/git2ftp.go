package main

import (
	"bytes"
	"fmt"
	"github.com/otiai10/copy"
	"github.com/secsy/goftp"
	"github.com/wailovet/git2ftp"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	git2ftpPath, _ := filepath.Abs(os.TempDir() + "/git2ftp")
	_ = git2ftp.AutoMkdirAll(git2ftpPath)

	currentPath, _ := git2ftp.GetCurrentPath()

	gitConfig, _ := filepath.Abs(currentPath + "/.git")
	if !git2ftp.IsExist(gitConfig) {
		log.Fatalln("当前目录不存在.git文件夹")
	}

	git2ftpConf := git2ftp.InitGit2ftpConfig()

	url, err := git2ftp.GetGitRemoteUrl()
	if err != nil {
		log.Fatalln(err.Error())
	}
	url = strings.Replace(url, "http://", "", -1)
	url = strings.Replace(url, "https://", "", -1)

	currentTempGitDir, _ := filepath.Abs(git2ftpPath + "/" + url)
	currentTempGitDirConf, _ := filepath.Abs(currentTempGitDir + "/.git")
	if !git2ftp.IsExist(currentTempGitDir) {
		copy.Copy(gitConfig, currentTempGitDirConf)
	}
	err = git2ftp.SyncGit(currentTempGitDir)
	if err != nil {
		log.Fatalln(err.Error())
	}
	//log.Println(currentTempGitDir)

	for e := range git2ftpConf.Ftp {
		item := git2ftpConf.Ftp[e]
		client, err := goftp.DialConfig(goftp.Config{
			User:     item.User,
			Password: item.Password,
		}, item.Host)
		if err != nil {
			panic(err)
		}
		_, err = client.Stat(item.Path + "/git2ftp.hash")
		if err != nil {
			log.Println(item.Host, ":", "hash文件不存在")

			logs, _ := git2ftp.Cmd("git", "--no-pager", `--git-dir=`+currentTempGitDirConf, `--work-tree=`+currentTempGitDir, "log", `--pretty=format:%H|%s`, "-30")
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

					var hash bytes.Buffer
					hash.WriteString(hashs)
					client.Store(item.Path+"/git2ftp.hash", &hash)
				}
			}
		}

		//获取线上版本号
		onlineHash, err := git2ftp.FtpRead(client, item.Path+"/git2ftp.hash")
		if err != nil {
			log.Println(err.Error())
		}
		log.Println("线上版本号:", onlineHash)

		//查看差异文件
		diffFile, err := git2ftp.Cmd("git", "--no-pager", `--git-dir=`+currentTempGitDirConf, `--work-tree=`+currentTempGitDir, "diff", "--name-only", onlineHash)
		if err != nil {
			log.Println(err.Error())
		}
		log.Println("差异文件:", diffFile)

		client.Close()
	}

	log.Println("success")
}
