/*
 *  Copyright (c) 2024 Nizar Izzuddin Yatim Fadlan <hello@nizarfadlan.dev>
 * All rights reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program. If not, see <http://www.gnu.org/licenses/>.
 */
package common

import (
	"encoding/json"
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

	err = os.WriteFile(PROGRESS_FILE, data, 0644)
	if err != nil {
		log.Printf("Error saving progress: %v", err)
	}
}

func LoadProgress() Progress {
	data, err := os.ReadFile(PROGRESS_FILE)
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
