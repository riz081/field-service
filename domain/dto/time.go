package dto

import (
	"time"

	"github.com/google/uuid"
)

type TimeRequest struct {
	StartTime string `json:"startTime" validate:"required"`
	EndTime   string `json:"endTime" validate:"required"`
}

type TimeResponse struct {
	UUID      uuid.UUID  `json:"uuid"`
	StartTime string     `json:"startTime"`
	EndTime   string     `json:"endTime"`
	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}
