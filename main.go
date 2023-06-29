package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type todo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
	Background  string `json:"background"`
	Foreground  string `json:"foreground"`
}

type dayData struct {
	Date      time.Time
	DayOfYear int
	DayName   string
	Todos     []todo
}

type week struct {
	Days       [7]*dayData
	WeekNumber int
}

func loadTodosFromFile(filename string) map[string][]todo {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to load todo file: %v", err)
	}

	var todos map[string][]todo
	err = json.Unmarshal(data, &todos)
	if err != nil {
		log.Fatalf("Failed to parse todo file: %v", err)
	}

	return todos
}

func getCurrentMonthCalendar() []week {
	todos := loadTodosFromFile("todos.json")

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, time.UTC)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	_, firstWeek := firstOfMonth.ISOWeek()
	_, lastWeek := lastOfMonth.ISOWeek()

	weeksInMonth := lastWeek - firstWeek + 1

	calendar := make([]week, weeksInMonth)
	for i := range calendar {
		calendar[i].Days = [7]*dayData{}
		calendar[i].WeekNumber = firstWeek + i
	}

	day := firstOfMonth
	for day.Month() == currentMonth {
		_, week := day.ISOWeek()
		index := week - firstWeek
		adjustedWeekday := (int(day.Weekday()) + 6) % 7 // Adjusting weekdays to start with Monday

		formattedDate := day.Format("2006/01/02")
		dayTodos, ok := todos[formattedDate]
		if ok {
			calendar[index].Days[adjustedWeekday] = &dayData{
				Date:      day,
				DayOfYear: day.YearDay(),
				DayName:   day.Weekday().String(),
				Todos:     dayTodos,
			}
		} else {
			calendar[index].Days[adjustedWeekday] = &dayData{
				Date:      day,
				DayOfYear: day.YearDay(),
				DayName:   day.Weekday().String(),
			}
		}

		day = day.AddDate(0, 0, 1)
	}

	return calendar
}

func main() {
	// Serve static files (CSS, JS, images)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Render the calendar
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFiles("calendar.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Month":    time.Now().Month().String(),
			"Year":     time.Now().Year(),
			"Calendar": getCurrentMonthCalendar(),
		}

		err = t.Execute(w, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Start the server
	http.ListenAndServe(":8080", nil)
}
