package common

import (
	"log"
	"os"
)

func LogError(message string, err error) {
	logFile, openErr := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		log.Printf("Failed to open log file: %v", openErr)
		return
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)
	logger.Printf("%s: %v", message, err)
}

func LogInfo(message string) {
	logFile, openErr := os.OpenFile("info.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if openErr != nil {
		log.Printf("Failed to open log file: %v", openErr)
		return
	}
	defer logFile.Close()

	logger := log.New(logFile, "", log.LstdFlags)
	logger.Println(message)
}
