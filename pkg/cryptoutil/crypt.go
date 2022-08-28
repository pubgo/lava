package cryptoutil

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"

	"github.com/pubgo/funk/logx"
)

var SK = []byte("696D897C9AA0611B")

func AESEncrypt(str string) string {
	buf := &bytes.Buffer{}
	buf.Grow(4096)
	_, err := hex.NewEncoder(buf).Write([]byte(str))
	if nil != err {
		logx.Error(err, "encrypt failed")
		return ""
	}
	data := buf.Bytes()
	block, err := aes.NewCipher(SK)
	if nil != err {
		logx.Error(err, "encrypt failed")
		return ""
	}
	cbc := cipher.NewCBCEncrypter(block, []byte("RandomInitVector"))
	content := data
	content = pkcs5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	cbc.CryptBlocks(crypted, content)
	return hex.EncodeToString(crypted)
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func AESDecrypt(cryptStr string) []byte {
	crypt, err := hex.DecodeString(cryptStr)
	if nil != err {
		logx.Error(err, "decrypt failed")
		return nil
	}

	block, err := aes.NewCipher(SK)
	if nil != err {
		return nil
	}
	cbc := cipher.NewCBCDecrypter(block, []byte("RandomInitVector"))
	decrypted := make([]byte, len(crypt))
	cbc.CryptBlocks(decrypted, crypt)
	return pkcs5Trimming(decrypted)
}

func pkcs5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}
