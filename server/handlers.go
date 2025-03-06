package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/b-o-e-v/doctor-aibolit/models"
	"github.com/b-o-e-v/doctor-aibolit/pkg/db"
	"github.com/b-o-e-v/doctor-aibolit/pkg/envs"
	"github.com/gin-gonic/gin"
)

const QUERY_USER_EXISTS = `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`
const QUERY_USER_CREATE = `INSERT INTO users (id, name) VALUES ($1, $2) RETURNING id`
const QUERY_MEDICATION_EXISTS = `SELECT EXISTS(SELECT 1 FROM medications WHERE id=$1)`
const QUERY_MEDICATION_CREATE = `INSERT INTO medications (id, name) VALUES ($1, $2) RETURNING id`
const QUERY_TAKING_CREATE = `INSERT INTO takings (schedule_id, taking_time) VALUES ($1, $2)`
const QUERY_SCHEDULE_CREATE = `
  INSERT INTO schedules (user_id, medication_id, frequency, duration, start_date)
  VALUES ($1, $2, $3, $4, NOW())
  RETURNING id, start_date, end_date, EXTRACT(EPOCH FROM frequency) AS frequency_seconds
`
const QUERY_USER_SCHEDULES = `SELECT id FROM schedules WHERE user_id = $1`

// поскольку у нас есть отдельная таблица с пользователями, нам нужно проверить его существование
// в случае отсутствия пользователя, создадим его, пока мы не реализоываем регистрацию
// далее это можно будет перенести в отдельный маршрут
func checkUserExists(tx *sql.Tx, userID int64) error {
	var userExists bool
	if err := tx.QueryRow(QUERY_USER_EXISTS, userID).Scan(&userExists); err != nil {
		return fmt.Errorf("failed to check if user exists")
	}

	if !userExists {
		if _, err := tx.Exec(QUERY_USER_CREATE, userID, "sick"); err != nil {
			return fmt.Errorf("failed to create user")
		}
	}

	return nil
}

// то же самое с лекарствами (в перспективе для создания лекарств будет отдельный маршрут)
func checkMedicationExists(tx *sql.Tx, medicationID int64) error {
	var medicationExists bool
	if err := tx.QueryRow(QUERY_MEDICATION_EXISTS, medicationID).Scan(&medicationExists); err != nil {
		return fmt.Errorf("failed to check if medication exists")
	}

	if !medicationExists {
		if _, err := tx.Exec(QUERY_MEDICATION_CREATE, medicationID, "ascorbic"); err != nil {
			return fmt.Errorf("failed to create medication")
		}
	}

	return nil
}

func createSchedule(c *gin.Context) {
	var data models.ScheduleRequest

	// парсим JSON
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect data"})
		return
	}

	// если не передана продолжительность, устанавливаем ее по умолчанию
	if data.Duration == "" {
		data.Duration = envs.Config.DefaultDuration
	}

	tx, err := db.Conn.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}
	defer tx.Rollback()

	if err := checkUserExists(tx, data.UserID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	if err := checkMedicationExists(tx, data.MedicationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}

	var scheduleID, frequencySeconds int64
	var startDate, endDate time.Time
	if err := tx.QueryRow(
		QUERY_SCHEDULE_CREATE,
		data.UserID,
		data.MedicationID,
		data.Frequency,
		data.Duration,
	).Scan(&scheduleID, &startDate, &endDate, &frequencySeconds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error writing schedule to database, check input data"})
		return
	}

	timing := generateSchedule(startDate, endDate, time.Duration(frequencySeconds)*time.Second)

	for _, taking := range timing {
		if _, err := tx.Exec(QUERY_TAKING_CREATE, scheduleID, taking); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to insert taking time"})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedule_id": scheduleID})
}

func getSchedules(c *gin.Context) {
	userID := c.Query("user_id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	rows, err := db.Conn.Query("SELECT id FROM schedules WHERE user_id = $1", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch schedules"})
		return
	}
	defer rows.Close()

	var scheduleIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}
		scheduleIDs = append(scheduleIDs, id)
	}

	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating rows"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userID, "schedules": scheduleIDs})
}

func getSchedule(c *gin.Context) {
	userID := c.Query("user_id")
	scheduleID := c.Query("schedule_id")
	c.JSON(http.StatusOK, gin.H{"message": "schedule fetched", "user_id": userID, "schedule_id": scheduleID})
}

func getNextTakings(c *gin.Context) {
	userID := c.Query("user_id")
	c.JSON(http.StatusOK, gin.H{"message": "next takings fetched", "user_id": userID})
}
