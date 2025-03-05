package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func createSchedule(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "schedule created"})
}

func getSchedules(c *gin.Context) {
	userID := c.Query("user_id")
	c.JSON(http.StatusOK, gin.H{"message": "schedules fetched", "user_id": userID})
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
