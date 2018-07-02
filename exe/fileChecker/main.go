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

type commonConfigData struct {
	fmt     string
	name    string
	root    string
	match   string
	glob    string
	pattern *regexp.Regexp
	depth   int
}

type flagConfigData struct {
	helpPtr   *bool
	fmtPtr    *string
	namePtr   *string
	rootPtr   *string
	matchPtr  *string
	globPtr   *string
	regexpPtr *string
	depthPtr  *int
}

func (thls *flagConfigData) toCommon() (cfg *commonConfigData, err error) {
	cfg = new(commonConfigData)
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

var g_cfg *commonConfigData = nil

func main() {
	flagCfg := flagConfigData{}

	flagCfg.helpPtr = flag.Bool("help", false, "show this help")
	flagCfg.fmtPtr = flag.String("fmt", "<RELNAME>, <MD5>, <SIZE>", "combine with <MD5>,<SIZE>,<MTIME>,<NAME>,<RELNAME>,<ABSNAME>")
	flagCfg.namePtr = flag.String("name", "", "set file name")
	flagCfg.rootPtr = flag.String("root", ".", "set root path")
	flagCfg.matchPtr = flag.String("match", "NAME", "one of NAME,RELNAME,ABSNAME")
	flagCfg.globPtr = flag.String("glob", "", "match with glob")
	flagCfg.regexpPtr = flag.String("regexp", "", "match with regexp")
	flagCfg.depthPtr = flag.Int("depth", 0, "set path maximum depth")
	//所有标志都声明完成以后，调用 flag.Parse() 来执行命令行解析。
	flag.Parse()

	if *flagCfg.helpPtr {
		flag.Usage()
		return
	}

	var err error
	if g_cfg, err = flagCfg.toCommon(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(100)
		return
	}

	if g_cfg.name != EmptyStr {
		var absName, rootDir string
		if absName, err = filepath.Abs(g_cfg.name); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(100)
			return
		}
		rootDir = filepath.Dir(absName)
		var info os.FileInfo
		if info, err = os.Lstat(g_cfg.name); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(100)
			return
		}
		var fmttedData string
		if fmttedData, err = _formatData(rootDir, absName, info, g_cfg.fmt); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(100)
			return
		}
		fmt.Println(fmt.Sprintf("RootPath=%v", rootDir))
		fmt.Println(fmttedData)
		return
	}

	//os.Getwd() [Getwd返回与当前目录对应的根路径名].
	fmt.Println(fmt.Sprintf("RootPath=%v", g_cfg.root))
	if err = filepath.Walk(g_cfg.root, _WalkCallbackFunc); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("ERROR, filepath.Walk FAIL, err=%v", err))
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

	if info.IsDir() && 0 < g_cfg.depth {
		relativePath := strings.SplitN(path, g_cfg.root, 2)[1]
		curLevel := len(strings.Split(relativePath, string(os.PathSeparator))) - 1
		if g_cfg.depth < curLevel {
			return filepath.SkipDir
		}
	}

	if info.IsDir() {
		return err
	}

	if !isMatch(g_cfg.root, path, info, g_cfg.glob, g_cfg.pattern, g_cfg.match) {
		return nil
	}

	fmtData, err := _formatData(g_cfg.root, path, info, g_cfg.fmt)
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
