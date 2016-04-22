package render

import (
	"errors"
	"io/ioutil"
)

func getFileContent(fPath string) (string, error) {
	data, err := ioutil.ReadFile(fPath)
	if err != nil {
		return "", err
	}

	s := string(data)
	if len(s) < 1 {
		return "", errors.New("EmptyTemplate")
	}

	return s, nil
}
