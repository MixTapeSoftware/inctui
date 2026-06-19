package incusui

import (
	"fmt"
	"log"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"

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

type statesLookup map[string]api.InstanceState
type sampleState struct {
	statesLookup statesLookup
	sampleTime   time.Time
}
type model struct {
	instances       []api.Instance
	sampleState     sampleState
	lastSampleState sampleState
	cursor          int
	selected        map[int]struct{}
}

func newClient() incus.InstanceServer {
	client, err := incusapi.NewClient()
	if err != nil {
		log.Fatal("Could connect to Incus")
	}
	return client
}

func instances() []api.Instance {
	client := newClient()
	instances, err := incusapi.Instances(client)
	if err != nil {
		log.Fatal("Couldn't load Incus Instances")
	}
	return instances

}

func initialModel() model {
	initialStates := map[string]api.InstanceState{}
	initialSample := sampleState{statesLookup: initialStates, sampleTime: time.Time{}}
	return model{
		lastSampleState: initialSample,
		instances:       instances(),
		selected:        make(map[int]struct{}),
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
		newSampleState := getSampleState(names)
		m.lastSampleState = m.sampleState
		m.sampleState = newSampleState
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

		row := toRow(cursor, instance, m)
		b.WriteString(row)
	}
	b.WriteString("\nPress q to quit.\n")

	return tea.NewView(b.String())
}

func toRow(cursor string, instance api.Instance, m model) string {
	indicator := StatusIndicator(instance.StatusCode)
	status := StatusText(instance.Status)
	cores := numCores(instance.ExpandedConfig["limits.cpu"])
	cpuPercent := calcCPUPercent(instance.Name, m, cores)
	return fmt.Sprintf("%s %v %s %s %.2f%%\n", cursor, indicator, instance.Name, status, cpuPercent)
}

// Calculates Percent of CPU over a period of time.
// Assuming 1 second == 100% cpu usage, and we have a CPu count, we can subtract
// samples of the nanoseconds used over a period of nanoseconds
// and get the % of CPUs uses
func calcCPUPercent(name string, m model, cores int) float64 {
	lastTime := m.lastSampleState.sampleTime
	thisTime := m.sampleState.sampleTime
	state := m.sampleState.statesLookup[name]
	lastState := m.lastSampleState.statesLookup[name]
	if lastTime == (time.Time{}) {
		return 0
	}
	elapsedNanos := thisTime.Sub(lastTime).Nanoseconds()
	return float64(state.CPU.Usage-lastState.CPU.Usage) * 1000 / float64(elapsedNanos*int64(cores))

}

func numCores(configuredCPULimits string) int {

	configuredCores, err := strconv.Atoi(configuredCPULimits)

	if err != nil || configuredCores == 0 {
		return runtime.NumCPU()

	} else {
		return configuredCores
	}

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
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func getSampleState(instanceNames []string) sampleState {
	var eg errgroup.Group
	var mu sync.Mutex
	client := newClient()
	states := make(map[string]api.InstanceState, len(instanceNames))
	for _, name := range instanceNames {
		eg.Go(func() error {
			instanceState, err := incusapi.InstanceState(client, name)
			mu.Lock()
			states[name] = *instanceState
			mu.Unlock()
			return err
		})
	}
	eg.Wait()

	return sampleState{statesLookup: states, sampleTime: time.Now()}
}
