package token

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Source struct {
	Name   string
	Code   []byte
	Exists bool
}

func NewSourceFromFile(name string) (*Source, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return &Source{name, b, true}, nil
}

func NewSourceFromStdin() (*Source, error) {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	return &Source{"<stdin>", b, false}, nil
}

func NewDummySource(code string) *Source {
	return &Source{"dummy", []byte(code), false}
}

func (src *Source) BaseName() string {
	if !src.Exists {
		return "out"
	}
	b := filepath.Base(src.Name)
	return strings.TrimSuffix(b, filepath.Ext(b))
}
