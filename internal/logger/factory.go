package logger

import (
	"fmt"
	"os"

	"github.com/jersonsatoru/cnb/internal/core"
)

func NewTransactionLogger(transactionType string) (core.TransactionLogger, error) {
	switch transactionType {
	case "file":
		filename := os.Getenv("TRANSACTION_FILENAME")
		tl, err := NewFileTransactionLogger(filename)
		if err != nil {
			return nil, err
		}
		return tl, nil
	case "postgres":
		pgParams := PostgresDBParams{
			DbName:   os.Getenv("POSTGRES_DBNAME"),
			Host:     os.Getenv("POSTGRES_HOST"),
			User:     os.Getenv("POSTGRES_USER"),
			Password: os.Getenv("POSTGRES_PASSWORD"),
		}
		tl, err := NewPostgresTransactionLogger(pgParams)
		if err != nil {
			return nil, err
		}
		return tl, nil
	default:
		return nil, fmt.Errorf("undefined transaction logger type")
	}
}
