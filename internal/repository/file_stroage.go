package repository

import (
	"encoding/json"
	"os"
)

type Entity struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type producerT struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filename string) (*producerT, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &producerT{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *producerT) WriteEntity(event *Entity) error {
	err := p.encoder.Encode(event)
	return err
}

func (p *producerT) Close() error {
	return p.file.Close()
}

type consumerT struct {
	file    *os.File
	decoder *json.Decoder
}

func NewConsumer(fileName string) (*consumerT, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0777)
	if err != nil {
		return nil, err
	}
	return &consumerT{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *consumerT) ReadEntity() (*Entity, error) {
	event := &Entity{}
	err := c.decoder.Decode(event)
	return event, err
}

func (c *consumerT) Close() error {
	return c.file.Close()
}
