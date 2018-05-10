package file

/*
go get -u -v golang.org/x/text/transform
*/
import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func AppendLine(path string, content string, panicWhenError bool) error {
	contents := []string{content}
	return AppendAllLines(path, contents, panicWhenError)
}

// 模仿了C#的[System.IO.File.AppendAllLines("path", new string[] { })]函数的行为.
// 参考了ioutil.WriteFile("path", nil, os.ModeAppend)函数.
func AppendAllLines(path string, contents []string, panicWhenError bool) error {

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		if panicWhenError {
			panic(err)
		}
		return err
	}

	for _, content := range contents {

		data := []byte(content + "\n")

		if n, err1 := file.Write(data); err1 != nil {
			err = err1
			break
		} else if n < len(data) {
			err = io.ErrShortWrite
			break
		}
	}

	if err1 := file.Close(); err == nil {
		err = err1
	}

	if err != nil && panicWhenError {
		panic(err)
	}
	return err
}

func AppendAllLines_bak(path string, contents []string, encodingType string) error {
	/* 我建议使用golang.org/x/text/encoding的标准软件包，可能还会使用golang.org/x/net/charset */
	calcEncoding := func(eType string) encoding.Encoding {
		if eType == "GBK" {
			return simplifiedchinese.GBK
		} else if eType == "GB18030" {
			return simplifiedchinese.GB18030
		} else {
			return nil
		}
	}

	eInterface := calcEncoding(encodingType)
	if eInterface == nil {
		return errors.New(fmt.Sprintf("Unknown encodingType=%v", encodingType))
	}

	allData := make([][]byte, 0)
	for _, content := range contents {
		content += "\n"
		if data, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(content)), eInterface.NewDecoder())); err != nil {
			return errors.New(fmt.Sprintf("convert encoding fail."))
		} else {
			allData = append(allData, data)
		}
	}

	for _, data := range allData {
		ioutil.WriteFile(path, data, os.ModeAppend)
	}

	return nil
}

type iterator_reader struct { //内部使用的类,不能被外部创建.
	f    *os.File
	r    *bufio.Reader
	last bool
}

func (self *iterator_reader) Next() (line string, err error, first bool, last bool) {
	if self.last {
		err = io.EOF
		return
	}

	first = false
	line, err = self.r.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			last = true
			self.last = last
			err = nil
		}
	} else {
		last = false
	}

	return
}

func (self *iterator_reader) init(filename string) (line string, err error, first bool, last bool) {
	self.f, err = os.Open(filename)
	if err == nil {
		self.r = bufio.NewReader(self.f)
		line, err, first, last = self.Next()
		first = true
	}
	return
}

// 函数用法如下所示:
//	var eOut string
//	for line, err, first, last, iter := file.ReadLine("D:/_a.txt", &eOut); err == nil; line, err, first, last = iter.Next() {
//		line = strings.TrimRight(line, "\r\n")
//		fmt.Println(first, last, line)
//	}
//	if 0 < len(eOut) {
//		fmt.Println(eOut)
//	}
func ReadLine(filename string, eOut *string) (line string, err error, isFirst bool, isLast bool, iter *iterator_reader) {
	iter = &iterator_reader{nil, nil, false}
	line, err, isFirst, isLast = iter.init(filename)
	if eOut != nil {
		if err != nil {
			*eOut = err.Error()
		} else {
			*eOut = ""
		}
	}
	return
}
