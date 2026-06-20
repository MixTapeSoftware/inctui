package incusui

import (
	"charm.land/lipgloss/v2/table"
	"fmt"
	"github.com/samber/lo"
	"log"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"
	"github.com/lxc/incus/shared/api"
	"golang.org/x/sync/errgroup"
)

type InstanceServer interface {
	Instances() ([]api.Instance, error)
	InstanceState(name string) (*api.InstanceState, error)
	ToggleInstance(name string, statusCode api.StatusCode)
}

type statesLookup map[string]api.InstanceState

type stateSnapshot struct {
	statesLookup statesLookup
	sampleTime   time.Time
}

type model struct {
	logger            *os.File
	server            InstanceServer
	instances         []api.Instance
	stateSnapshot     stateSnapshot
	lastStateSnapshot stateSnapshot
	cursor            int
	selected          map[int]struct{}
}

// Initialize and Run the Instances UI
func InstancesUI(server InstanceServer) (tea.Model, error) {
	p := tea.NewProgram(initialModel(server))
	return p.Run()
}

// ************************************************
// Bubbletea tea.Model Interface lifecyled hooks
// Init(), Update(), View()
// ************************************************

// We start a heartbeat to refresh live state data like CPU usage
func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tickMsg:
		m.instances = instances(m.server)
		names := instanceNames(m.instances)
		newStateSnapshot := getStateSnapshot(names, m.server)
		m.lastStateSnapshot = m.stateSnapshot
		m.stateSnapshot = newStateSnapshot
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

		case "enter", "space", "s":
			if len(m.instances) > 0 {
				selected := m.instances[m.cursor]
				m.server.ToggleInstance(selected.Name, selected.StatusCode)
			}
		}
	}
	return m, nil
}

func instanceNames(instances []api.Instance) []string {
	return lo.Map(instances, func(instance api.Instance, index int) string {
		return string(instance.Name)
	})
}

func (m model) View() tea.View {
	var b strings.Builder
	t := instancesTable()

	for i, instance := range m.instances {
		cursor := " "
		if m.cursor == i {
			cursor = "→"
		}

		appendRow(t, cursor, instance, m)
	}
	b.WriteString(t.String())
	b.WriteString("\n")
	b.WriteString(key())

	return tea.NewView(b.String())
}

func initialModel(server InstanceServer) model {
	f, _ := os.Create("debug.log")
	initialStates := map[string]api.InstanceState{}
	initialSample := stateSnapshot{statesLookup: initialStates, sampleTime: time.Time{}}
	return model{
		logger:            f,
		server:            server,
		lastStateSnapshot: initialSample,
		instances:         instances(server),
		selected:          make(map[int]struct{}),
	}
}

func instances(server InstanceServer) []api.Instance {
	instances, err := server.Instances()
	if err != nil {
		log.Fatal("Couldn't load Incus Instances")
	}
	return instances
}

func instancesTable() *table.Table {
	var (
		purple    = lipgloss.Color("99")
		gray      = lipgloss.Color("245")
		lightGray = lipgloss.Color("241")

		headerStyle  = lipgloss.NewStyle().Foreground(purple).Bold(true).Align(lipgloss.Center)
		cellStyle    = lipgloss.NewStyle().Padding(0, 1).Width(14)
		oddRowStyle  = cellStyle.Foreground(gray)
		evenRowStyle = cellStyle.Foreground(lightGray)
	)

	return table.New().
		Width(60).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(purple)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return headerStyle
			case row%2 == 0:
				return evenRowStyle
			default:
				return oddRowStyle
			}
		}).
		Headers("Instance", "CPU", "Memory")
}

func appendRow(t *table.Table, cursor string, instance api.Instance, m model) {
	indicator := StatusIndicator(instance.StatusCode)
	status := StatusText(instance.Status)
	cores := numCores(instance.ExpandedConfig["limits.cpu"])
	cpuPercent := fmt.Sprintf("%.2f%%", calcCPUPercent(instance.Name, m, cores))
	memory := calcMemory(instance.Name, m)
	t.Row(fmt.Sprintf("%s %s %s %s", cursor, indicator, instance.Name, status), cpuPercent, memory)
}

func calcMemory(name string, m model) string {
	state := m.stateSnapshot.statesLookup[name]
	return fmt.Sprintf("%vMB", state.Memory.Usage/(1024*1024))
}

// Calculates Percent of CPU over a period of time.
// Assuming 1 second == 100% cpu usage, and we have a CPu count, we can subtract
// samples of the nanoseconds used over a period of nanoseconds
// and get the % of CPUs uses
func calcCPUPercent(name string, m model, cores int) float64 {
	newTime := m.stateSnapshot.sampleTime
	lastTime := m.lastStateSnapshot.sampleTime
	newState := m.stateSnapshot.statesLookup[name]
	lastState := m.lastStateSnapshot.statesLookup[name]
	if lastTime.Equal(time.Time{}) {
		return 0
	}
	elapsedNanos := newTime.Sub(lastTime).Nanoseconds()
	return float64(newState.CPU.Usage-lastState.CPU.Usage) * 1000 / float64(elapsedNanos*int64(cores))

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

// Gets a runtime state snapshot
//
// Runtime state /1.0/instances/{name}/state must be fetched one at a time.
// We use goroutines to fetch them concurrently, populate a lookup map keyed
// by instance name and then wait until all requests have completed to rerturn
// them in a batch.
func getStateSnapshot(instanceNames []string, server InstanceServer) stateSnapshot {
	var eg errgroup.Group
	var mu sync.Mutex
	states := make(map[string]api.InstanceState, len(instanceNames))
	for _, name := range instanceNames {
		eg.Go(func() error {
			instanceState, err := server.InstanceState(name)
			if err != nil {
				return err
			}
			mu.Lock()
			states[name] = *instanceState
			mu.Unlock()
			return nil
		})
	}
	eg.Wait()

	return stateSnapshot{statesLookup: states, sampleTime: time.Now()}
}

func key() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Bold(true).
		PaddingLeft(2).
		Width(50).
		Render("start/stop with space; Press q to quit")

}
