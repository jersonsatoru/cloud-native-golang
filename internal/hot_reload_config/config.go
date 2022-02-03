package hot_reload_config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type Config struct {
	Host string                 `json:"host"`
	Port int                    `json:"port"`
	Tags map[string]interface{} `json:"tags"`
}

var ConfigFile Config

func init() {
	updates, errors, err := WatchConfig("config.json")
	if err != nil {
		panic(err)
	}

	go StartListening(updates, errors)
}

func LoadConfiguration(filename string) (Config, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY, 0777)
	if err != nil {
		return Config{}, fmt.Errorf("failed to open file: %s, %v", filename, err)
	}
	stream, err := io.ReadAll(file)
	if err != nil {
		return Config{}, fmt.Errorf("failed to handle stream: %v", err)
	}
	config := Config{}
	err = json.Unmarshal(stream, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to encode json: %v", err)
	}
	return config, nil
}

func StartListening(updates <-chan string, errors <-chan error) {
	for {
		select {
		case filepath := <-updates:
			c, err := LoadConfiguration(filepath)
			if err != nil {
				log.Println("error loading config:", err)
				continue
			}
			ConfigFile = c
		case err := <-errors:
			log.Println("error watching config:", err)
		}
	}
}

func WatchConfig(filename string) (chan string, chan error, error) {
	ch := make(chan string)
	errors := make(chan error)
	hash := ""

	go func() {
		ticker := time.NewTicker(time.Second * 3)
		for range ticker.C {
			log.Println(hash)
			newHash, err := CalculateHash(filename)
			if err != nil {
				errors <- err
				continue
			}
			if hash != newHash {
				hash = newHash
				ch <- filename
			}
		}
	}()

	return ch, errors, nil
}

func CalculateHash(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	sum := fmt.Sprintf("%x", hash.Sum(nil))
	return sum, nil
}
