package source

import (
	"io/ioutil"
	"os"
)

type Source struct {
	Name string
	Code []byte
}

func FromFile(name string) (*Source, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return &Source{name, b}, nil
}

func FromStdin() (*Source, error) {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, err
	}
	return &Source{"<stdin>", b}, nil
}
