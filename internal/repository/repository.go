package repository

import (
	"errors"
	"io"
	"log"
	"math/rand"
	"sync"
	"time"
)

type Repository struct {
	storage     storageT
	storageLock sync.Mutex
	producer    *producerT
	//consumer *consumerT
}
type storageT map[string]string

//type DiskSaver interface {
//	WriteEvent(event *Event) error
//	ReadEvent() (*Event, error)
//}

func New(fileStoragePath string) *Repository {
	producer, err := NewProducer(fileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	repository := Repository{
		storage:  make(storageT, 100),
		producer: producer,
	}
	repository.restoreStorage(fileStoragePath)
	return &repository
}

func (r *Repository) restoreStorage(fileStoragePath string) {
	consumer, err := NewConsumer(fileStoragePath)
	if err != nil {
		log.Fatal(err)
	}
	defer consumer.Close()

	for {
		readedEvent, err := consumer.ReadEvent()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		r.storage[readedEvent.ID] = readedEvent.URL
	}
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

func (r *Repository) Shorten(url string) (string, error) {
	const idLen = 5
	const attemptsNumber = 10
	r.storageLock.Lock()
	defer r.storageLock.Unlock()
	for i := 0; i < attemptsNumber; i++ {
		id := randStringRunes(idLen)
		if _, ok := r.storage[id]; !ok {
			r.storage[id] = url
			err := r.producer.WriteEvent(&Event{
				ID:  id,
				URL: url,
			})
			return id, err
		}
	}
	return "", errors.New("can't generate random ID")
}

func (r *Repository) Close() error {
	return r.producer.Close()
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
