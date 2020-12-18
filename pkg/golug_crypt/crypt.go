package crypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/pubgo/xlog"
	"math/rand"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

//en
func AesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = PKCS5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

//de
func AesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	err, origData = PKCS5UnPadding(origData)
	return origData, err
}

//Completion when the length is insufficient
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//Remove excess
func PKCS5UnPadding(origData []byte) (error, []byte) {
	length := len(origData)
	unpadding := int(origData[length-1])
	if (length - unpadding) < 0 {
		return errors.New("len error"), nil
	}
	return nil, origData[:(length - unpadding)]
}

//Generate 32-bit MD5 strings
func Md5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//Generating Random Verification Key
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// New generates a new key based on a passphrase and salt
func New(passphrase []byte, usersalt []byte) (key []byte, salt []byte, err error) {
	if len(passphrase) < 1 {
		err = fmt.Errorf("need more than that for passphrase")
		return
	}
	if usersalt == nil {
		salt = make([]byte, 8)
		// http://www.ietf.org/rfc/rfc2898.txt
		// Salt.
		if _, err := rand.Read(salt); err != nil {
			xlog.Fatalf("can't get random salt: %v", err)
		}
	} else {
		salt = usersalt
	}
	key = pbkdf2.Key(passphrase, salt, 100, 32, sha256.New)
	return
}

// Encrypt will encrypt using the pre-generated key
func Encrypt(plaintext []byte, key []byte) (encrypted []byte, err error) {
	// generate a random iv each time
	// http://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf
	// Section 8.2
	ivBytes := make([]byte, 12)
	if _, err := rand.Read(ivBytes); err != nil {
		xlog.Fatalf("can't initialize crypto: %v", err)
	}
	b, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	aesgcm, err := cipher.NewGCM(b)
	if err != nil {
		return
	}
	encrypted = aesgcm.Seal(nil, ivBytes, plaintext, nil)
	encrypted = append(ivBytes, encrypted...)
	return
}

// Decrypt using the pre-generated key
func Decrypt(encrypted []byte, key []byte) (plaintext []byte, err error) {
	if len(encrypted) < 13 {
		err = fmt.Errorf("incorrect passphrase")
		return
	}
	b, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	aesgcm, err := cipher.NewGCM(b)
	if err != nil {
		return
	}
	plaintext, err = aesgcm.Open(nil, encrypted[:12], encrypted[12:], nil)
	return
}
