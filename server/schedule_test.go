package server

import (
	"testing"
	"time"
)

func TestGenerateSchedule(t *testing.T) {
	tests := []struct {
		name           string
		startDate      time.Time
		endDate        time.Time
		frequency      time.Duration
		expectedResult []time.Time
	}{
		{
			name:      "Testing long frequency",
			startDate: time.Date(2025, 3, 5, 9, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 3, 12, 23, 59, 0, 0, time.UTC),
			frequency: 17 * time.Hour,
			expectedResult: []time.Time{
				time.Date(2025, 3, 5, 9, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 6, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 7, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 8, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 9, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 10, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 11, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 12, 8, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "Testing hours with minutes in frequency",
			startDate: time.Date(2025, 3, 5, 15, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 3, 6, 12, 0, 0, 0, time.UTC),
			frequency: 1*time.Hour + 44*time.Minute,
			expectedResult: []time.Time{
				time.Date(2025, 3, 5, 15, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 5, 16, 45, 0, 0, time.UTC),
				time.Date(2025, 3, 5, 18, 30, 0, 0, time.UTC),
				time.Date(2025, 3, 5, 20, 15, 0, 0, time.UTC),
				time.Date(2025, 3, 6, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 6, 9, 45, 0, 0, time.UTC),
				time.Date(2025, 3, 6, 11, 30, 0, 0, time.UTC),
			},
		},
		{
			name:      "Testing rounding to multiples of 15",
			startDate: time.Date(2025, 3, 5, 17, 7, 0, 0, time.UTC),
			endDate:   time.Date(2025, 3, 7, 13, 27, 0, 0, time.UTC),
			frequency: 6 * time.Hour,
			expectedResult: []time.Time{
				time.Date(2025, 3, 5, 17, 15, 0, 0, time.UTC),
				time.Date(2025, 3, 6, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 6, 14, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 6, 20, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 7, 8, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "Testing table creation late at night",
			startDate: time.Date(2025, 3, 5, 22, 0, 0, 0, time.UTC),
			endDate:   time.Date(2025, 3, 6, 9, 0, 0, 0, time.UTC),
			frequency: time.Hour,
			expectedResult: []time.Time{
				time.Date(2025, 3, 6, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 6, 9, 0, 0, 0, time.UTC),
			},
		},
		{
			name:      "Testing table creation in the morning",
			startDate: time.Date(2025, 3, 5, 7, 58, 0, 0, time.UTC),
			endDate:   time.Date(2025, 3, 5, 8, 15, 0, 0, time.UTC),
			frequency: time.Minute * 5,
			expectedResult: []time.Time{
				time.Date(2025, 3, 5, 8, 0, 0, 0, time.UTC),
				time.Date(2025, 3, 5, 8, 15, 0, 0, time.UTC),
			},
		},
		{
			name:           "Test empty schedule",
			startDate:      time.Date(2025, 3, 5, 10, 0, 0, 0, time.UTC),
			endDate:        time.Date(2025, 3, 5, 9, 0, 0, 0, time.UTC),
			frequency:      time.Hour,
			expectedResult: []time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSchedule(tt.startDate, tt.endDate, tt.frequency)
			if !equal(result, tt.expectedResult) {
				t.Errorf("generateSchedule() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// функция для сравнения временных меток
func equal(a, b []time.Time) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equal(b[i]) {
			return false
		}
	}
	return true
}

func TestIsValidFrequency(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected bool
	}{
		// допустимые значения частоты
		{"1Hour", 1 * time.Hour, true},
		{"2Hours", 2 * time.Hour, true},
		{"1Day", 24 * time.Hour, true},
		{"Near1Day", 23*time.Hour + 59*time.Minute, true},

		// недопустимые значения частоты
		{"ZeroHours", 0 * time.Hour, false},
		{"LessThan1Hour", 30 * time.Minute, false},
		{"MoreThan1Day", 25 * time.Hour, false},
		{"2Days", 48 * time.Hour, false},
		{"NegativeDuration", -1 * time.Hour, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidFrequency(tt.duration)
			if got != tt.expected {
				t.Errorf("isValidFrequency(%v) = %v; want %v", tt.duration, got, tt.expected)
			}
		})
	}
}
