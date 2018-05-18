package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zx9229/zxgo/file"
)

func parseFileContent(filename string) []*QtSqliteStruct {
	slice_ := make([]*QtSqliteStruct, 0)

	var currData *QtSqliteStruct = nil
	var currDeal bool = false
	var lineNum int = 0

	var eOut string
	for line, err, _, _, iter := file.ReadLine(filename, &eOut); err == nil; line, err, _, _ = iter.Next() {
		lineNum += 1
		line = strings.TrimRight(line, "\r\n")
		if strings.HasPrefix(line, "struct ") {
			currDeal = true
			currData = newQtSqliteStruct()
		} else if strings.HasPrefix(line, "};") {
			slice_ = append(slice_, currData)
			currData = nil
			currDeal = false
		}
		if currDeal {
			if err := currData.parseContent(line); err != nil {
				panic(fmt.Sprintf("lineNum=%v,err=%v", lineNum, err))
			}
		}
	}
	if 0 < len(eOut) {
		fmt.Println(fmt.Sprintf("[ERROR] %v", eOut))
	}

	return slice_
}

func main() {
	helpShowPtr := flag.Bool("help", false, "show this help")
	inputFilenamePtr := flag.String("input", "input.orm", "set input filename")
	outputFilenamePtr := flag.String("output", "output.cpp.txt", "set output filename")
	flag.Parse()

	if *helpShowPtr {
		flag.Usage()
		return
	}

	allStruct := parseFileContent(*inputFilenamePtr)
	if 0 < len(allStruct) {
		if file, err := os.OpenFile(*outputFilenamePtr, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666); err == nil {
			file.Close()
		}
	}

	file.AppendLine(*outputFilenamePtr, "#include <QObject>", true)
	file.AppendLine(*outputFilenamePtr, "#include <QString>", true)
	file.AppendLine(*outputFilenamePtr, "#include <QVariant>", true)
	file.AppendLine(*outputFilenamePtr, "#include <QSqlQuery>", true)
	file.AppendLine(*outputFilenamePtr, "", true)
	file.AppendLine(*outputFilenamePtr, "", true)

	for _, currStruct := range allStruct {
		strData := currStruct.generate_cxx_definition()
		file.AppendLine(*outputFilenamePtr, strData, true)
	}

	fmt.Println("DONE.", time.Now().Format("2006-01-02 15:04:05"))
}
