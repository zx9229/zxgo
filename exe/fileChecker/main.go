package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zx9229/zxgo"
)

type ConfigData struct {
	ShowSize bool   //显示文件的大小
	ShowTime bool   //显示文件的修改时间
	ShowHash bool   //显示文件的哈希值
	HashType string //要显示的哈希类型
	RootPath string //要遍历的根目录
	Depth    int    //路径的深度
	pattern  *regexp.Regexp
}

var GlobalConfig = &ConfigData{}

func WalkCallbackFunc(path string, info os.FileInfo, err error) error {

	if err != nil {
		fmt.Println(fmt.Sprintf("ERROR, path=%v, err=%v", path, err))
		return err
	}

	if 0 < GlobalConfig.Depth {
		relativePath := strings.SplitN(path, GlobalConfig.RootPath, 2)[1]
		curLevel := len(strings.Split(relativePath, string(os.PathSeparator))) - 1
		if GlobalConfig.Depth < curLevel {
			return nil
		}
	}

	if GlobalConfig.pattern != nil {
		if GlobalConfig.pattern.MatchString(info.Name()) == false {
			return nil
		}
	}

	if info.IsDir() {
		return nil
	}

	hexStr, err := zxgo.CalcHash(path, GlobalConfig.HashType, true)
	if err != nil {
		fmt.Println(fmt.Sprintf("ERROR, CalcHash FAIL, path=%v, err=%v", path, err))
		return err
	}

	showStr := fmt.Sprintf("ROOTPATH%v", strings.SplitN(path, GlobalConfig.RootPath, 2)[1])
	if GlobalConfig.ShowHash {
		showStr += fmt.Sprintf(", %v", hexStr)
	}
	if GlobalConfig.ShowTime {
		showStr += fmt.Sprintf(", %v", info.ModTime().Format("2006-01-02 15:04:05"))
	}
	if GlobalConfig.ShowSize {
		showStr += fmt.Sprintf(", %v", info.Size())
	}

	fmt.Println(showStr)

	return nil
}

func main() {
	helpShowPtr := flag.Bool("help", false, "show this help")
	sizeShowPtr := flag.Bool("size", true, "show file size")
	timeShowPtr := flag.Bool("time", false, "show file modify time")
	hashShowPtr := flag.Bool("hash", true, "show file hash")
	hashTypePtr := flag.String("type", "md5", "set hash type")
	rootPathPtr := flag.String("root", ".", "set root path")
	patternPtr := flag.String("pattern", "", "set regexp pattern")
	depthPtr := flag.Int("depth", 0, "set path maximum depth")
	//所有标志都声明完成以后，调用 flag.Parse() 来执行命令行解析。
	flag.Parse()

	if *helpShowPtr {
		flag.Usage()
		return
	}

	GlobalConfig.ShowSize = *sizeShowPtr
	GlobalConfig.ShowTime = *timeShowPtr
	GlobalConfig.ShowHash = *hashShowPtr
	GlobalConfig.HashType = *hashTypePtr
	GlobalConfig.Depth = *depthPtr
	if absRootPath, err := filepath.Abs(*rootPathPtr); err != nil {
		fmt.Println(fmt.Sprintf("ERROR, filepath.Abs FAIL, path=%v, err=%v", *rootPathPtr, err))
		os.Exit(100)
		return
	} else {
		GlobalConfig.RootPath = absRootPath
	}

	if len(*patternPtr) > 0 {
		var err error = nil
		GlobalConfig.pattern, err = regexp.Compile(*patternPtr)
		if err != nil {
			fmt.Println(fmt.Printf("ERROR, regexp.Compile FAIL, expr=%v, err=%v", *patternPtr, err))
			os.Exit(100)
			return
		}
	} else {
		GlobalConfig.pattern = nil
	}

	//os.Getwd() [Getwd返回与当前目录对应的根路径名].
	fmt.Println(fmt.Sprintf("RootPath=%v", GlobalConfig.RootPath))
	if err := filepath.Walk(GlobalConfig.RootPath, WalkCallbackFunc); err != nil {
		fmt.Println(fmt.Printf("ERROR, filepath.Walk FAIL, err=%v", err))
		os.Exit(100)
		return
	}
}
