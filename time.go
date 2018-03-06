package zxgo

import (
	"time"
)

// layout 可以是 2006-01-02 15:04:05.000000
// pwe => panicWhenError
func Str2Time(layout, value string, pwe bool) (t time.Time, err error) {
	t, err = time.ParseInLocation(layout, value, time.Local)
	if pwe {
		panic(err)
	}
	return
}

func Time2Str(t time.Time, layout string) string {
	return t.Format(layout)
}
