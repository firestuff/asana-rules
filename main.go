package main

import . "github.com/firestuff/asana-rules/asanarules"

func main() {
	EverySeconds(30).
		InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned").
		OnlyIncomplete().
		DueInDays(0).
		PrintTasks().
		MoveToMyTasksSection("Today")

	EverySeconds(30).
		InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned").
		OnlyIncomplete().
		DueInAtLeastDays(1).
		DueInAtMostDays(7).
		PrintTasks().
		MoveToMyTasksSection("Upcoming")

	Loop()
}
