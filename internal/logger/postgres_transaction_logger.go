package logger

import (
	"database/sql"
	"fmt"

	"github.com/jersonsatoru/cnb/internal/core"
	_ "github.com/lib/pq"
)

type PostgresTransactionLogger struct {
	chEvents chan core.Event
	chErrors chan error
	db       *sql.DB
}

type PostgresDBParams struct {
	DbName   string
	Host     string
	User     string
	Password string
}

func NewPostgresTransactionLogger(params PostgresDBParams) (core.TransactionLogger, error) {
	connStr := fmt.Sprintf("host=%s dbname=%s user=%s password=%s", params.Host, params.DbName, params.User, params.Password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}
	createTransactionTableIfNotExists(db)
	return &PostgresTransactionLogger{
		db: db,
	}, nil
}

func createTransactionTableIfNotExists(db *sql.DB) error {
	db.QueryRow(`
		create table transactions if not exists (
			sequence serial,
			event_type byte,
			key varchar,
			value varchar
		)
	`)
	return nil
}

func (p *PostgresTransactionLogger) WritePut(key, value string) {
	p.chEvents <- core.Event{EventType: core.EventPut, Value: value, Key: key}
}

func (p *PostgresTransactionLogger) WriteDelete(key string) {
	p.chEvents <- core.Event{EventType: core.EventDelete, Key: key}
}

func (p *PostgresTransactionLogger) Err() <-chan error {
	return p.chErrors
}

func (p *PostgresTransactionLogger) Run() {
	events := make(chan core.Event)
	p.chEvents = events
	errors := make(chan error)
	p.chErrors = errors

	go func() {
		sql := `
		INSERT INTO 
			transactions (event_type, key, value)
		($1, $2, $3)
	`
		for e := range events {
			_, err := p.db.Exec(sql, e.EventType, e.Key, e.Value)
			if err != nil {
				errors <- err
			}
		}
	}()
}

func (p *PostgresTransactionLogger) ReadEvents() (<-chan core.Event, <-chan error) {
	chEvents := make(chan core.Event)
	chErrors := make(chan error)

	go func() {
		defer close(chEvents)
		defer close(chErrors)
		sql := `
			SELECT sequence, event_type, key, value FROM transactions
		`
		rows, err := p.db.Query(sql)
		if err != nil {
			chErrors <- fmt.Errorf("error querying events: %v", err)
			rows.Close()
			return
		}
		var e core.Event
		for rows.Next() {
			err := rows.Scan(&e.Sequence, &e.EventType, &e.Key, &e.Value)
			if err != nil {
				chErrors <- fmt.Errorf("error scanning values: %v", err)
				return
			}
			chEvents <- e
		}

		err = rows.Err()
		if err != nil {
			chErrors <- fmt.Errorf("transaction log read failure: %v", err)
		}
	}()

	return chEvents, chErrors
}
