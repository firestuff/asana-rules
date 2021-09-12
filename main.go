package main

import . "github.com/firestuff/asana-rules/asanarules"

func main() {
	EverySeconds(30).
		InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Upcoming", "Later", "Someday").
		OnlyIncomplete().
		DueInDays(0).
		PrintTasks().
		MoveToMyTasksSection("Today")

	EverySeconds(30).
		InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Meetings", "Maybe Today", "Tonight", "Later", "Someday").
		OnlyIncomplete().
		DueInAtLeastDays(1).
		DueInAtMostDays(7).
		PrintTasks().
		MoveToMyTasksSection("Upcoming")

	EverySeconds(30).
		InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Meetings", "Maybe Today", "Tonight", "Upcoming", "Someday").
		OnlyIncomplete().
		DueInAtLeastDays(8).
		PrintTasks().
		MoveToMyTasksSection("Later")

	EverySeconds(30).
		InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Meetings", "Maybe Today", "Tonight", "Upcoming", "Someday").
		OnlyIncomplete().
		DueInAtLeastDays(8).
		PrintTasks().
		MoveToMyTasksSection("Later")

	EverySeconds(30).
		InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Meetings", "Maybe Today", "Tonight", "Upcoming", "Later").
		OnlyIncomplete().
		WithoutDue().
		PrintTasks().
		MoveToMyTasksSection("Someday")

	Loop()
}
