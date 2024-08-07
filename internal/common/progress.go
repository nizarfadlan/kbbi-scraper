package common

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const PROGRESS_FILE = "scrape_progress.json"

type Progress struct {
	CurrentLetter string `json:"current_letter"`
	CurrentPage   int    `json:"current_page"`
}

func SaveProgress(p Progress) {
	data, err := json.Marshal(p)
	if err != nil {
		log.Printf("Error marshaling progress: %v", err)
		return
	}

	err = ioutil.WriteFile(PROGRESS_FILE, data, 0644)
	if err != nil {
		log.Printf("Error saving progress: %v", err)
	}
}

func LoadProgress() Progress {
	data, err := ioutil.ReadFile(PROGRESS_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return Progress{}
		}
		log.Printf("Error reading progress file: %v", err)
		return Progress{}
	}

	var p Progress
	err = json.Unmarshal(data, &p)
	if err != nil {
		log.Printf("Error unmarshaling progress: %v", err)
		return Progress{}
	}

	return p
}
