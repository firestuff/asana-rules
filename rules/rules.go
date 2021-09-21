package rules

import "bytes"
import "fmt"
import "math/rand"
import "strings"
import "time"

import "cloud.google.com/go/civil"
import "github.com/firestuff/automana/client"
import "golang.org/x/net/html"
import "golang.org/x/net/html/atom"

type queryMutator func(*client.WorkspaceClient, *client.SearchQuery) error
type taskActor func(*client.WorkspaceClient, *client.Task) error
type taskFilter func(*client.WorkspaceClient, *client.Task) (bool, error)
type workspaceClientGetter func(*client.Client) (*client.WorkspaceClient, error)

type periodic struct {
	period int
	done   chan bool

	workspaceClientGetter workspaceClientGetter
	queryMutators         []queryMutator
	taskFilters           []taskFilter
	taskActors            []taskActor
}

var periodics = []*periodic{}

func EverySeconds(seconds int) *periodic {
	ret := &periodic{
		period: seconds,
		done:   make(chan bool),
	}

	periodics = append(periodics, ret)

	return ret
}

func Loop() {
	c := client.NewClientFromEnv()

	for _, periodic := range periodics {
		periodic.start(c)
	}

	for _, periodic := range periodics {
		periodic.wait()
	}
}

func (p *periodic) InWorkspace(name string) *periodic {
	if p.workspaceClientGetter != nil {
		panic("Multiple calls to InWorkspace()")
	}

	p.workspaceClientGetter = func(c *client.Client) (*client.WorkspaceClient, error) {
		return c.InWorkspace(name)
	}

	return p
}

// Query mutators
func (p *periodic) InMyTasksSections(names ...string) *periodic {
	p.queryMutators = append(p.queryMutators, func(wc *client.WorkspaceClient, q *client.SearchQuery) error {
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
	p.queryMutators = append(p.queryMutators, func(wc *client.WorkspaceClient, q *client.SearchQuery) error {
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
	p.queryMutators = append(p.queryMutators, func(wc *client.WorkspaceClient, q *client.SearchQuery) error {
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
	p.queryMutators = append(p.queryMutators, func(wc *client.WorkspaceClient, q *client.SearchQuery) error {
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
	p.queryMutators = append(p.queryMutators, func(wc *client.WorkspaceClient, q *client.SearchQuery) error {
		if q.Completed != nil {
			return fmt.Errorf("Multiple clauses set Completed")
		}

		q.Completed = client.FALSE
		return nil
	})

	return p
}

func (p *periodic) OnlyComplete() *periodic {
	p.queryMutators = append(p.queryMutators, func(wc *client.WorkspaceClient, q *client.SearchQuery) error {
		if q.Completed != nil {
			return fmt.Errorf("Multiple clauses set Completed")
		}

		q.Completed = client.TRUE
		return nil
	})

	return p
}

func (p *periodic) WithTagsAnyOf(names ...string) *periodic {
	p.queryMutators = append(p.queryMutators, func(wc *client.WorkspaceClient, q *client.SearchQuery) error {
		if len(q.TagsAny) > 0 {
			return fmt.Errorf("Multiple clauses set TagsAny")
		}

		tagsByName, err := wc.GetTagsByName()
		if err != nil {
			return err
		}

		for _, name := range names {
			tag, found := tagsByName[name]
			if !found {
				return fmt.Errorf("Tag '%s' not found", name)
			}

			q.TagsAny = append(q.TagsAny, tag)
		}

		return nil
	})

	return p
}

func (p *periodic) WithoutTagsAnyOf(names ...string) *periodic {
	p.queryMutators = append(p.queryMutators, func(wc *client.WorkspaceClient, q *client.SearchQuery) error {
		if len(q.TagsNot) > 0 {
			return fmt.Errorf("Multiple clauses set TagsNot")
		}

		tagsByName, err := wc.GetTagsByName()
		if err != nil {
			return err
		}

		for _, name := range names {
			tag, found := tagsByName[name]
			if !found {
				return fmt.Errorf("Tag '%s' not found", name)
			}

			q.TagsNot = append(q.TagsNot, tag)
		}

		return nil
	})

	return p
}

// Task filters
func (p *periodic) WithUnlinkedURL() *periodic {
	p.taskFilters = append(p.taskFilters, func(wc *client.WorkspaceClient, t *client.Task) (bool, error) {
		return hasUnlinkedURL(t.ParsedHTMLNotes), nil
	})

	return p
}

func (p *periodic) WithoutDue() *periodic {
	// We can't mutate the query because due_on=null is buggy in the Asana API
	p.taskFilters = append(p.taskFilters, func(wc *client.WorkspaceClient, t *client.Task) (bool, error) {
		return t.ParsedDueOn == nil, nil
	})

	return p
}

// Task actors
func (p *periodic) FixUnlinkedURL() *periodic {
	p.taskActors = append(p.taskActors, func(wc *client.WorkspaceClient, t *client.Task) error {
		fixUnlinkedURL(t.ParsedHTMLNotes)

		buf := &bytes.Buffer{}

		err := html.Render(buf, t.ParsedHTMLNotes)
		if err != nil {
			return err
		}

		notes := buf.String()

		update := &client.Task{
			GID:       t.GID,
			HTMLNotes: strings.TrimSuffix(strings.TrimPrefix(notes, "<html><head></head>"), "</html>"),
		}

		return wc.UpdateTask(update)
	})

	return p
}

func (p *periodic) MoveToMyTasksSection(name string) *periodic {
	p.taskActors = append(p.taskActors, func(wc *client.WorkspaceClient, t *client.Task) error {
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
	p.taskActors = append(p.taskActors, func(wc *client.WorkspaceClient, t *client.Task) error {
		fmt.Printf("%s\n", t)
		return nil
	})

	return p
}

// Infra
func (p *periodic) start(client *client.Client) {
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

func (p *periodic) loop(client *client.Client) {
	for {
		time.Sleep(time.Duration(rand.Intn(p.period)) * time.Second)

		err := p.exec(client)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			// continue
		}
	}

	close(p.done)
}

func (p *periodic) exec(c *client.Client) error {
	wc, err := p.workspaceClientGetter(c)
	if err != nil {
		return err
	}

	q := &client.SearchQuery{}

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

	filteredTasks := []*client.Task{}
	for _, task := range tasks {
		included := true

		for _, filter := range p.taskFilters {
			include, err := filter(wc, task)
			if err != nil {
				return err
			}

			if !include {
				included = false
				break
			}
		}

		if included {
			filteredTasks = append(filteredTasks, task)
		}
	}

	for _, task := range filteredTasks {
		for _, act := range p.taskActors {
			err = act(wc, task)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Helpers
func fixUnlinkedURL(node *html.Node) {
	if node == nil {
		return
	}

	if node.Type == html.ElementNode && node.Data == "a" {
		// Don't go down this tree, since it's a link
		return
	}

	fixUnlinkedURL(node.FirstChild)
	fixUnlinkedURL(node.NextSibling)

	if nodeHasUnlinkedURL(node) {
		next := node.NextSibling

		splitTextNodes(node)

		for iter := node; iter != next; iter = iter.NextSibling {
			if !nodeHasUnlinkedURL(iter) {
				continue
			}

			iter.FirstChild = &html.Node{
				Type: html.TextNode,
				Data: iter.Data,
			}

			iter.Type = html.ElementNode
			iter.Attr = append(iter.Attr, html.Attribute{
				Key: "href",
				Val: iter.Data,
			})
			iter.Data = "a"
			iter.DataAtom = atom.A
		}
	}
}

func splitTextNodes(node *html.Node) {
	prev := node
	lines := strings.Split(node.Data, "\n")
	node.Data = ""

	for i, line := range lines {
		prev.NextSibling = &html.Node{
			Type:        html.TextNode,
			Data:        line,
			NextSibling: prev.NextSibling,
		}
		prev = prev.NextSibling

		if i == len(lines)-1 {
			// No newline after last line
			break
		}

		prev.NextSibling = &html.Node{
			Type:        html.TextNode,
			Data:        "\n",
			NextSibling: prev.NextSibling,
		}
		prev = prev.NextSibling
	}
}

func hasUnlinkedURL(node *html.Node) bool {
	if node == nil {
		return false
	}

	if node.Type == html.ElementNode && node.Data == "a" {
		// Don't go down this tree, since it's a link
		return false
	}

	if nodeHasUnlinkedURL(node) {
		return true
	}

	if hasUnlinkedURL(node.FirstChild) {
		return true
	}

	if hasUnlinkedURL(node.NextSibling) {
		return true
	}

	return false
}

func nodeHasUnlinkedURL(node *html.Node) bool {
	if node.Type == html.TextNode {
		for _, line := range strings.Split(node.Data, "\n") {
			if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
				return true
			}
		}
	}

	return false
}
