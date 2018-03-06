package zxgo

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

func CalcHash(filename, style string, toUpper bool) (hexStr string, err error) {

	style = strings.ToLower(style)

	var hs hash.Hash = nil
	switch style {
	case "md5":
		hs = md5.New()
	case "sha1":
		hs = sha1.New()
	case "sha256":
		hs = sha256.New()
	case "sha512":
		hs = sha512.New()
	default:
		err = errors.New(fmt.Sprintf("Unknown style=%v", style))
		return
	}

	if file, err1 := os.OpenFile(filename, os.O_RDONLY, 0); err1 != nil {
		err = err1
		return
	} else {
		io.Copy(hs, file)
		file.Close()
	}

	hexStr = hex.EncodeToString(hs.Sum(nil))
	if toUpper {
		hexStr = strings.ToUpper(hexStr)
	}

	return
}
