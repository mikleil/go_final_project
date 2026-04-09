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
		writeJSONError(w, "ошибка десериализации JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		writeJSONError(w, "Не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	if err := checkDate(&task); err != nil {
		writeJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeJSONError(w, "ошибка добавления задачи: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJson(w, map[string]string{
		"id": fmt.Sprintf("%d", id),
	})
}
