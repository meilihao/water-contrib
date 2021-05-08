package render

import (
	"bytes"
	"errors"
	"io/ioutil"
)

var (
	ErrEmptyTmpl = errors.New("EmptyTemplate")
)

func getFileContent(fp string) (string, error) {
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		return "", err
	}

	s := string(bytes.TrimSpace(data))
	if len(s) < 1 {
		return "", ErrEmptyTmpl
	}

	return s, nil
}
