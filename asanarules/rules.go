package asanarules

import "fmt"
import "time"

import "cloud.google.com/go/civil"
import "github.com/firestuff/asana-rules/asanaclient"

type queryMutator func(*asanaclient.WorkspaceClient, *asanaclient.SearchQuery) error
type taskActor func(*asanaclient.WorkspaceClient, *asanaclient.Task) error
type workspaceClientGetter func(*asanaclient.Client) (*asanaclient.WorkspaceClient, error)

type periodic struct {
	duration time.Duration
	done     chan bool

	workspaceClientGetter workspaceClientGetter
	queryMutators         []queryMutator
	taskActors            []taskActor
}

var periodics = []*periodic{}

func EverySeconds(seconds int) *periodic {
	ret := &periodic{
		duration: time.Duration(seconds) * time.Second,
		done:     make(chan bool),
	}

	periodics = append(periodics, ret)

	return ret
}

func Loop() {
	client := asanaclient.NewClientFromEnv()

	for _, periodic := range periodics {
		periodic.start(client)
	}

	for _, periodic := range periodics {
		periodic.wait()
	}
}

func (p *periodic) InWorkspace(name string) *periodic {
	if p.workspaceClientGetter != nil {
		panic("Multiple calls to InWorkspace()")
	}

	p.workspaceClientGetter = func(c *asanaclient.Client) (*asanaclient.WorkspaceClient, error) {
		return c.InWorkspace(name)
	}

	return p
}

// Query mutators
func (p *periodic) InMyTasksSections(names ...string) *periodic {
	p.queryMutators = append(p.queryMutators, func(wc *asanaclient.WorkspaceClient, q *asanaclient.SearchQuery) error {
		utl, err := wc.GetMyUserTaskList()
		if err != nil {
			return err
		}

		secsByName, err := wc.GetSectionsByName(utl)
		if err != nil {
			return err
		}

		for _, name := range names {
			sec, found := secsByName[name]
			if !found {
				return fmt.Errorf("Section '%s' not found", name)
			}

			q.SectionsAny = append(q.SectionsAny, sec)
		}

		return nil
	})

	return p
}

func (p *periodic) DueInDays(days int) *periodic {
	p.queryMutators = append(p.queryMutators, func(wc *asanaclient.WorkspaceClient, q *asanaclient.SearchQuery) error {
		if q.DueOn != nil {
			return fmt.Errorf("Multiple clauses set DueOn")
		}

		d := civil.DateOf(time.Now())
		d = d.AddDays(days)
		q.DueOn = &d
		return nil
	})

	return p
}

func (p *periodic) DueInAtLeastDays(days int) *periodic {
	p.queryMutators = append(p.queryMutators, func(wc *asanaclient.WorkspaceClient, q *asanaclient.SearchQuery) error {
		if q.DueAfter != nil {
			return fmt.Errorf("Multiple clauses set DueAfter")
		}

		d := civil.DateOf(time.Now())
		d = d.AddDays(days)
		q.DueAfter = &d
		return nil
	})

	return p
}

func (p *periodic) DueInAtMostDays(days int) *periodic {
	p.queryMutators = append(p.queryMutators, func(wc *asanaclient.WorkspaceClient, q *asanaclient.SearchQuery) error {
		if q.DueBefore != nil {
			return fmt.Errorf("Multiple clauses set DueBefore")
		}

		d := civil.DateOf(time.Now())
		d = d.AddDays(days)
		q.DueBefore = &d
		return nil
	})

	return p
}

func (p *periodic) OnlyIncomplete() *periodic {
	p.queryMutators = append(p.queryMutators, func(wc *asanaclient.WorkspaceClient, q *asanaclient.SearchQuery) error {
		if q.Completed != nil {
			return fmt.Errorf("Multiple clauses set Completed")
		}

		q.Completed = asanaclient.FALSE
		return nil
	})

	return p
}

func (p *periodic) OnlyComplete() *periodic {
	p.queryMutators = append(p.queryMutators, func(wc *asanaclient.WorkspaceClient, q *asanaclient.SearchQuery) error {
		if q.Completed != nil {
			return fmt.Errorf("Multiple clauses set Completed")
		}

		q.Completed = asanaclient.TRUE
		return nil
	})

	return p
}

// Task actors
func (p *periodic) MoveToMyTasksSection(name string) *periodic {
	p.taskActors = append(p.taskActors, func(wc *asanaclient.WorkspaceClient, t *asanaclient.Task) error {
		utl, err := wc.GetMyUserTaskList()
		if err != nil {
			return err
		}

		sec, err := wc.GetSectionByName(utl, name)
		if err != nil {
			return err
		}

		return wc.AddTaskToSection(t, sec)
	})

	return p
}

func (p *periodic) PrintTasks() *periodic {
	p.taskActors = append(p.taskActors, func(wc *asanaclient.WorkspaceClient, t *asanaclient.Task) error {
		fmt.Printf("%s\n", t)
		return nil
	})

	return p
}

func (p *periodic) start(client *asanaclient.Client) {
	err := p.validate()
	if err != nil {
		panic(err)
	}

	go p.loop(client)
}

func (p *periodic) validate() error {
	return nil
}

func (p *periodic) wait() {
	<-p.done
}

func (p *periodic) loop(client *asanaclient.Client) {
	ticker := time.NewTicker(p.duration)

	for {
		<-ticker.C
		err := p.exec(client)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			// continue
		}
	}

	close(p.done)
}

func (p *periodic) exec(c *asanaclient.Client) error {
	wc, err := p.workspaceClientGetter(c)
	if err != nil {
		return err
	}

	q := &asanaclient.SearchQuery{}

	for _, mut := range p.queryMutators {
		err = mut(wc, q)
		if err != nil {
			return err
		}
	}

	tasks, err := wc.Search(q)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		for _, act := range p.taskActors {
			err = act(wc, task)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
