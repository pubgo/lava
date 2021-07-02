package restc

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	fieldFormat = "--%s\r\nContent-Disposition: form-data; name=\"%s\"\r\n\r\n%s\r\n"
	fileHeader  = "Content-type: application/octet-stream"
	fileFormat  = "--%s\r\nContent-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n%s\r\n\r\n"
)

type Uploader struct {
	c        Client
	boundary string
	params   map[string][]string
	files    []struct {
		name     string
		formname string
		content  io.Reader
	}
}

func NewUploader(c Client) *Uploader {
	return &Uploader{
		c:        c,
		boundary: "myboundary",
		params:   make(map[string][]string),
	}
}

func (u *Uploader) buildBodyTop() string {
	var parts = make([]string, 0, len(u.params))
	for k, v := range u.params {
		for _, v1 := range v {
			parts = append(parts, fmt.Sprintf(fieldFormat, u.boundary, k, v1))
		}
	}

	return strings.Join(parts, "")
}

func (u *Uploader) SetBoundary(boundary string) *Uploader {
	u.boundary = boundary
	return u
}

func (u *Uploader) SetParams(params map[string][]string) *Uploader {
	u.params = params
	return u
}

func (u *Uploader) AddParams(params map[string][]string) *Uploader {
	for k, v := range params {
		u.params[k] = v
	}
	return u
}

func (u *Uploader) AddParam(name, value string) *Uploader {
	if _, ok := u.params[name]; ok {
		u.params[name] = append(u.params[name], value)
	} else {
		u.params[name] = []string{value}
	}
	return u
}

func (u *Uploader) AddFile(formname, filename string, fileReader io.Reader) *Uploader {
	u.files = append(u.files, struct {
		name     string
		formname string
		content  io.Reader
	}{name: filename, formname: formname, content: fileReader})
	return u
}

func (u *Uploader) ContentType() string {
	return fmt.Sprintf("multipart/form-data; boundary=%s", u.boundary)
}

func (u *Uploader) Body() (io.Reader, error) {
	var rds = []io.Reader{
		strings.NewReader(u.buildBodyTop()),
	}

	for _, file := range u.files {
		var bs = make([]byte, 1024)
		size, err := file.content.Read(bs)
		if err != nil {
			return nil, err
		}

		newRd := io.MultiReader(bytes.NewReader(bs[:size]), file.content)
		contentType := http.DetectContentType(bs)
		if contentType == "" {
			contentType = fileHeader
		} else {
			contentType = "Content-type: " + contentType
		}

		rds = append(rds, strings.NewReader(fmt.Sprintf(fileFormat, u.boundary, file.formname, file.name, contentType)))
		rds = append(rds, newRd)
		rds = append(rds, strings.NewReader("\r\n"))
	}
	rds = append(rds, strings.NewReader(fmt.Sprintf("--%s--\r\n", u.boundary)))
	return io.MultiReader(rds...), nil
}
