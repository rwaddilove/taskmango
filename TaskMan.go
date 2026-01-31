// TaskManGo - A simple command-line task manager in Go
// By R.A.Waddilove 2025 - github.com/rwaddilove
// No copyright - free to use and modify

package main

import (
	"bufio"
	"cmp"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

const ( // ANSI color codes for terminal output
	Reset   = "\033[0m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[97m"
)

type Task struct {
	title    string
	due      time.Time
	priority string
	repeat   string
	label    string
	done     string
	notes    string
}

var taskList []Task // global task list

type Config struct { // global configuration data
	folderPath string // path to folder containing data file
	filePath   string // full path to data file
	extra2     string // reserved for future use
}

var config Config

// ReadConfig reads configuration from file, or creates default config if file not found
func ReadConfig() {
	path, _ := os.UserHomeDir() // should check for error, but no home folder? Unlikely
	file, err := os.Open(path + "/TaskManGoConfig.txt")
	if err != nil { // Create default config file
		config.folderPath = GetFolderPath() // get folder to store data file
		config.filePath = config.folderPath + "/TaskManGo.txt"
		WriteConfig()
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	data := []string{"", "", ""}
	for i := 0; scanner.Scan() && i < len(data); i++ {
		data[i] = strings.TrimSpace(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	config.folderPath = data[0]
	config.filePath = data[1]
	config.extra2 = data[2]
}

// WriteConfig writes current configuration to file in user's home directory
func WriteConfig() {
	path, _ := os.UserHomeDir() // should check for error, but no home folder? Unlikely
	file, err := os.Create(path + "/TaskManGoConfig.txt")
	if err != nil {
		fmt.Println("Error creating config file!")
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(config.folderPath + "\n")
	if err != nil {
		return
	}
	_, err = writer.WriteString(config.filePath + "\n")
	if err != nil {
		return
	}
	_, err = writer.WriteString(config.extra2 + "\n")
	if err != nil {
		return
	}
	writer.Flush()
}

// GetFolderPath prompts user for folder path to store data file when config file not found
func GetFolderPath() string {
	fmt.Println("\nWhere do you want to store your data file?")
	fmt.Println("Eg. /Users/name/Documents or C:\\Users\\name\\Documents")
	path := inputStr("Enter path: ", 150)

	// check the path is valid folder
	info, err := os.Stat(path)
	if err != nil {
		path, _ := os.UserHomeDir() // should check for error, but no home folder? Unlikely
		fmt.Println("Invalid path! Using home directory:", path)
	} else if !info.IsDir() {
		path, _ := os.UserHomeDir() // should check for error...
		fmt.Println("Invalid path! Using home directory:", path)
	}
	return path
}

// ReadTasksFile reads tasks from data file into taskList
func ReadTasksFile() {
	data, err := os.Open(config.filePath)
	if err != nil {
		fmt.Println("\nError opening '", config.filePath)
		fmt.Println()
		return
	}
	defer data.Close()

	taskList = nil // reset taskList
	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		str := strings.TrimSpace(scanner.Text())
		str = str[1 : len(str)-1] // Remove the leading and trailing quotes
		result := strings.Split(str, "\",\"")
		dueDate, _ := time.Parse("2006-01-02", result[1])
		taskList = append(taskList, Task{
			title:    result[0],
			due:      dueDate,
			priority: result[2],
			repeat:   result[3],
			label:    result[4],
			done:     result[5],
			notes:    result[6],
		})
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
}

// WriteTasksFile writes tasks from taskList to data file
func WriteTasksFile() {
	data, err := os.Create(config.folderPath + "/TaskManGo.txt")
	if err != nil {
		fmt.Println("Error creating file!")
		return
	}
	defer data.Close()

	writer := bufio.NewWriter(data)
	for _, task := range taskList {
		line := fmt.Sprintf("\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\"\n",
			task.title, task.due.Format("2006-01-02"), task.priority, task.repeat, task.label, task.done, task.notes)
		_, err := writer.WriteString(line)
		if err != nil {
			fmt.Println("Error writing to file!")
			return
		}
	}
	writer.Flush()
	fmt.Println("Tasks saved to:", config.filePath)
}

// Input helper functions
func inputStr(prompt string, length int) string { // input a string, limit length
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := strings.TrimSpace(scanner.Text())
	if len(text) > length {
		return text[:length]
	}
	return text
}

func yesNoInput(prompt string) string { // input yes/no, return "Yes" or "No"
	response := strings.ToLower(inputStr(prompt+" (y/n): ", 5))
	if response == "y" || response == "yes" {
		return "Yes"
	} else {
		return "No"
	}
}

func inputInt(prompt string, min int, max int) int { // input an integer within a range
	idx, err := strconv.Atoi(inputStr(prompt, 4))
	if err != nil || idx < min || idx > max {
		fmt.Printf("Number out of range (%d to %d).", min, max)
		return -9999 // negative indicates invalid input
	}
	return idx
}

// add a new Task to taskList
func addTask() {
	fmt.Println("\n----- Add new task -----")

	title := inputStr("Task title: ", 20)
	if title == "" {
		fmt.Println("Task title cannot be empty!")
		return
	}

	dueDate := inputStr("Due date (YYYY-MM-DD): ", 12)
	due, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		due, _ = time.Parse("2006-01-02", "2099-12-31")
	}
	due = time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, due.Location()) // set time to 00:00

	priority := inputStr("Priority (1, 2, 3): ", 3)
	if priority != "1" && priority != "2" {
		priority = "3" // default priority
	}

	repeat := strings.ToLower(inputStr("Repeat (d)aily, (w)eekly, (m)onthly: ", 10))
	switch repeat {
	case "d", "daily":
		repeat = "Daily"
	case "w", "weekly":
		repeat = "Weekly"
	case "m", "monthly":
		repeat = "Monthly"
	default:
		repeat = ""
	}

	label := inputStr("Label/category: ", 12)
	done := yesNoInput("Is the task done? ")
	notes := inputStr("Additional notes: ", 100)

	// Add the new task to the task list
	taskList = append(taskList, Task{
		title:    title,
		due:      due,
		priority: priority,
		repeat:   repeat,
		label:    label,
		done:     done,
		notes:    notes,
	})
}

// EditTask edits an existing task in taskList
func EditTask() {
	if len(taskList) == 0 {
		fmt.Println("No tasks to edit!")
		return
	}
	id := inputInt("Enter task ID to edit: ", 0, len(taskList)-1)
	if id < 0 {
		fmt.Println("Invalid task ID!")
		return
	}

	task := &taskList[id] // get pointer to the task to edit
	fmt.Println("\n----- Edit task -----")
	fmt.Println("1 Title:", task.title)
	due := task.due.Format("2006-01-02")
	if task.due.Equal(time.Date(2099, 12, 31, 0, 0, 0, 0, task.due.Location())) {
		due = ""
	}
	fmt.Println("2 Due date:", due)
	fmt.Println("3 Priority:", task.priority)
	fmt.Println("4 Repeat:", task.repeat)
	fmt.Println("5 Label:", task.label)
	fmt.Println("6 Done:", task.done)
	fmt.Println("7 Notes:", task.notes)
	choice := inputInt("\nNumber of field to edit (Enter cancels): ", 0, 7)

	switch choice {
	case 1:
		newTitle := inputStr("New title: ", 30)
		if newTitle != "" {
			task.title = newTitle
		}
	case 2:
		dueDate := inputStr("Due date (YYYY-MM-DD): ", 12)
		due, err := time.Parse("2006-01-02", dueDate)
		if err != nil {
			due, _ = time.Parse("2006-01-02", "2099-12-31")
		}
		task.due = time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, due.Location()) // set time to 00:00
	case 3:
		priority := inputStr("New priority (1, 2, 3): ", 3)
		if priority != "1" && priority != "2" {
			priority = "3" // default priority
		}
		task.priority = priority
	case 4:
		repeat := strings.ToLower(inputStr("New (d)aily, (w)eekly, (m)onthly: ", 10))
		switch repeat {
		case "d", "daily":
			task.repeat = "Daily"
		case "w", "weekly":
			task.repeat = "Weekly"
		case "m", "monthly":
			task.repeat = "Monthly"
		default:
			task.repeat = ""
		}
	case 5:
		task.label = inputStr("New label: ", 12)
	case 6:
		task.done = yesNoInput("Is the task done? ")
	case 7:
		task.notes = inputStr("Additional notes: ", 100)
	}
}

// ListTasks lists all tasks, optionally filtered by label
func ListTasks(filterBy string) {
	if len(taskList) == 0 {
		fmt.Println("No tasks found. Create one now!")
		return
	}
	fmt.Print("\033[H\033[2J") // clear the terminal screen
	PrintTitleHeader()
	for i, task := range taskList {
		if filterBy != "" && task.label != filterBy {
			continue
		}
		PrintTask(i, task)
	}
}

// PrintTitleHeader prints the header for the task list
func PrintTitleHeader() {
	fmt.Printf("\n%-3s", "ID")
	fmt.Printf("%-20s", "Title")
	fmt.Printf("%-12s", "Due")
	fmt.Printf("%-6s", "Prty")
	fmt.Printf("%-9s", "Repeat")
	fmt.Printf("%-12s", "Label")
	fmt.Printf("%-4s", "Done")
	fmt.Println()
	for range 70 {
		fmt.Print("-")
	}
	fmt.Println()
}

// PrintTask prints a single task with color coding
func PrintTask(i int, task Task) {
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()) // set time to 00:00

	if task.done == "Yes" {
		fmt.Print(Green) // highlight done tasks in green
	} else {
		if task.due.Before(today) {
			fmt.Print(Red) // highlight due/overdue tasks in red
		} else if task.due.Equal(today) {
			fmt.Print(Blue) // highlight tasks due today in yellow
		}
	}

	fmt.Printf("%-03d", i)
	fmt.Printf("%-20s", task.title)
	due := task.due.Format("2006-01-02")
	if due == "2099-12-31" {
		due = ""
	}
	fmt.Printf("%-12s", due)
	fmt.Printf(" %-5s", task.priority)
	fmt.Printf("%-10s", task.repeat)
	fmt.Printf("%-11s", task.label)
	fmt.Printf("%-5s", task.done)
	fmt.Println(Reset) // reset color
}

// RemoveTask removes a task from taskList by ID
func RemoveTask() string {
	if len(taskList) == 0 {
		return "No tasks to delete!"
	}
	id := inputInt("Enter task ID to delete: ", 0, len(taskList)-1)
	if id < 0 {
		return "Invalid task ID!"
	}
	if len(taskList) == 1 {
		taskList = nil
	} else {
		taskList = append(taskList[:id], taskList[id+1:]...)
	}
	return "Task deleted."
}

// DueTasks lists tasks that are due soon
func DueTasks() { // tasks due soon
	if len(taskList) == 0 {
		return
	}
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()) // set time to 00:00
	nextWeek := today.AddDate(0, 0, 3)                                                        // 3 days ahead
	nextWeek = time.Date(nextWeek.Year(), nextWeek.Month(), nextWeek.Day(), 0, 0, 0, 0, nextWeek.Location())

	flag := true // to print header only once
	for _, task := range taskList {
		if task.done == "Yes" {
			continue
		}
		if task.due.Before(today) || task.due.After(nextWeek) {
			continue
		}
		if flag {
			fmt.Println("\n-- Tasks Due soon ----")
			flag = false
		}
		// PrintTask(i, task)
		due := task.due.Format("2006-01-02")
		fmt.Printf("%s (%s), ", task.title, due)
	}
	if !flag {
		fmt.Println()
	}
}

// SortTasksByDueDate sorts taskList by due date
func SortTasksByDueDate() {
	sortFunc := func(x, y Task) int {
		return cmp.Compare(x.due.Format("2006-01-02"), y.due.Format("2006-01-02"))
	}
	taskList = slices.SortedStableFunc(slices.Values(taskList), sortFunc)
}

// SortTasksByPriority sorts taskList by priority
func SortTasksByPriority() {
	sortFunc := func(x, y Task) int {
		return cmp.Compare(x.priority, y.priority)
	}
	taskList = slices.SortedStableFunc(slices.Values(taskList), sortFunc)
}

// SortTasksByName sorts taskList by name
func SortTasksByName() {
	sortFunc := func(x, y Task) int {
		return cmp.Compare(x.title, y.title)
	}
	taskList = slices.SortedStableFunc(slices.Values(taskList), sortFunc)
}

// SortTasks prompts user for sort option and sorts taskList accordingly
func SortTasks() {
	s := inputStr("Sort by (n)ame, (p)riority, (d)ue: ", 5)
	switch strings.ToLower(s) {
	case "n", "name":
		SortTasksByName()
	case "p", "priority":
		SortTasksByPriority()
	case "d", "due":
		SortTasksByDueDate()
	default:
		fmt.Println("Invalid sort option!")
		return
	}
}

// UpdateRecurringTasks updates recurring tasks that are marked as done
func UpdateRecurringTasks() {
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location()) // set time to 00:00
	for i, task := range taskList {
		if task.done == "Yes" && task.repeat != "" {
			if task.due.Before(today) || task.due.Equal(today) {
				switch task.repeat {
				case "Daily":
					taskList[i].due = task.due.AddDate(0, 0, 1)
				case "Weekly":
					taskList[i].due = task.due.AddDate(0, 0, 7)
				case "Monthly":
					taskList[i].due = task.due.AddDate(0, 1, 0)
				}
				taskList[i].done = "No" // mark as not done
			}
		}
	}
}

// DoneTask marks a task as done by ID
func DoneTask() {
	if len(taskList) == 0 {
		fmt.Println("No tasks to mark as done!")
		return
	}
	id := inputInt("Enter task ID to mark as done: ", 0, len(taskList)-1)
	if id < 0 {
		fmt.Println("Invalid task ID!")
		return
	}
	taskList[id].done = "Yes"
	fmt.Println(taskList[id])
}

// main function - start here!
func main() {
	ReadConfig()
	ReadTasksFile()
	SortTasksByDueDate()
	fmt.Println()
	fmt.Print("\033[H\033[2J") // clear the terminal screen

	fmt.Println("TaskManGo Task Manager:")
	label := ""
	quit := false
	for !quit {
		UpdateRecurringTasks()
		ListTasks(label) // list tasks, filtered by label if set
		DueTasks()
		choice := strings.ToLower(inputStr("\nOptions: (a)dd, (e)dit, (d)one, (s)ort, (f)ilter, (r)emove, (q)uit? ", 5))
		switch choice {
		case "a", "add":
			addTask()
		case "e", "edit":
			EditTask()
		case "d", "done":
			DoneTask()
		case "s", "sort":
			SortTasks()
		case "f", "filter":
			label = inputStr("Enter label to filter by (leave empty for no filter): ", 12)
		case "r", "remove":
			fmt.Println(RemoveTask())
		case "q", "quit":
			quit = true
		}
	}
	WriteTasksFile()
}
