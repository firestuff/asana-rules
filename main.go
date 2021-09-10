package main

import "fmt"

import "github.com/firestuff/asana-rules/asana"

func main() {
	a := asana.NewClientFromEnv()

	me, err := a.GetMe()
	if err != nil {
		panic(err)
	}

	fmt.Printf("User: %s\n", me)

	wrk, err := a.GetWorkspace()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Workspace: %s\n", wrk)

	utl, err := a.GetUserTaskList(me.GID, wrk.GID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("User Task List: %s\n", utl)

	secs, err := a.GetSections(utl.GID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Sections:\n")
	for _, sec := range secs {
		fmt.Printf("\t%s\n", sec)

		tasks, err := a.GetTasksFromSection(sec.GID)
		if err != nil {
			panic(err)
		}

		for _, task := range tasks {
			fmt.Printf("\t\t%s\n", task)
		}
	}
}
