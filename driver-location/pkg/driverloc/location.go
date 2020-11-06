package driverloc

import (
	"time"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Location struct {
	*Coordinates
	Time time.Time `json:"updated_at"`
}
