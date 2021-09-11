package main

// import "fmt"
import "time"

import . "github.com/firestuff/asana-rules/asanarules"

// import "github.com/firestuff/asana-rules/asanaclient"

func main() {
	Every(5 * time.Second).
		MyTasks()

	Loop()
	/*
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

		utl, err := a.GetUserTaskList(me, wrk)
		if err != nil {
			panic(err)
		}

		fmt.Printf("User Task List: %s\n", utl)

		secs, err := a.GetSections(utl)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Sections:\n")
		for _, sec := range secs {
			fmt.Printf("\t%s\n", sec)

			if sec.Name != "Recently Assigned" {
				continue
			}

			q := &asana.SearchQuery{
				SectionsAny: []*asana.Section{sec},
				Completed:   asana.FALSE,
			}

			tasks, err := a.Search(wrk, q)
			if err != nil {
				panic(err)
			}

			for _, task := range tasks {
				fmt.Printf("\t\t%s\n", task)
				a.AddTaskToSection(task, &asana.Section{
					GID: "1200372179004456",
				})
			}
		}
	*/
}
