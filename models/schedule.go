package models

type ScheduleRequest struct {
	UserID       int64  `json:"user_id"`
	MedicationID int64  `json:"medication_id"`
	Frequency    string `json:"frequency"` // формат INTERVAL, например "1 hour"
	Duration     string `json:"duration"`  // формат INTERVAL, например "7 days"
}
