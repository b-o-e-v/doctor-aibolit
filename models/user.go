package models

type UserRequest struct {
	UserID int64  `json:"user_id"`
	Name   string `json:"name"`
}
