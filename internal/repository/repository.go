package repository

import (
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Repository struct {
	storage     storageT
	storageLock sync.Mutex
	fileWriter  fileWriterT
}

type storageT map[string]Entity

type Entity struct {
	UserID string `json:"user_id"`
	ID     string `json:"id"`
	URL    string `json:"url"`
}

type fileWriterT struct {
	file    *os.File
	encoder *json.Encoder
}

func New(fileName string) (*Repository, error) {
	repository := Repository{
		storage:    make(storageT, 100),
		fileWriter: fileWriterT{},
	}

	err := repository.restoreFromFile(fileName)
	if err != nil {
		return &repository, err
	}

	err = repository.fileWriter.new(fileName)
	if err != nil {
		return &repository, err
	}
	return &repository, nil
}

func (fw *fileWriterT) new(filename string) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}
	*fw = fileWriterT{
		file:    file,
		encoder: json.NewEncoder(file),
	}
	return nil
}

// restoreFromFile Восстановление хранилища в оперативной памяти из текстового файла
func (r *Repository) restoreFromFile(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	entity := Entity{}
	for {
		err = decoder.Decode(&entity)
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		r.storage[entity.ID] = entity
	}
}

func (r *Repository) Shorten(userID string, url string) (string, error) {
	const idLen = 5
	const attemptsNumber = 10
	r.storageLock.Lock()
	defer r.storageLock.Unlock()

	// generate and save random string ID for short url
	for i := 0; i < attemptsNumber; i++ {
		id := randStringRunes(idLen)
		if _, ok := r.storage[id]; !ok {
			entity := Entity{
				UserID: userID,
				ID:     id,
				URL:    url,
			}
			r.storage[id] = entity
			err := r.fileWriter.encoder.Encode(&entity)
			return id, err
		}
	}
	return "", errors.New("can't generate random ID")
}

func (r *Repository) Expand(id string) (string, error) {
	r.storageLock.Lock()
	defer r.storageLock.Unlock()
	entity, ok := r.storage[id]
	if ok {
		return entity.URL, nil
	} else {
		return "", errors.New("a non-existent ID was requested")
	}
}

func (r *Repository) SelectByUser(userID string) []Entity {
	r.storageLock.Lock()
	defer r.storageLock.Unlock()
	selection := make([]Entity, 0, 10)
	for _, entity := range r.storage {
		if userID == entity.UserID {
			selection = append(selection, entity)
		}
	}
	return selection
}

func (r *Repository) Close() {
	_ = r.fileWriter.file.Close()
}

func randStringRunes(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
