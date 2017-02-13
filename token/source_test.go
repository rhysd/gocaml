package token

import (
	"testing"
)

func TestReadFromFile(t *testing.T) {
	s, err := NewSourceFromFile("./source.go")
	if err != nil {
		t.Fatal(err)
	}

	if s.Name != "./source.go" {
		t.Errorf("Unexpected file name %s", s.Name)
	}

	if s.Code == nil {
		t.Errorf("Code was not read properly")
	}

	if !s.Exists {
		t.Errorf("File must exist")
	}
}

func TestUnexistFile(t *testing.T) {
	_, err := NewSourceFromFile("./__unknown_file.ml")
	if err == nil {
		t.Fatalf("Unknown error must cause an error")
	}
}

func TestReadFromStdin(t *testing.T) {
	s, err := NewSourceFromStdin()
	if err != nil {
		t.Fatal(err)
	}

	if s.Name != "<stdin>" {
		t.Errorf("Unexpected file name %s", s.Name)
	}

	if s.Code == nil {
		t.Errorf("Code was not read properly")
	}

	if s.Exists {
		t.Errorf("File must not exist")
	}
}
