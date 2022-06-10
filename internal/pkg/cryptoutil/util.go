package cryptoutil

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

func Md5(content string) (md string) {
	h := md5.New()
	_, _ = io.WriteString(h, content)
	md = fmt.Sprintf("%x", h.Sum(nil))
	return
}

func SHA256Sum(data interface{}) string {
	h := sha256.New()
	if v, ok := data.([]byte); ok {
		h.Write(v)
	} else if v, ok := data.(string); ok {
		h.Write([]byte(v))
	} else {
		return ""
	}

	return hex.EncodeToString(h.Sum(nil))
}
