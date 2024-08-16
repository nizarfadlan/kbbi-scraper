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

const SESSION_FILE = "session.json"

type Session struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Cookie   string `json:"cookie"`
}

func CheckSessionExists() bool {
	_, err := os.Stat(SESSION_FILE)
	return !os.IsNotExist(err)
}

func SaveSession(s Session) {
	data, err := json.Marshal(s)
	if err != nil {
		log.Printf("Error marshaling session: %v", err)
		return
	}

	err = os.WriteFile(SESSION_FILE, data, 0644)
	if err != nil {
		log.Printf("Error saving session: %v", err)
	}
}

func LoadSession() Session {
	data, err := os.ReadFile(PROGRESS_FILE)
	if err != nil {
		if os.IsNotExist(err) {
			return Session{}
		}
		log.Printf("Error reading session file: %v", err)
		return Session{}
	}

	var s Session
	err = json.Unmarshal(data, &s)
	if err != nil {
		log.Printf("Error unmarshaling session: %v", err)
		return Session{}
	}

	return s
}
