package models

type ScheduleRequest struct {
	UserID       int64  `json:"user_id"`
	MedicationID int64  `json:"medication_id"`
	Frequency    string `json:"frequency"`
	Duration     string `json:"duration"`
}
