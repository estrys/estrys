package models

import "time"

type Tweet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Published time.Time `json:"published"`
	Sensitive bool      `json:"sensitive"`
}
