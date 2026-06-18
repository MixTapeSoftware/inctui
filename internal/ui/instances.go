package incusui

import (
	"fmt"
	"log"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/MixTapeSoftware/inctui/internal/incusapi"
	"github.com/lxc/incus/shared/api"
)

func InstancesUI() (tea.Model, error) {
	p := tea.NewProgram(initialModel())
	return p.Run()
}

type model struct {
	instances []api.Instance
	cursor    int
	selected  map[int]struct{}
}

func initialModel() model {
	client, err := incusapi.NewClient()
	if err != nil {
		log.Fatal("Could connect to Incus")
	}
	instances, err := incusapi.Instances(client)
	if err != nil {
		log.Fatal("Couldn't load Incus Instances")
	}
	return model{
		instances: instances,
		selected:  make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.instances)-1 {
				m.cursor++
			}

		case "enter", "space":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}
	}
	return m, nil
}

func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString("Select an Instance \n")

	for i, instance := range m.instances {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		//	checked := " "
		//	if _, ok := m.selected[i]; ok {
		//		checked = "x"
		//	}
		row := toRow(cursor, instance)
		b.WriteString(row)
	}
	b.WriteString("\nPress q to quit.\n")

	return tea.NewView(b.String())
}

func toRow(cursor string, instance api.Instance) string {
	indicator := StatusIndicator(instance.StatusCode)
	return fmt.Sprintf("%s %v %s [%s]\n", cursor, indicator, instance.Name, instance.Status)
}

// Components

func StatusIndicator(code api.StatusCode) string {
	//  OperationCreated StatusCode = 100
	//	Started          StatusCode = 101
	//	Stopped          StatusCode = 102
	//	Running          StatusCode = 103
	//	Cancelling       StatusCode = 104
	//	Pending          StatusCode = 105
	//	Starting         StatusCode = 106
	//	Stopping         StatusCode = 107
	//	Aborting         StatusCode = 108
	//	Freezing         StatusCode = 109
	//	Frozen           StatusCode = 110
	//	Thawed           StatusCode = 111
	//	Error            StatusCode = 112
	//	Ready            StatusCode = 113

	switch code {
	case 101:
		return lipgloss.NewStyle().
			Foreground(lipgloss.BrightYellow).
			Render("●")

	case 102:
		return lipgloss.NewStyle().
			Foreground(lipgloss.BrightRed).
			Render("●")
	case 103:
		return lipgloss.NewStyle().
			Foreground(lipgloss.BrightGreen).
			Render("●")
	default:
		return lipgloss.NewStyle().
			Render("○")

	}
}
