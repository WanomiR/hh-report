package storage

import (
	"app/internal/lib/e"
	"crypto/sha1"
	"fmt"
	"io"
	"strconv"
)

type File struct {
	ChatID int
	Query  string
}

func NewFile(chatID int, query string) *File {
	return &File{chatID, query}
}

func (f *File) Hash() (hash string, err error) {
	defer func() { err = e.WrapIfErr("couldn't calculate hash", err) }()
	h := sha1.New()

	if _, err = io.WriteString(h, strconv.Itoa(f.ChatID)); err != nil {
		return "", err
	}

	if _, err = io.WriteString(h, f.Query); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
