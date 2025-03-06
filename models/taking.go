package models

import "time"

type Taking struct {
	ID           int64     `json:"id"`
	TakingTime   time.Time `json:"taking_time"`
	MedicationID int64     `json:"medication_id"` // это поле приходит из таблицы schedules
}
