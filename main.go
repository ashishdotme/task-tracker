package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/ncruces/zenity"
	"golang.org/x/sys/windows"
)

type Todo struct {
	ID            int         `json:"id"`
	Content       string      `json:"content"`
	Category      string      `json:"category"`
	ProjectID     string      `json:"projectId"`
	TodoID        string      `json:"todoId"`
	Completed     bool        `json:"completed"`
	CompletedDate interface{} `json:"completedDate"`
	DueDate       time.Time   `json:"dueDate"`
	CreatedAt     time.Time   `json:"createdAt"`
	UpdatedAt     time.Time   `json:"updatedAt"`
}

var user32 = windows.NewLazyDLL("user32.dll")

func getForegroundWindow() uintptr {
	hwnd, _, _ := user32.NewProc("GetForegroundWindow").Call()
	return hwnd
}
func main() {

	var csvFile = getFile()
	csvwriter := csv.NewWriter(csvFile)

	for true {
		hwnd := getForegroundWindow()
		// if time is right

		// Get todos
		todosResponse, err := apiReq("https://api.ashish.me/todos/incomplete", "GET")
		if err != nil {
			fmt.Println("[E] Failed to get tasks", err)
		}

		// parse response
		var todos []Todo
		err = json.Unmarshal([]byte(todosResponse), &todos)
		if err != nil {
			fmt.Println("[E] Failed to parse", err)
		}

		var chore = Todo{}
		chore.Category = "Chore"
		chore.Content = "Household Chore"

		var dev = Todo{}
		dev.Category = "Programming"
		dev.Content = "Other coding task"

		var youtube = Todo{}
		youtube.Category = "Social"
		youtube.Content = "Youtube"

		var news = Todo{}
		news.Category = "Social"
		news.Content = "News"

		var fb = Todo{}
		fb.Category = "Social"
		fb.Content = "Facebook/Instagram/Discord"

		todos = append(todos, dev)
		todos = append(todos, chore)
		todos = append(todos, youtube)
		todos = append(todos, news)
		todos = append(todos, fb)
		newTask := Todo{
			Category: "General",
			Content:  "New task",
		}
		todos = append(todos, newTask)

		// Create list
		var taskArray []string
		for i := 0; i < len(todos); i++ {
			taskArray = append(taskArray, todos[i].Content)
		}

		// ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		// defer cancel()

		// Show alert
		task, err := zenity.List(
			"Select task from the list below:",
			taskArray[:],
			zenity.Title("Time tracker"),
			zenity.CancelLabel("Exit"),
			zenity.Height(400),
			zenity.Attach(hwnd),
		)

		if task == "New task" {
			var newTask, _ = zenity.Entry("Add new task")
			if err := sendNewTask(newTask, os.Getenv("ASHISHDOTME_TOKEN")); err != nil {
				fmt.Printf("Error: %v\n", err)
			} else {
				fmt.Println("Task sent successfully!")
			}
			time.Sleep(8 * time.Minute)
		} else {

			// find task
			idx := Find(todos, func(value interface{}) bool {
				return value.(Todo).Content == task
			})

			// save data in csv file
			if task != "" && err == nil && idx >= 0 {
				t, _ := time.Now().UTC().MarshalText()
				row := []string{string(t), time.Now().Format("01-02-2006"), todos[idx].Content, todos[idx].Category}
				csvwriter.Write(row)
			} else {
				if err == zenity.ErrCanceled {
					fmt.Println("Cancelled")
				}
				t, _ := time.Now().UTC().MarshalText()
				row := []string{string(t), time.Now().Format("01-02-2006"), "Timeout", "Timeout"}
				csvwriter.Write(row)
			}
			csvwriter.Flush()
			time.Sleep(8 * time.Minute)
		}
	}
	csvFile.Close()
}

func getFile() *os.File {
	if _, err := os.Stat("history.csv"); os.IsNotExist(err) {
		csvFile, err := os.Create("history.csv")
		if err != nil {
			panic(err)
		}
		csvwriter := csv.NewWriter(csvFile)
		row := []string{"Timestamp", "Day", "Task", "Category"}
		csvwriter.Write(row)
		csvwriter.Flush()
		return csvFile
	} else {
		csvFile, _ := os.OpenFile("history.csv", os.O_RDWR|os.O_APPEND, 0660)
		return csvFile
	}
}

func Find(slice interface{}, f func(value interface{}) bool) int {
	s := reflect.ValueOf(slice)
	if s.Kind() == reflect.Slice {
		for index := 0; index < s.Len(); index++ {
			if f(s.Index(index).Interface()) {
				return index
			}
		}
	}
	return -1
}

func apiReq(url string, method string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return "", err
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("[e] Failed", err)
		return "", err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println("[e] Failed", err)
		return "", err
	}
	return string(body), nil
}

func sendNewTask(newtask, apikey string) error {
	task := map[string]string{"content": newtask}
	jsonData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("error marshalling task: %v", err)
	}

	req, err := http.NewRequest("POST", "https://api.ashish.me/todos", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating POST request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apiKey", apikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("received non-OK response status: %s", resp.Status)
	}

	return nil
}
