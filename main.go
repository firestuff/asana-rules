package main

import . "github.com/firestuff/automana/rules"

func main() {
	InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Meetings", "Tonight", "Upcoming", "Later", "Someday").
		OnlyIncomplete().
		DueInDays(0).
		WithoutTagsAnyOf("section=Tonight", "section=Meetings").
		PrintTasks().
		MoveToMyTasksSection("Today")

	InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Meetings", "Maybe Today", "Upcoming", "Later", "Someday").
		OnlyIncomplete().
		DueInDays(0).
		WithTagsAnyOf("section=Tonight").
		PrintTasks().
		MoveToMyTasksSection("Tonight")

	InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Maybe Today", "Tonight", "Upcoming", "Later", "Someday").
		OnlyIncomplete().
		DueInDays(0).
		WithTagsAnyOf("section=Meetings").
		PrintTasks().
		MoveToMyTasksSection("Meetings")

	InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Meetings", "Maybe Today", "Tonight", "Later", "Someday").
		OnlyIncomplete().
		DueInAtLeastDays(1).
		DueInAtMostDays(7).
		PrintTasks().
		MoveToMyTasksSection("Upcoming")

	InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Meetings", "Maybe Today", "Tonight", "Upcoming", "Someday").
		OnlyIncomplete().
		DueInAtLeastDays(8).
		PrintTasks().
		MoveToMyTasksSection("Later")

	InWorkspace("flamingcow.io").
		InMyTasksSections("Today", "Meetings", "Tonight", "Upcoming", "Later").
		OnlyIncomplete().
		WithoutDue().
		PrintTasks().
		MoveToMyTasksSection("Someday")

	InWorkspace("flamingcow.io").
		InMyTasksSections("Recently Assigned", "Today", "Meetings", "Maybe Today", "Tonight", "Upcoming", "Later", "Someday").
		OnlyIncomplete().
		WithUnlinkedURL().
		PrintTasks().
		FixUnlinkedURL()

	Loop()
}
