package repository

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type Repository struct {
	Storage     storageT
	storageLock sync.Mutex
}
type storageT map[string]string

func (r *Repository) Load(id string) (string, error) {
	r.storageLock.Lock()
	defer r.storageLock.Unlock()
	longURL, ok := r.Storage[id]
	if ok {
		return longURL, nil
	} else {
		return longURL, errors.New("a non-existent ID was requested")
	}
}

func (r *Repository) Store(url string) (string, error) {
	const idLen = 5
	const attemptsNumber = 10
	r.storageLock.Lock()
	defer r.storageLock.Unlock()
	for i := 0; i < attemptsNumber; i++ {
		id := randStringRunes(idLen)
		if _, ok := r.Storage[id]; !ok {
			r.Storage[id] = url
			return id, nil
		}
	}
	return "", errors.New("can't generate random ID")
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

func (r *Repository) Init() {
	r.Storage = make(storageT, 100)
}
