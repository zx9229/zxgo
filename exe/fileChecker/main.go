package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/zx9229/zxgo"
)

type CommonConfigData struct {
	fmt     string
	name    string
	root    string
	match   string
	glob    string
	pattern *regexp.Regexp
	depth   int
}

type FlagConfigData struct {
	helpPtr   *bool
	fmtPtr    *string
	namePtr   *string
	rootPtr   *string
	matchPtr  *string //NAME,RELNAME,ABSNAME
	globPtr   *string
	regexpPtr *string
	depthPtr  *int
}

func (thls *FlagConfigData) toCommon() (cfg *CommonConfigData, err error) {
	cfg = new(CommonConfigData)
	for range "1" {
		if len(*thls.fmtPtr) <= 0 {
			err = errors.New("format an empty string")
			break
		}
		cfg.fmt = *thls.fmtPtr

		if *thls.namePtr != EmptyStr {
			cfg.name = *thls.namePtr
			cfg.root = EmptyStr
		} else if *thls.rootPtr != EmptyStr {
			if cfg.root, err = filepath.Abs(*thls.rootPtr); err != nil {
				break
			}
			cfg.name = EmptyStr
		} else {
			err = errors.New("name and root are empty")
			break
		}

		if *thls.matchPtr != NAME && *thls.matchPtr != RELNAME && *thls.matchPtr != ABSNAME {
			err = errors.New("unknown match type")
			break
		}
		cfg.match = *thls.matchPtr

		if *thls.globPtr != EmptyStr {
			if _, err2 := filepath.Match(*thls.globPtr, ""); err2 != nil {
				err = errors.New("syntax error in glob")
				break
			}
			cfg.glob = *thls.globPtr
			cfg.pattern = nil
		} else if *thls.regexpPtr != EmptyStr {
			if cfg.pattern, err = regexp.Compile(*thls.regexpPtr); err != nil {
				err = errors.New("syntax error in regexp")
				break
			}
			cfg.glob = EmptyStr
		} else {
			cfg.glob = EmptyStr
			cfg.pattern = nil
		}
	}

	if err != nil {
		cfg = nil
	}

	return
}

const (
	NAME     string = "NAME"    //文件的basename
	RELNAME  string = "RELNAME" //文件相对root的相对目录
	ABSNAME  string = "ABSNAME" //文件的绝对目录
	EmptyStr string = ""
)

var GlobalCfg *CommonConfigData = nil

func main() {
	flagConfig := FlagConfigData{}

	flagConfig.helpPtr = flag.Bool("help", false, "show this help")
	flagConfig.fmtPtr = flag.String("fmt", "<RELNAME>, <MD5>, <SIZE>", "combine with <MD5>,<SIZE>,<MTIME>,<NAME>,<RELNAME>,<ABSNAME>")
	flagConfig.namePtr = flag.String("name", "", "set file name")
	flagConfig.rootPtr = flag.String("root", ".", "set root path")
	flagConfig.matchPtr = flag.String("match", "NAME", "one of NAME,RELNAME,ABSNAME")
	flagConfig.globPtr = flag.String("glob", "", "match with glob")
	flagConfig.regexpPtr = flag.String("regexp", "", "match with regexp")
	flagConfig.depthPtr = flag.Int("depth", 0, "set path maximum depth")
	//所有标志都声明完成以后，调用 flag.Parse() 来执行命令行解析。
	flag.Parse()

	if *flagConfig.helpPtr {
		flag.Usage()
		return
	}

	if cfg, err := flagConfig.toCommon(); err == nil {
		GlobalCfg = cfg
	} else {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(100)
		return
	}

	if GlobalCfg.name != EmptyStr {
		info, err := os.Lstat(GlobalCfg.name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(100)
			return
		}
		tmpAbsName, _ := filepath.Abs(GlobalCfg.name)
		tmpRootStr := filepath.Dir(tmpAbsName)
		fmtData, err := _formatData(tmpRootStr, tmpAbsName, info, GlobalCfg.fmt)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(100)
			return
		}
		fmt.Println(fmt.Sprintf("RootPath=%v", tmpRootStr))
		fmt.Println(fmtData)
		return
	}

	//os.Getwd() [Getwd返回与当前目录对应的根路径名].
	fmt.Println(fmt.Sprintf("RootPath=%v", GlobalCfg.root))
	if err := filepath.Walk(GlobalCfg.root, _WalkCallbackFunc); err != nil {
		fmt.Println(fmt.Printf("ERROR, filepath.Walk FAIL, err=%v", err))
		os.Exit(100)
		return
	}
}

func _formatData(rootDir string, absName string, info os.FileInfo, fmtData string) (data string, err error) {
	var md5Data string
	if info.IsDir() {
		err = errors.New("is not file")
		return
	}
	if strings.Contains(fmtData, "<MD5>") {
		if md5Data, err = zxgo.CalcHash(absName, "md5", true); err != nil {
			return
		}
	}
	if md5Data != EmptyStr {
		fmtData = strings.Replace(fmtData, "<MD5>", md5Data, -1)
	}
	if strings.Contains(fmtData, "<SIZE>") {
		fmtData = strings.Replace(fmtData, "<SIZE>", strconv.FormatInt(info.Size(), 10), -1)
	}
	if strings.Contains(fmtData, "<MTIME>") {
		fmtData = strings.Replace(fmtData, "<MTIME>", info.ModTime().Format("2006-01-02 15:04:05"), -1)
	}
	if strings.Contains(fmtData, "<NAME>") {
		fmtData = strings.Replace(fmtData, "<NAME>", info.Name(), -1)
	}
	if strings.Contains(fmtData, "<RELNAME>") {
		if rootDir != EmptyStr {
			relName := calcRelName(rootDir, absName)
			fmtData = strings.Replace(fmtData, "<RELNAME>", relName, -1)
		}
	}
	if strings.Contains(fmtData, "<ABSNAME>") {
		fmtData = strings.Replace(fmtData, "<ABSNAME>", absName, -1)
	}
	data = fmtData
	return
}

func guessRootPath(relName string) string {
	var err error
	var absPath string
	if absPath, err = filepath.Abs(relName); err != nil {
		panic(err)
	}
	fields := strings.Split(absPath, string(os.PathSeparator))
	if os.PathSeparator == '/' {
		if fields[0] != `/` {
			panic("logic error")
		}
		return fields[0]
	} else if os.PathSeparator == '\\' {
		matched, err := filepath.Match("[a-zA-Z]:", fields[0])
		if err != nil || !matched {
			panic("logic error")
		}
		return fields[0]
	} else {
		panic("logic error")
	}
}

func calcRelName(rootDir string, absName string) string {
	//这里假定(rootDir是AbsName的父代目录)
	headData := "ROOTPATH"
	if strings.HasSuffix(rootDir, string(os.PathSeparator)) {
		headData += string(os.PathSeparator)
	}
	return fmt.Sprintf("%v%v", headData, strings.SplitN(absName, rootDir, 2)[1])
}

func isMatch(rootDir, absName string, info os.FileInfo, glob string, pattern *regexp.Regexp, matchType string) bool {

	if glob == EmptyStr && pattern == nil {
		return true
	}

	var matched bool

	var err error
	doMatchOperation := func(someName string) {
		if pattern != nil {
			matched = pattern.MatchString(someName)
		} else {
			if matched, err = filepath.Match(glob, someName); err != nil {
				panic(err)
			}
		}
	}

	switch matchType {
	case NAME:
		doMatchOperation(info.Name())
	case RELNAME:
		relName := calcRelName(rootDir, absName)
		doMatchOperation(relName)
	case ABSNAME:
		doMatchOperation(absName)
	default:
		panic(fmt.Sprintf("unknown match type: %v", matchType))
	}

	return matched
}

func _WalkCallbackFunc(path string, info os.FileInfo, errIn error) error {
	var err error
	if errIn != nil {
		err = errIn
		fmt.Println(fmt.Sprintf("[ERROR] %v, err=%v", path, err))
		return err
	}

	if info.IsDir() && 0 < GlobalCfg.depth {
		relativePath := strings.SplitN(path, GlobalCfg.root, 2)[1]
		curLevel := len(strings.Split(relativePath, string(os.PathSeparator))) - 1
		if GlobalCfg.depth < curLevel {
			return filepath.SkipDir
		}
	}

	if info.IsDir() {
		return err
	}

	if !isMatch(GlobalCfg.root, path, info, GlobalCfg.glob, GlobalCfg.pattern, GlobalCfg.match) {
		return nil
	}

	fmtData, err := _formatData(GlobalCfg.root, path, info, GlobalCfg.fmt)
	if err != nil {
		fmt.Println(fmt.Sprintf("[ERROR] %v, err=%v", path, err))
		err = nil
		return err
	}
	fmt.Println(fmtData)

	return nil
}

/*
-match="ABSNAME,RELNAME,NAME"
-pattern
-glob
-root
-name
因为文件名不能含有"<>"所以可以将它们选择为打印项
哈希值的大小写

-f "<MD5>,<NAME>,<SIZE>,<MTIME>,<ABSNAME>,<RELNAME>"
<SIZE>: 文件的大小
<MTIME>: 文件的修改时间
<ATIME>: 文件的访问时间
<NAME>: 文件的名字
<ABSNAME>: 绝对路径名字
<RELNAME>: 相对路径名字
<Pn>: <NAME>的第几个父目录,<P1>文件的父目录,<P2>文件的爷目录,以此类推.
*/
