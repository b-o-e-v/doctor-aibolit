package server

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/b-o-e-v/doctor-aibolit/pkg/db"
	"github.com/b-o-e-v/doctor-aibolit/pkg/envs"
	"github.com/b-o-e-v/doctor-aibolit/pkg/utils"
	"github.com/gin-gonic/gin"
)

// для создания расписания необходим пользователь и лекарство
// (в перспективе для их создания нужен отдельный маршрут)
func checkEntityExists(tx *sql.Tx, query string, createQuery string, id int64, defaultName string) error {
	var exists bool
	if err := tx.QueryRow(query, id).Scan(&exists); err != nil {
		return err
	}

	if !exists {
		if _, err := tx.Exec(createQuery, id, defaultName); err != nil {
			return err
		}
	}

	return nil
}

type ScheduleRequest struct {
	UserID       int64  `json:"user_id"`
	MedicationID int64  `json:"medication_id"`
	Frequency    string `json:"frequency"`
	Duration     string `json:"duration,omitempty"`
}

// создание расписания
func createSchedule(ctx *gin.Context) {
	var data ScheduleRequest

	// парсим JSON
	if err := ctx.ShouldBindJSON(&data); err != nil {
		handleError(ctx, "incorrect data", err, http.StatusBadRequest)
		return
	}

	// если не передана продолжительность, устанавливаем ее по умолчанию
	if data.Duration == "" {
		data.Duration = envs.Config.DefaultDuration
	}

	tx, err := db.Conn.Begin()
	if err != nil {
		handleError(ctx, "failed to start database transaction", err)
		return
	}
	defer tx.Rollback()

	// пользователь и лекарство необходимы для связей таблиц
	if err := checkEntityExists(tx, QUERY_USER_EXISTS, QUERY_USER_CREATE, data.UserID, "sick"); err != nil {
		handleError(ctx, "failed to check user exists", err)
		return
	}

	if err := checkEntityExists(tx, QUERY_MEDICATION_EXISTS, QUERY_MEDICATION_CREATE, data.MedicationID, "ascorbic"); err != nil {
		handleError(ctx, "failed to check medication exists", err)
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
		handleError(ctx, "error writing schedule to database, check input data", err)
		return
	}

	// генерируем расписание
	timing := generateSchedule(startDate, endDate, time.Duration(frequencySeconds)*time.Second)

	for _, taking := range timing {
		if _, err := tx.Exec(QUERY_TAKING_CREATE, scheduleID, taking); err != nil {
			handleError(ctx, "failed to insert taking time", err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		handleError(ctx, "failed to commit transaction", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"schedule_id": scheduleID})
}

// получение всех ids расписаний юзера
func getSchedules(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		handleError(ctx, "invalid user_id, must be an integer", err, http.StatusBadRequest)
		return
	}

	if userID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "data is required",
		})
		return
	}

	rows, err := db.Conn.Query(QUERY_USER_SCHEDULES, userID)
	if err != nil {
		handleError(ctx, "failed to fetch user schedules", err)
		return
	}
	defer rows.Close()

	var scheduleIDs []int
	for rows.Next() {
		var ID int
		if err := rows.Scan(&ID); err != nil {
			handleError(ctx, "failed to scan schedule ID", err)
			return
		}
		scheduleIDs = append(scheduleIDs, ID)
	}

	if rows.Err() != nil {
		handleError(ctx, "failed to process rows", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user_id": userID, "schedules": scheduleIDs})
}

type Taking struct {
	ID           int64     `json:"id"`
	TakingTime   time.Time `json:"taking_time"`
	MedicationID int64     `json:"medication_id"`
}

// получение конкретного расписания на сегодняшний день
func getSchedule(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	scheduleIDStr := ctx.Query("schedule_id")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		handleError(ctx, "invalid user_id, must be an integer", err, http.StatusBadRequest)
		return
	}

	scheduleID, err := strconv.Atoi(scheduleIDStr)
	if err != nil {
		handleError(ctx, "invalid schedule_id, must be an integer", err, http.StatusBadRequest)
		return
	}

	if userID == 0 || scheduleID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "data is required",
		})
		return
	}

	rows, err := db.Conn.Query(QUERY_USER_SCHEDULE, userID, scheduleID)
	if err != nil {
		handleError(ctx, "failed to fetch user schedule", err)
		return
	}
	defer rows.Close()

	var takings []Taking
	for rows.Next() {
		var taking Taking
		if err := rows.Scan(
			&taking.ID,
			&taking.TakingTime,
			&taking.MedicationID,
		); err != nil {
			handleError(ctx, "failed to scan takings", err)
			return
		}
		takings = append(takings, taking)
	}

	if rows.Err() != nil {
		handleError(ctx, "failed to process rows", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_id":     userID,
		"schedule_id": scheduleID,
		"takings":     takings,
	})
}

type TakingWithScheduleID struct {
	Taking
	ScheduleID int64 `json:"schedule_id"`
}

// получение ближайших приемов лекарств
func getNextTakings(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		handleError(ctx, "invalid user_id, must be an integer", err, http.StatusBadRequest)
		return
	}

	if userID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "data is required",
		})
		return
	}

	rows, err := db.Conn.Query(QUERY_USER_TAKINGS, envs.Config.ComingPeriod, userID)
	if err != nil {
		handleError(ctx, "failed to fetch user takings", err)
		return
	}
	defer rows.Close()

	var takings []TakingWithScheduleID
	for rows.Next() {
		var taking TakingWithScheduleID
		if err := rows.Scan(
			&taking.ID,
			&taking.TakingTime,
			&taking.MedicationID,
			&taking.ScheduleID,
		); err != nil {
			handleError(ctx, "failed to scan takings", err)
			return
		}
		takings = append(takings, taking)
	}

	if rows.Err() != nil {
		handleError(ctx, "failed to process rows", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"takings": takings,
	})
}

func handleError(ctx *gin.Context, message string, err error, statusCode ...int) {
	if len(statusCode) == 0 {
		statusCode = append(statusCode, http.StatusInternalServerError)
	}

	ctx.JSON(statusCode[0], gin.H{
		"error": utils.FormatErrorMessage(message, err),
	})
}
