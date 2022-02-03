package logger

import (
	"bufio"
	"fmt"
	"os"

	"github.com/jersonsatoru/cnb/internal/core"
)

type FileTransactionLogger struct {
	events       chan<- core.Event
	errors       <-chan error
	lastSequence uint64
	file         *os.File
}

func NewFileTransactionLogger(filename string) (core.TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return nil, err
	}
	return &FileTransactionLogger{file: file}, nil
}

func (f *FileTransactionLogger) WritePut(key, value string) {
	f.events <- core.Event{Key: key, Value: value, EventType: core.EventPut}
}

func (f *FileTransactionLogger) WriteDelete(key string) {
	f.events <- core.Event{Key: key, EventType: core.EventDelete}
}

func (f *FileTransactionLogger) Err() <-chan error {
	return f.errors
}

func (f *FileTransactionLogger) Run() {
	events := make(chan core.Event, 16)
	f.events = events
	errors := make(chan error, 1)
	f.errors = errors

	go func() {
		defer f.file.Close()
		for event := range events {
			f.lastSequence++
			_, err := fmt.Fprintf(f.file, "%d\t%d\t%s\t%s",
				f.lastSequence,
				event.EventType,
				event.Key,
				event.Value)
			if err != nil {
				errors <- err
				return
			}
		}
	}()
}

func (f *FileTransactionLogger) ReadEvents() (<-chan core.Event, <-chan error) {
	scanner := bufio.NewScanner(f.file)
	chEvents := make(chan core.Event)
	chErrors := make(chan error)

	go func() {
		var e core.Event
		defer close(chEvents)
		defer close(chErrors)
		for scanner.Scan() {
			line := scanner.Text()
			if _, err := fmt.Sscanf(line,
				"%d\t%d\t%s\t%s",
				&e.Sequence,
				&e.EventType,
				&e.Key,
				&e.Value); err != nil {
				chErrors <- err
				return
			}
			if f.lastSequence >= e.Sequence {
				chErrors <- fmt.Errorf("transacton numbers out of sequence")
				return
			}
			f.lastSequence = e.Sequence
			chEvents <- e
		}

		if err := scanner.Err(); err != nil {
			chErrors <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()

	return chEvents, chErrors
}
