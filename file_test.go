package zxgo

import (
	"testing"
)

func test_AppendAllLines(t *testing.T) {
	filename := "file_test.txt"
	contents := make([]string, 0)
	contents = append(contents, "行A数据")
	contents = append(contents, "行B数据")
	AppendAllLines(filename, contents)
}
