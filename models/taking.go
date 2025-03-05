package models

import "time"

type Taking struct {
	ID         int64     `json:"id"`
	ScheduleID int64     `json:"schedule_id"`
	TakingTime time.Time `json:"taking_time"`
}
