package api

import (
	"net/http"
)

func Init() {
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/tasks", tasksHandler)
	http.HandleFunc("/api/task", taskHandler)
	http.HandleFunc("/api/task/done", taskDoneHandler)

}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	case http.MethodDelete:
		taskDeleteHandler(w, r) // делегируем удаление

	}
}
