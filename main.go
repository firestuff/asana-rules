package main

import . "github.com/firestuff/asana-rules/asanarules"

func main() {
	EverySeconds(30).
		InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Upcoming", "Later").
		OnlyIncomplete().
		DueInDays(0).
		PrintTasks().
		MoveToMyTasksSection("Today")

	EverySeconds(30).
		InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Later").
		OnlyIncomplete().
		DueInAtLeastDays(1).
		DueInAtMostDays(7).
		PrintTasks().
		MoveToMyTasksSection("Upcoming")

	EverySeconds(30).
		InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Upcoming").
		OnlyIncomplete().
		DueInAtLeastDays(8).
		PrintTasks().
		MoveToMyTasksSection("Later")

	Loop()
}
