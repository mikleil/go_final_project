package api

import (
	"encoding/json"
	"fmt"
	"go_final_project/pkg/db"

	"net/http"
)

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&task); err != nil {
		writeJson(w, map[string]string{
			"error": "ошибка десериализации JSON: " + err.Error(),
		})
		return
	}

	if task.Title == "" {
		writeJson(w, map[string]string{
			"error": "Не указан заголовок задачи",
		})
		return
	}

	if err := checkDate(&task); err != nil {
		writeJson(w, map[string]string{
			"error": err.Error(),
		})
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJson(w, map[string]string{
			"error": "ошибка добавления задачи: " + err.Error(),
		})
		return
	}

	writeJson(w, map[string]string{
		"id": fmt.Sprintf("%d", id),
	})
}
