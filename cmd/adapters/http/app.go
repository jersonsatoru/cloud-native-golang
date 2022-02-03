package http

import (
	"log"
	"os"

	"github.com/jersonsatoru/cnb/internal/core"
	"github.com/jersonsatoru/cnb/internal/logger"
)

func _() {
	appPort := os.Getenv("APP_PORT")
	loggerType := os.Getenv("LOGGER_TYPE")
	transactionLogger, err := logger.NewTransactionLogger(loggerType)
	if err != nil {
		log.Fatal(err)
	}
	keyValueStore := core.NewKeyValueStore(transactionLogger)
	err = keyValueStore.Restore()
	if err != nil {
		log.Fatal(err)
	}
	server := NewHttpServer(keyValueStore)
	log.Fatal(server.Start(appPort))
}
