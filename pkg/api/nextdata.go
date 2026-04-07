package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const DateFormat = "20060102"

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	// Если правило пустое — задача без повторения, возвращаем пустую строку
	if repeat == "" {
		return "", nil
	}

	startDate, err := time.ParseInLocation(DateFormat, dstart, now.Location())
	if err != nil {
		return "", fmt.Errorf("некорректный формат даты старта: %v", err)
	}

	loc := now.Location()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	startDate = time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc)

	parts := strings.Fields(repeat)
	if len(parts) < 1 {
		return "", fmt.Errorf("неверный формат правила повторения")
	}
	ruleType := parts[0]

	switch ruleType {
	case "d": // Повторение дни
		if len(parts) != 2 {
			return "", fmt.Errorf("неверный формат правила d: требуется число")
		}
		interval, err := strconv.Atoi(parts[1])
		if err != nil || interval <= 0 || interval > 400 {
			return "", fmt.Errorf("неверный интервал дней (1-400): %s", parts[1])
		}

		currentDate := startDate.AddDate(0, 0, interval)
		for !currentDate.After(now) {
			currentDate = currentDate.AddDate(0, 0, interval)
		}
		return currentDate.Format(DateFormat), nil

	case "y": // Повторение год
		currentDate := startDate.AddDate(1, 0, 0)
		for !currentDate.After(now) {
			currentDate = currentDate.AddDate(1, 0, 0)
		}
		return currentDate.Format(DateFormat), nil

	case "w": // Дни недели
		weekDays, err := parseIntList(parts[1:], 1, 7)
		if err != nil {
			return "", fmt.Errorf("неверный формат правила w: %v", err)
		}

		validWeekdays := make(map[time.Weekday]bool)
		for _, wd := range weekDays {
			if wd == 7 {
				validWeekdays[time.Sunday] = true
			} else {
				validWeekdays[time.Weekday(wd)] = true
			}
		}

		currentDate := startDate.AddDate(0, 0, 1)

		limit := now.AddDate(2, 0, 0)

		for currentDate.Before(limit) {
			if validWeekdays[currentDate.Weekday()] && currentDate.After(now) {
				return currentDate.Format(DateFormat), nil
			}
			currentDate = currentDate.AddDate(0, 0, 1)
		}
		return "", fmt.Errorf("не удалось найти дату для правила %s", repeat)

	case "m": // Дни месяца
		if len(parts) < 2 {
			return "", fmt.Errorf("неверный формат правила m")
		}

		daysList, err := parseDayList(parts[1])
		if err != nil {
			return "", fmt.Errorf("ошибка дней в правиле m: %v", err)
		}

		monthsList := []int{}
		if len(parts) >= 3 {
			monthsList, err = parseIntList(strings.Split(parts[2], ","), 1, 12)
			if err != nil {
				return "", fmt.Errorf("ошибка месяцев в правиле m: %v", err)
			}
		}

		currentDate := startDate.AddDate(0, 0, 1)

		limit := now.AddDate(2, 0, 0)

		for currentDate.Before(limit) {
			if matchMonthRule(currentDate, daysList, monthsList, loc) && currentDate.After(now) {
				return currentDate.Format(DateFormat), nil
			}
			currentDate = currentDate.AddDate(0, 0, 1)
		}
		return "", fmt.Errorf("не удалось найти дату для правила %s", repeat)

	default:
		return "", fmt.Errorf("неподдерживаемое правило повторения: %s", ruleType)
	}
}

// parseIntList парсит строку "1,3,5"
func parseIntList(parts []string, min, max int) ([]int, error) {
	var result []int
	for _, part := range parts {
		subParts := strings.Split(part, ",")
		for _, s := range subParts {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			val, err := strconv.Atoi(s)
			if err != nil {
				return nil, err
			}
			if val < min || val > max {
				return nil, fmt.Errorf("значение %d вне диапазона [%d, %d]", val, min, max)
			}
			result = append(result, val)
		}
	}
	return result, nil
}

// parseDayList парсит дни месяца
func parseDayList(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	var result []int
	for _, p := range parts {
		p = strings.TrimSpace(p)
		val, err := strconv.Atoi(p)
		if err != nil {
			return nil, err
		}
		if val >= 1 && val <= 31 {
			result = append(result, val)
		} else if val == -1 || val == -2 {
			result = append(result, val)
		} else {
			return nil, fmt.Errorf("недопустимый день месяца: %d", val)
		}
	}
	return result, nil
}

// matchMonthRule проверяет правило дней месяц
func matchMonthRule(date time.Time, allowedDays []int, allowedMonths []int, loc *time.Location) bool {
	day := date.Day()
	month := int(date.Month())

	// Проверка месяца
	if len(allowedMonths) > 0 {
		found := false
		for _, m := range allowedMonths {
			if m == month {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Проверка дня
	for _, d := range allowedDays {
		if d > 0 {
			if d == day {
				return true
			}
		} else {
			lastDay := time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, loc).Day()

			if d == -1 && day == lastDay {
				return true
			}
			if d == -2 && day == lastDay-1 {
				return true
			}
		}
	}

	return false
}

// nextDateHandler HTTP обработчик /api/nextdate
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeatStr := r.FormValue("repeat")

	var nowTime time.Time
	var err error

	if nowStr == "" {
		nowTime = time.Now()
	} else {
		nowTime, err = time.ParseInLocation(DateFormat, nowStr, time.Local)
		if err != nil {
			http.Error(w, fmt.Sprintf("Неверный формат now: %v", err), http.StatusBadRequest)
			return
		}
	}

	nextDate, err := NextDate(nowTime, dateStr, repeatStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}
