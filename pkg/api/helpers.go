package api

import (
	"encoding/json"
	"fmt"
	"go_final_project/pkg/db"
	"log"
	"net/http"
	"time"
)

// writeJSON сериализует данные в JSON и отправляет ответ
func writeJson(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"encoding failed"}`, http.StatusInternalServerError)
	}
}

// checkDate проверяет и корректирует дату задачи
func checkDate(task *db.Task) error {
	now := time.Now()
	// Если дата не указана, используем сегодняшнюю
	if task.Date == "" {
		task.Date = now.Format("20060102")
		return nil
	}

	// Парсим указанную дату
	t, err := time.Parse("20060102", task.Date)
	if err != nil {
		return fmt.Errorf("неверный формат даты: %w", err)
	}

	// Если указано правило повторения проверяем его и вычисляем следующую дату
	var nextDate string
	if task.Repeat != "" {
		nextDate, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return fmt.Errorf("ошибка в правиле повторения: %w", err)
		}
	}

	// Если дата задачи в прошлом
	if afterNow(now, t) {
		if task.Repeat == "" {
			task.Date = now.Format("20060102")
		} else {
			task.Date = nextDate
		}
	}

	return nil
}

// afterNow возвращает true, если сейчас позже указанной даты
func afterNow(now, t time.Time) bool {
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return nowDate.After(t)
}

// Отправка ошибки в формате JSON
func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		log.Printf("[API_ERROR] encode failed: %v | message: %q | status: %d",
			err, message, status)
	}
}
