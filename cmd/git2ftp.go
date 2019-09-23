package main

import (
	"fmt"
	"github.com/otiai10/copy"
	"github.com/secsy/goftp"
	"github.com/wailovet/git2ftp"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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

		//获取线上版本号
		onlineHash := git2ftp.GetHashByFtp(client, item, currentTempGitDir)

		log.Println("线上版本号:", onlineHash)

		diffFiles := git2ftp.GetDiffFiles(currentTempGitDir, onlineHash)

		for k := range diffFiles {
			localPath, _ := filepath.Abs(currentTempGitDir + "/" + diffFiles[k])

			isEnd := false
			fmt.Print(item.Host, "[", diffFiles[k], "]开始上传")
			go func() {
				for !isEnd {
					fmt.Print(".")
					time.Sleep(time.Second)
				}
			}()
			err := git2ftp.FtpWriteByFile(client, localPath, git2ftp.FtpAbs(item.Path+"/"+diffFiles[k]))
			isEnd = true
			if err != nil {
				fmt.Println("上传失败:", err.Error())
			} else {
				fmt.Println("上传完成")
			}
		}

		client.Close()
	}

	log.Println("success")
}
