package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/ncruces/zenity"
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

func main() {

	var csvFile = getFile()
	csvwriter := csv.NewWriter(csvFile)

	// Get todos
	todosResponse, err := apiReq("https://systemapi.prod.ashish.me/todos/incomplete", "GET")
	if err != nil {
		fmt.Println("[E] Failed to get tasks", err)
		os.Exit(1)
	}

	// parse response
	var todos []Todo
	err = json.Unmarshal([]byte(todosResponse), &todos)
	if err != nil {
		fmt.Println("[E] Failed to parse", err)
		os.Exit(1)
	}

	var chore = Todo{}
	chore.Category = "Chore"
	chore.Content = "Other chore"

	var social = Todo{}
	social.Category = "Social"
	social.Content = "Mobile App"

	todos = append(todos, chore)
	todos = append(todos, social)

	// Create list
	var taskArray []string
	for i := 0; i < len(todos); i++ {
		taskArray = append(taskArray, todos[i].Content)
	}

	for true {

		// if time is right
		if float64(rand.Intn(9999999)) < 100 {

			// Show alert
			task, err := zenity.List(
				"Select task from the list below:",
				taskArray[:],
				zenity.Title("Time tracker"),
				zenity.Height(400),
			)

			// find task
			idx := Find(todos, func(value interface{}) bool {
				return value.(Todo).Content == task
			})

			// save data in csv file
			if task != "" && err == nil && idx >= 0 {
				row := []string{time.Now().Format(time.RFC3339), todos[idx].Content, todos[idx].Category}
				csvwriter.Write(row)
			} else {
				row := []string{time.Now().Format(time.RFC3339), "Timeout", "Timeout"}
				csvwriter.Write(row)
			}
			csvwriter.Flush()
			time.Sleep(120 * time.Second)
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
		row := []string{"Date", "Task", "Category"}
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
