package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func Up(port string) error {
	r := gin.Default()

	// создание расписания
	r.POST("/schedule", createSchedule)
	// получение всех расписаний для пользователя
	r.GET("/schedules", getSchedules)
	// получение конкретного расписания
	r.GET("/schedule", getSchedule)
	// получение следующего приема
	r.GET("/next_takings", getNextTakings)

	if err := r.Run(fmt.Sprintf(":%s", port)); err != nil {
		return fmt.Errorf("failed to Listen and Serve: %w", err)
	}

	return nil
}
