package storage

import (
	"app/internal/lib/e"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Storage interface {
	Save(*File) error
	Remove(*File) error
	ReadAll(int) ([]*File, error)
	IsExist(*File) (bool, error)
}

type QueriesStorage struct {
	basPath string
}

func NewQueriesStorage(basePath string) *QueriesStorage {
	return &QueriesStorage{basPath: basePath}
}

func (s *QueriesStorage) Save(f *File) (err error) {
	defer func() { err = e.WrapIfErr("couldn't save file", err) }()

	dir := filepath.Join(s.basPath, strconv.Itoa(f.ChatID))

	if err = os.MkdirAll(dir, 0774); err != nil {
		return err
	}

	fileName, err := getFileName(f)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, fileName)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	if err = gob.NewEncoder(file).Encode(f); err != nil {
		return err
	}

	log.Println("file created: ", path)

	return nil
}

func (s *QueriesStorage) Remove(f *File) (err error) {
	fileName, err := getFileName(f)
	if err != nil {
		return e.WrapIfErr("couldn't remove file", err)
	}

	path := filepath.Join(s.basPath, strconv.Itoa(f.ChatID), fileName)

	if err = os.Remove(path); err != nil {
		return e.WrapIfErr(fmt.Sprintf("couldn't remove file %s", path), err)
	}

	log.Println("file removed: ", path)

	return nil
}

func (s *QueriesStorage) ReadAll(chatId int) (files []*File, err error) {
	defer func() { err = e.WrapIfErr("couldn't read files", err) }()

	files = make([]*File, 0)
	dir := filepath.Join(s.basPath, strconv.Itoa(chatId))

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		var file *File
		if !entry.IsDir() {
			path := filepath.Join(dir, entry.Name())

			if file, err = s.decodeFile(path); err != nil {
				log.Println(e.WrapIfErr(fmt.Sprintf("couldn't decode file %s", path), err))
				continue
			}

			files = append(files, file)
		}
	}

	log.Printf("read %d files for chat %d\n", len(files), chatId)

	return files, nil
}

func (s *QueriesStorage) IsExist(f *File) (bool, error) {
	fileName, err := getFileName(f)
	if err != nil {
		return false, e.WrapIfErr("couldn't check if file exists", err)
	}

	path := filepath.Join(s.basPath, strconv.Itoa(f.ChatID), fileName)

	switch _, err = os.Stat(path); {
	case errors.Is(err, os.ErrNotExist):
		return false, nil
	case err != nil:
		return false, e.WrapIfErr(fmt.Sprintf("couldn't check if file %s exists", path), err)
	}

	log.Println("file exists: ", path)

	return true, nil
}

func (s *QueriesStorage) decodeFile(filePath string) (file *File, err error) {
	defer func() { err = e.WrapIfErr("couldn't decode file", err) }()

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err = gob.NewDecoder(f).Decode(&file); err != nil {
		return nil, err
	}

	return file, nil
}

func getFileName(f *File) (string, error) {
	return f.Hash()
}
