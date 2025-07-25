package tui

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/terzigolu/josepshbrain-go/internal/api"
	"github.com/terzigolu/josepshbrain-go/internal/config"
	"github.com/terzigolu/josepshbrain-go/internal/models"
)

// Run starts the terminal UI.
func Run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}

	client := api.NewClient()

	projectID := cfg.ActiveProjectID

	app := tview.NewApplication()
	list := tview.NewList()
	details := tview.NewTextView()
	details.SetBorder(true)
	details.SetTitle("Details")
	details.SetDynamicColors(true)

	loadTasks := func() ([]models.Task, error) {
		tasks, err := client.ListTasks(projectID, "")
		if err != nil {
			return nil, err
		}
		list.Clear()
		for _, t := range tasks {
			main := fmt.Sprintf("%s %s %s", priorityIcon(t.Priority), statusIcon(t.Status), truncateString(t.Title, 40))
			list.AddItem(main, t.ID.String()[:8], 0, nil)
		}
		return tasks, nil
	}

	tasks, err := loadTasks()
	if err != nil {
		return err
	}

	list.SetBorder(true).SetTitle("Tasks")
	list.SetChangedFunc(func(i int, main, secondary string, r rune) {
		if i < 0 || i >= len(tasks) {
			return
		}
		t := tasks[i]
		details.SetText(fmt.Sprintf("[yellow]ID:[white] %s\n[yellow]Title:[white] %s\n[yellow]Status:[white] %s\n[yellow]Priority:[white] %s\n\n%s", t.ID.String(), t.Title, t.Status, t.Priority, t.Description))
	})

	updateStatus := func(status string) {
		i := list.GetCurrentItem()
		if i < 0 || i >= len(tasks) {
			return
		}
		id := tasks[i].ID.String()
		if _, err := client.UpdateTask(id, map[string]interface{}{"status": status}); err == nil {
			tasks, _ = loadTasks()
			list.SetCurrentItem(i)
		}
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			app.Stop()
		case 'r':
			tasks, _ = loadTasks()
		case 'c':
			updateStatus("COMPLETED")
		case 's':
			updateStatus("IN_PROGRESS")
		}
		return event
	})

	flex := tview.NewFlex().AddItem(list, 0, 1, true).AddItem(details, 0, 1, false)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		return err
	}
	return nil
}

func statusIcon(status string) string {
	switch status {
	case "TODO":
		return "📋"
	case "IN_PROGRESS":
		return "🚀"
	case "IN_REVIEW":
		return "👀"
	case "COMPLETED":
		return "✅"
	}
	return "❓"
}

func priorityIcon(p string) string {
	switch p {
	case "H":
		return "🔴"
	case "M":
		return "🟡"
	case "L":
		return "🟢"
	}
	return "⚪"
}

func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
