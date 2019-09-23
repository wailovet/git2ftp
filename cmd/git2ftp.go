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

	nowHash, err := git2ftp.GitNowHash(currentTempGitDir)
	if err != nil || nowHash == "" {
		log.Fatalln("[当前版本hash获取失败:" + err.Error() + "]")
	}

	allResult := ""

	for e := range git2ftpConf.Ftp {
		item := git2ftpConf.Ftp[e]
		client, err := goftp.DialConfig(goftp.Config{
			User:     item.User,
			Password: item.Password,
		}, item.Host)
		if err != nil {
			log.Println("DialConfig:", err.Error())
			continue
		}

		_, err = client.ReadDir(item.Path)
		if err != nil {
			fmt.Println(item.Host, "-", "[FTP连接失败]", err.Error())
			allResult += fmt.Sprintln(item.Host, "-", "[FTP连接失败]", err.Error())
			continue
		}

		//获取线上版本号
		fmt.Println(item.Host, "-", "[获取线上版本号]......")
		onlineHash := git2ftp.GetHashByFtp(client, item, currentTempGitDir)
		if onlineHash == "" {
			fmt.Println(item.Host, "-", "[获取线上版本号失败]")
			allResult += fmt.Sprintln(item.Host, "-", "[获取线上版本号失败]")
			continue
		}
		if onlineHash == nowHash {
			fmt.Println(item.Host, "-", "[版本一致,跳过部署]")
			allResult += fmt.Sprintln(item.Host, "-", "[版本一致,跳过部署]")
			continue
		}

		fmt.Println(item.Host, "-", "[线上版本号]:", onlineHash, "[开始部署]")
		errNum := -1

		diffFiles, err := git2ftp.GetDiffFiles(currentTempGitDir, onlineHash)
		if err != nil {
			log.Println("GetDiffFiles:", err.Error())
			continue
		}

		for k := range diffFiles {
			if errNum < 0 {
				errNum = 0
			}

			localPath, _ := filepath.Abs(currentTempGitDir + "/" + diffFiles[k])

			isEnd := false
			fmt.Print(item.Host, " - ", "[", diffFiles[k], "]开始上传")
			go func() {
				for !isEnd {
					fmt.Print(".")
					time.Sleep(time.Second)
				}
			}()
			err := git2ftp.FtpWriteByFile(client, localPath, git2ftp.FtpAbs(item.Path+"/"+diffFiles[k]))
			isEnd = true
			if err != nil {
				errNum++
				fmt.Println("上传失败:", err.Error())
			} else {
				fmt.Println("上传完成")
			}
		}
		if errNum == -1 {
			fmt.Println(item.Host, "-", "[未启动部署,检测不到修改文件]")
			allResult += fmt.Sprintln(item.Host, "-", "[未启动部署,检测不到修改文件]")
		} else {
			if errNum > 0 {
				fmt.Println(item.Host, "-", "[部署失败][失败次数", errNum, "次]")
				allResult += fmt.Sprintln(item.Host, "-", "[部署失败][失败次数", errNum, "次]")
			} else {

				err = git2ftp.FtpWrite(client, git2ftp.FtpAbs(item.Path+"/git2ftp.hash"), []byte(nowHash))
				if err != nil {
					fmt.Println(item.Host, "-", "[部署失败][当前版本hash远程写入失败:"+err.Error()+"]")
					allResult += fmt.Sprintln(item.Host, "-", "[部署失败][当前版本hash远程写入失败:"+err.Error()+"]")
				} else {
					fmt.Println(item.Host, "-", "[部署成功]")
					allResult += fmt.Sprintln(item.Host, "-", "[部署成功]")
				}

			}
		}

		client.Close()
	}

	fmt.Println("部署结果汇总:")
	fmt.Println(allResult)
}
