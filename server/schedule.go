package server

import (
	"time"
)

func isValidFrequency(frequency time.Duration) bool {
	// проверяем, что duration больше или равно 1 час и меньше или равно 24 часам (1 день) (согласно ТЗ)
	return frequency >= time.Hour && frequency <= 24*time.Hour
}

func generateSchedule(startDate, endDate time.Time, frequency time.Duration) []time.Time {
	var schedule []time.Time

	for currentTime := startDate; !currentTime.After(endDate); currentTime = currentTime.Add(frequency) {
		remainder := currentTime.Minute() % 15
		// округляем в большую сторону до ближайших 15 минут, если минуты не кратны 15 (согласно ТЗ)
		if remainder != 0 {
			currentTime = currentTime.Add(time.Duration(15-remainder) * time.Minute)
		}

		// если мы уже в новых сутках, день не увеличиваем
		if currentTime.Hour() < 8 {
			currentTime = time.Date(
				currentTime.Year(),
				currentTime.Month(),
				currentTime.Day(),
				8, 0, 0, 0,
				currentTime.Location(),
			)
		}

		// если все еще в старых, переносим на утро завтрашнего дня
		if currentTime.Hour() > 21 {
			currentTime = time.Date(
				currentTime.Year(),
				currentTime.Month(),
				currentTime.Day()+1,
				8, 0, 0, 0,
				currentTime.Location(),
			)
		}

		// проверяем, не превысили ли мы endDate после округления
		if currentTime.After(endDate) {
			break
		}

		schedule = append(schedule, currentTime)
	}

	return schedule
}
