package protoc

import (
	"strings"
)

func InsertByteSlice(slice, insertion []byte, index int) []byte {
	result := make([]byte, len(slice)+len(insertion))
	at := copy(result, slice[:index])
	at += copy(result[at:], insertion)
	copy(result[at:], slice[index:])
	return result
}

func ReplaceByteSlice(slice, insertion []byte, start int, end int) []byte {
	result := make([]byte, len(slice)+len(insertion)-(end-start))
	at := copy(result, slice[:start])
	at += copy(result[at:], insertion)
	copy(result[at:], slice[end:])
	return result
}

func snakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '-')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data))
}
