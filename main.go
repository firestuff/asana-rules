package main

import "time"

import . "github.com/firestuff/asana-rules/asanarules"

func main() {
	Every(5 * time.Second).
    InWorkspace("flamingcow.io").
    InMyTasksSections("Recently Assigned").
    OnlyIncomplete().
    DueInDays(0).
    PrintTasks()

	Loop()
}
