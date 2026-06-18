package incusui

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"github.com/MixTapeSoftware/inctui/internal/incusapi"
	"github.com/lxc/incus/shared/api"
	"log"
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
	s := "Select an Instance"

	for i, instance := range m.instances {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		//	checked := " "
		//	if _, ok := m.selected[i]; ok {
		//		checked = "x"
		//	}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, instance.Status, instance.Name)
	}

	s += "\nPress q to quit.\n"

	return tea.NewView(s)
}
