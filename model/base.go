package model

import "time"

type BaseModelID struct {
	ID uint `json:"id"`
}

type BaseModelTime struct {
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
