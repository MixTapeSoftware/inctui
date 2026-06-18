package incusui

import (
	"fmt"
	"github.com/samber/lo"
	"log"
	"slices"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/MixTapeSoftware/inctui/internal/incusapi"
	"github.com/lxc/incus/client"
	"github.com/lxc/incus/shared/api"
	"golang.org/x/sync/errgroup"
)

func InstancesUI() (tea.Model, error) {
	p := tea.NewProgram(initialModel())
	return p.Run()
}

type model struct {
	instances      []api.Instance
	instanceStates map[string]api.InstanceState
	cursor         int
	selected       map[int]struct{}
}

func client() incus.InstanceServer {
	client, err := incusapi.NewClient()
	if err != nil {
		log.Fatal("Could connect to Incus")
	}
	return client
}

func instances() []api.Instance {
	client := client()
	instances, err := incusapi.Instances(client)
	if err != nil {
		log.Fatal("Couldn't load Incus Instances")
	}
	return instances

}

func initialModel() model {
	return model{
		instances: instances(),
		selected:  make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.instances = instances()
		names := lo.Map(m.instances, func(instance api.Instance, index int) string {
			return string(instance.Name)
		})
		m.instanceStates = getStates(names)
		return m, tick()
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
		row := toRow(cursor, instance, m.instanceStates[instance.Name])
		b.WriteString(row)
	}
	b.WriteString("\nPress q to quit.\n")

	return tea.NewView(b.String())
}

func toRow(cursor string, instance api.Instance, state api.InstanceState) string {
	indicator := StatusIndicator(instance.StatusCode)
	status := StatusText(instance.Status)
	return fmt.Sprintf("%s %v %s %s %v\n", cursor, indicator, instance.Name, status, state.CPU.Usage)
}

func StatusText(status string) string {
	displayableStates := []string{"Starting", "Freezing", "Frozen", "Thawed", "Error", "Pending", "Cancelling"}
	if slices.Contains(displayableStates, status) {
		return fmt.Sprintf("[%s]", status)
	} else {
		return ""
	}
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
			Foreground(lipgloss.BrightWhite).
			Render("○")
	case 103:
		return lipgloss.NewStyle().
			Foreground(lipgloss.BrightGreen).
			Render("●")
	default:
		return lipgloss.NewStyle().
			Render("○")

	}
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func getStates(instanceNames []string) map[string]api.InstanceState {
	var eg errgroup.Group
	client := client()
	states := make(map[string]api.InstanceState, len(instanceNames))
	for _, name := range instanceNames {
		eg.Go(func() error {
			instanceState, err := incusapi.InstanceState(client, name)
			states[name] = *instanceState
			return err
		})
		eg.Wait()
	}

	return states
}
