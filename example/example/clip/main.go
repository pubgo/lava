package main

import (
	"golang.design/x/clipboard"

	"os"
)

func main() {
	var b []byte

	var img bool
	b = clipboard.Read(clipboard.FmtText)
	if b == nil {
		img = true
		b = clipboard.Read(clipboard.FmtImage)
	}

	if img {
		os.WriteFile("test.png", b, os.ModePerm)
	} else {
		os.WriteFile("test.txt", b, os.ModePerm)
	}
}
