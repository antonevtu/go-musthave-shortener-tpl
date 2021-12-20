package repository

import (
	"errors"
	"io"
	"math/rand"
	"sync"
	"time"
)

type Repository struct {
	storage     storageT
	storageLock sync.Mutex
	producer    Producer
	consumer    Consumer
}

type storageT map[string]string

type Producer interface {
	WriteEntity(event *Entity) error
	Close() error
}

type Consumer interface {
	ReadEntity() (*Entity, error)
	Close() error
}

func New(producer Producer, consumer Consumer) (*Repository, error) {
	repository := Repository{
		storage:  make(storageT, 100),
		producer: producer,
		consumer: consumer,
	}
	// Восстановление хранилища в оперативной памяти
	for {
		readEvent, err := repository.consumer.ReadEntity()
		if err == io.EOF {
			return &repository, nil
		} else if err != nil {
			return &repository, err
		}
		repository.storage[readEvent.ID] = readEvent.URL
	}
}

func (r *Repository) Shorten(url string) (string, error) {
	const idLen = 5
	const attemptsNumber = 10
	r.storageLock.Lock()
	defer r.storageLock.Unlock()
	for i := 0; i < attemptsNumber; i++ {
		id := randStringRunes(idLen)
		if _, ok := r.storage[id]; !ok {
			r.storage[id] = url
			err := r.producer.WriteEntity(&Entity{
				ID:  id,
				URL: url,
			})
			return id, err
		}
	}
	return "", errors.New("can't generate random ID")
}

func (r *Repository) Expand(id string) (string, error) {
	r.storageLock.Lock()
	defer r.storageLock.Unlock()
	longURL, ok := r.storage[id]
	if ok {
		return longURL, nil
	} else {
		return longURL, errors.New("a non-existent ID was requested")
	}
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
