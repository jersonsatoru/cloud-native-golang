package core

import (
	"errors"
	"log"
	"sync"
)

type KeyValueStore struct {
	m map[string]string
	sync.RWMutex
	tl TransactionLogger
}

var ErrNoSUchKey = errors.New("no such key")

func NewKeyValueStore(tl TransactionLogger) *KeyValueStore {
	return &KeyValueStore{
		m:  make(map[string]string),
		tl: tl,
	}
}

func (kv *KeyValueStore) Put(key, value string) error {
	kv.Lock()
	kv.m[key] = value
	kv.Unlock()
	kv.tl.WritePut(key, value)
	return nil
}

func (kv *KeyValueStore) Delete(key string) error {
	kv.Lock()
	delete(kv.m, key)
	kv.Unlock()
	kv.tl.WriteDelete(key)
	return nil
}

func (kv *KeyValueStore) Get(key string) (string, error) {
	kv.RLock()
	value, ok := kv.m[key]
	kv.RUnlock()
	if !ok {
		return "", ErrNoSUchKey
	}
	return value, nil
}

func (kv *KeyValueStore) Restore() error {
	var err error
	e, ok := Event{}, true
	kv.tl.Run()
	chEvents, chErrors := kv.tl.ReadEvents()
	for ok && err == nil {
		select {
		case e, ok = <-chEvents:
			log.Printf("%v", e)
			switch e.EventType {
			case EventDelete:
				kv.Delete(e.Key)
			case EventPut:
				kv.Put(e.Key, e.Value)
			}
		case err, ok = <-chErrors:
			return err
		}
	}
	return err
}
