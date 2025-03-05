package models

type MedicationRequest struct {
	MedicationID int64  `json:"medication_id"`
	Name         string `json:"name"`
}
