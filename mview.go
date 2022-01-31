package main

// An example demonstrating an application with multiple views.
//
// Note that this example was produced before the Bubbles progress component
// was available (github.com/charmbracelet/bubbles/progress) and thus, we're
// implementing a progress bar from scratch here.

import (
	"fmt"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/termenv"
)

const (
	progressBarWidth  = 71
	progressFullChar  = "█"
	progressEmptyChar = "░"
)

// General stuff for styling the view
var (
	term          = termenv.ColorProfile()
	keyword       = makeFgStyle("211")
	subtle        = makeFgStyle("241")
	progressEmpty = subtle(progressEmptyChar)
	dot           = colorFg(" • ", "236")

	// Gradient colors we'll use for the progress bar
	ramp = makeRamp("#B14FFF", "#00FFA3", progressBarWidth)
)
var titleStyle = lipgloss.NewStyle().Padding(0, 4).Bold(true).Background(lipgloss.Color("#4169E1")).Foreground(lipgloss.Color("#00FA9A")).MarginBottom(1)

func main() {

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FA9A"))
	initialModel := model{0, false, 10, 0, 0, false, false, s}
	p := tea.NewProgram(initialModel)

	if err := p.Start(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

type tickMsg struct{}
type frameMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}

type model struct {
	Choice   int
	Chosen   bool
	Ticks    int
	Frames   int
	Progress float64
	Loaded   bool
	Quitting bool
	spinner  spinner.Model
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
	//return tick()
}

// Main update function.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Make sure these keys always quit
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.Quitting = true
			return m, tea.Quit
		}
	}

	// Hand off the message and model to the appropriate update function for the
	// appropriate view based on the current state.
	if !m.Chosen {
		return updateChoices(msg, m)
	}
	return updateChosen(msg, m)
}

// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.Quitting {
		return "\n  Até mais!\n\n"
	}
	if !m.Chosen {
		s = choicesView(m)
	} else {
		s = chosenView(m)
	}
	return indent.String("\n"+s+"\n\n", 2)
}

// Sub-update functions

// Update loop for the first view where you're choosing a task.
func updateChoices(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		case "down":
			m.Choice += 1
			if m.Choice > 3 {
				m.Choice = 3
			}
		case "up":
			m.Choice -= 1
			if m.Choice < 0 {
				m.Choice = 0
			}
		case "enter":
			m.Chosen = true
			return m, frame()
		}
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tickMsg:
		/*if m.Ticks == 0 {
			m.Quitting = true
			return m, tea.Quit
		}*/
		m.Ticks -= 1
		return m, nil
	}

	return m, nil
}

// Update loop for the second view after a choice has been made
func updateChosen(msg tea.Msg, m model) (tea.Model, tea.Cmd) {

	switch msg.(type) {

	case frameMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		if m.Loaded {
			/*if m.Ticks == 0 {
				m.Quitting = true
				return m, tea.Quit
			}*/
			m.Ticks -= 1
			return m, nil
		}
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// Sub-views

// The first view, where you're choosing a task
func choicesView(m model) string {
	c := m.Choice

	tpl := titleStyle.Render("Assistente de deploy")
	tpl += "%s\n\n"
	tpl += subtle("up/down: Selecionar") + dot + subtle("enter: Escolher") + dot + subtle("esc: Sair")

	choices := fmt.Sprintf(
		"\n%s\n%s",
		checkbox("Instalar dependências com composer", c == 0),
		checkbox("Executar migrate com seeders", c == 1),
	)

	return fmt.Sprintf(tpl, choices)
}
func worker(m model) {
	fmt.Print("working...")
	time.Sleep(time.Second * 3)
	fmt.Println("done")
	m.Quitting = true
	m.Choice = -1
}

// The second view, after a task has been chosen
func chosenView(m model) string {

	var msg string

	switch m.Choice {
	case 0:
		msg = fmt.Sprintf("Instalando dependências...")
		worker(m)
	case 1:
		msg = fmt.Sprintf("Executando migrate com seeders")
	}

	label := "Executando tarefa..."
	/*if m.Loaded {
		label = fmt.Sprintf("Downloaded. Exiting in %s seconds...", colorFg(strconv.Itoa(m.Ticks), "79"))
	}*/

	return msg + "\n\n" + label + m.spinner.View()
}

func checkbox(label string, checked bool) string {
	if checked {
		return colorFg("[x] "+label, "#00FA9A")
	}
	return colorFg("[ ] "+label, "#4169E1")
}

// Utils

// Color a string's foreground with the given value.
func colorFg(val, color string) string {
	return termenv.String(val).Foreground(term.Color(color)).String()
}

// Return a function that will colorize the foreground of a given string.
func makeFgStyle(color string) func(string) string {
	return termenv.Style{}.Foreground(term.Color(color)).Styled
}

// Color a string's foreground and background with the given value.
func makeFgBgStyle(fg, bg string) func(string) string {
	return termenv.Style{}.
		Foreground(term.Color(fg)).
		Background(term.Color(bg)).
		Styled
}

// Generate a blend of colors.
func makeRamp(colorA, colorB string, steps float64) (s []string) {
	cA, _ := colorful.Hex(colorA)
	cB, _ := colorful.Hex(colorB)

	for i := 0.0; i < steps; i++ {
		c := cA.BlendLuv(cB, i/steps)
		s = append(s, colorToHex(c))
	}
	return
}

// Convert a colorful.Color to a hexadecimal format compatible with termenv.
func colorToHex(c colorful.Color) string {
	return fmt.Sprintf("#%s%s%s", colorFloatToHex(c.R), colorFloatToHex(c.G), colorFloatToHex(c.B))
}

// Helper function for converting colors to hex. Assumes a value between 0 and
// 1.
func colorFloatToHex(f float64) (s string) {
	s = strconv.FormatInt(int64(f*255), 16)
	if len(s) == 1 {
		s = "0" + s
	}
	return
}
