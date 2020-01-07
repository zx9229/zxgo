//golang 上传文件(Web版)
//https://www.yuque.com/docs/share/f46698d3-5847-4e44-a60e-67be999bbd56

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

var homedir string

func main() {
	var (
		argHelp bool
		argHost string
		argPort int
	)
	flag.BoolVar(&argHelp, "help", false, "[M] show this help.")
	flag.IntVar(&argPort, "port", 9999, "[M] port")
	flag.StringVar(&argHost, "host", "localhost", "[M] host")
	flag.StringVar(&homedir, "homedir", ".", "[M] home directory")
	flag.Parse()

	for range "1" {
		if argHelp {
			flag.Usage()
			break
		}
		if argPort <= 0 || 65535 < argPort {
			log.Printf("illegal port (%v)", argPort)
			break
		}

		log.Printf("homedir: [%v]", homedir)
		addr := fmt.Sprintf("%s:%d", argHost, argPort)

		http.Handle("/", http.FileServer(http.Dir(homedir)))
		http.HandleFunc("/upload", pageUpload)
		http.HandleFunc("/uploadProcess", funcUpload)

		log.Printf("http://%s, ListenAndServe ...", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Printf("ListenAndServe, %v", err)
			break
		}
	}
}

func pageUpload(writer http.ResponseWriter, request *http.Request) {
	htmlContent := `
<html>
<head>
<title>上传文件</title>
</head>
<body>
<form enctype="multipart/form-data" action="/uploadProcess" method="post">
<table border="1">
<tr>
<td>要上传到哪个目录</td>
<td><input type="text" name="dirname" size="80"></td>
</tr>
<tr>
<td>要上传的本地文件</td>
<td><input type="file" name="filename"></td>
</tr>
<tr>
<td>文件名附加时间戳</td>
<td><input type="checkbox" name="timestamp"></td>
</tr>
<tr>
<td></td>
<td><input type="hidden" name="token" value="{...{.}...}"></td>
</tr>
<tr>
<td>上传按钮</td>
<td><input type="submit" value="upload"></td>
</tr>
</table>
</form>
<a href="/">      <input type="button" value='切换到下载页面'></a><br>
<a href="/upload"><input type="button" value='切换到上传页面'></a>
</body>
</html>`
	writer.Write([]byte(htmlContent))
}

func funcUpload(writer http.ResponseWriter, request *http.Request) {
	var message string

	for range "1" {
		var err error

		if request.Method != "POST" {
			err = errors.New("request Method is not POST")
			message += fmt.Sprintf("<br>error_message: [%v]\n", err)
			log.Println(err)
			break
		}

		//http://docscn.studygolang.com/pkg/net/http/#Request.ParseMultipartForm
		request.ParseMultipartForm(32 << 20) //32MB

		dirnameValue := request.FormValue("dirname")
		message += fmt.Sprintf("<br>dirname: [%v]\n", dirnameValue)
		if !filepath.IsAbs(dirnameValue) {
			if dirnameValue, err = filepath.Abs(filepath.Join(homedir, dirnameValue)); err != nil {
				message += fmt.Sprintf("<br>error_message: [%v]\n", err)
				log.Println(err)
				break
			}
		}
		if isExist, isDir, err := PathIsExist(dirnameValue); err != nil {
			message += fmt.Sprintf("<br>error_message: [%v]\n", err)
			log.Println(err)
			break
		} else if isExist && !isDir {
			err = errors.New("the path exists but not the directory")
			message += fmt.Sprintf("<br>error_message: [%v]\n", err)
			log.Println(err)
			break
		} else if !isExist {
			if err = os.MkdirAll(dirnameValue, 0777); err != nil {
				message += fmt.Sprintf("<br>error_message: [%v]\n", err)
				log.Println(err)
				break
			}
		}

		multipartFile, multipartFileHeader, err := request.FormFile("filename")
		if err != nil {
			message += fmt.Sprintf("<br>error_message: [%v]\n", err)
			log.Println(err)
			break
		}
		defer multipartFile.Close()

		message += fmt.Sprintf("<br>filename: [%v]\n", multipartFileHeader.Filename)
		svrFilename := multipartFileHeader.Filename
		filenameWithTimestamp := strings.ToLower(request.FormValue("timestamp")) == "on" //"checkbox"(复选框)被选中为"on"
		message += fmt.Sprintf("<br>timestamp: [%v]\n", request.FormValue("timestamp"))
		if filenameWithTimestamp {
			ext := path.Ext(svrFilename) //获取文件后缀
			svrFilename = svrFilename + "." + string(time.Now().Format("20060102_150405")) + ext
			svrFilename = filepath.Join(dirnameValue, svrFilename)
		}

		f, err := os.OpenFile(svrFilename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			message += fmt.Sprintf("<br>error_message: [%v]\n", err)
			log.Println(err)
			break
		}
		defer f.Close()

		io.Copy(f, multipartFile)

		log.Println(svrFilename)
		message += fmt.Sprintf("<br>error_message: [%v]\n", "SUCCESS")
	}

	htmlContent := `
<html>
<head>
<title>上传文件处理结果</title>
</head>
<body>
<a href="/upload"><input type="button" value='切换到上传页面'></a>
` + message + `
</body>
</html>`
	writer.Write([]byte(htmlContent))
}

//PathIsExist 只有"无法判断"文件是否存在时,err才非nil
func PathIsExist(path string) (isExist bool, isDir bool, err error) {
	fileinfo, err := os.Stat(path)
	if fileinfo != nil {
		isDir = fileinfo.IsDir()
	}
	isExist = (err == nil) || os.IsExist(err)
	if os.IsExist(err) || os.IsNotExist(err) {
		err = nil
	}
	return
}
