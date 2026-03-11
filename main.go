package main

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type inputType int

const (
	inputTypeURL inputType = iota
	inputTypeDuration
	inputTypeClients
	inputTypeThreads
	inputTypeMaxStreams
	inputTypeHeader
)

type textInput struct {
	textinput.Model
	Type inputType
}

type model struct {
	ok     bool
	active int
	inputs []textInput
	useH1  bool
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit

		case tea.KeyEnter:
			m.ok = true
			return m, tea.Quit

		case tea.KeyShiftTab, tea.KeyUp:
			m.prev()

		case tea.KeyTab, tea.KeyDown:
			m.next()

		case tea.KeyCtrlH:
			m.addHeader()

		case tea.KeyCtrlX:
			m.remove()

		case tea.KeyCtrlO:
			m.useH1 = !m.useH1
		}
	}

	m.inputs[m.active].Model, cmd = m.inputs[m.active].Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s string

	url := m.filteredInputs(inputTypeURL)[0]
	dur := m.filteredInputs(inputTypeDuration)[0]
	clients := m.filteredInputs(inputTypeClients)[0]
	threads := m.filteredInputs(inputTypeThreads)[0]
	maxStreams := m.filteredInputs(inputTypeMaxStreams)[0]

	s += "h2load " + url.View() + "\n"
	if m.useH1 {
		s += "  --h1 (use http/1.1)\n"
	}
	s += "    -D " + dur.View() + " (duration)\n"
	s += "    -c " + clients.View() + " (clients)\n"
	s += "    -t " + threads.View() + " (threads)\n"
	s += "    -m " + maxStreams.View() + " (max concurrent streams)\n"

	for _, in := range m.filteredInputs(inputTypeHeader) {
		s += "    -H " + in.View() + "\n"
	}

	s += "\n<ctrl-h> add header | <ctrl-x> rm header | <ctrl-o> toggle h1 | <enter> build"

	return s + "\n"
}

func (m model) filteredInputs(t inputType) []textInput {
	var inputs []textInput
	for _, in := range m.inputs {
		if in.Type == t {
			inputs = append(inputs, in)
		}
	}
	return inputs
}

func (m *model) next() {
	m.active++
	if m.active > len(m.inputs)-1 {
		m.active = 0
	}
	m.focus()
}

func (m *model) prev() {
	m.active--
	if m.active < 0 {
		m.active = len(m.inputs) - 1
	}
	m.focus()
}

func (m model) focus() {
	for i := range m.inputs {
		if i == m.active {
			m.inputs[i].Focus()
		} else {
			m.inputs[i].Blur()
		}
	}
}

func (m *model) addHeader() {
	// Add the new header to the end of the inputs list
	m.inputs = append(m.inputs, newTextInput(inputTypeHeader, "Accept-Encoding: gzip"))
	m.active = len(m.inputs) - 1
	m.focus()
}

func (m *model) remove() {
	// Prevent removing core required fields
	if m.inputs[m.active].Type != inputTypeHeader {
		return
	}

	m.inputs = append(m.inputs[:m.active], m.inputs[m.active+1:]...)

	if m.active > len(m.inputs)-1 {
		m.active = len(m.inputs) - 1
	}
	m.focus()
}

func initModel() model {
	m := model{}

	m.inputs = append(m.inputs, newTextInput(inputTypeURL, "https://localhost:8443"))

	inDur := newTextInput(inputTypeDuration, "10s")
	m.inputs = append(m.inputs, inDur)

	inCli := newTextInput(inputTypeClients, "100")
	m.inputs = append(m.inputs, inCli)

	inThr := newTextInput(inputTypeThreads, "1")
	m.inputs = append(m.inputs, inThr)

	inMaxStr := newTextInput(inputTypeMaxStreams, "100")
	m.inputs = append(m.inputs, inMaxStr)

	m.inputs[0].Focus()
	return m
}

func newTextInput(t inputType, placeholder string) textInput {
	in := textInput{Type: t, Model: textinput.New()}
	in.Prompt = ""
	in.Placeholder = placeholder
	return in
}

func build(m model) string {
	var (
		url        = m.filteredInputs(inputTypeURL)[0]
		dur        = m.filteredInputs(inputTypeDuration)[0]
		clients    = m.filteredInputs(inputTypeClients)[0]
		threads    = m.filteredInputs(inputTypeThreads)[0]
		maxStreams = m.filteredInputs(inputTypeMaxStreams)[0]
		headers    = m.filteredInputs(inputTypeHeader)
	)

	var cmd string
	cmd += "h2load"

	if dur.Value() != "" {
		cmd += " -D " + dur.Value()
	}
	if clients.Value() != "" {
		cmd += " -c " + clients.Value()
	}
	if threads.Value() != "" {
		cmd += " -t " + threads.Value()
	}
	if maxStreams.Value() != "" {
		cmd += " -m " + maxStreams.Value()
	}
	if m.useH1 {
		cmd += " --h1"
	}

	for _, h := range headers {
		if h.Value() != "" {
			cmd += " -H '" + h.Value() + "'"
		}
	}

	target := url.Value()
	if target == "" {
		target = url.Placeholder
	}
	cmd += " " + target

	return cmd
}

func pastecmd(s string) {
	cbs, err := syscall.ByteSliceFromString(s)
	if err != nil {
		panic(err)
	}
	for _, c := range cbs {
		syscall.RawSyscall(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCSTI, uintptr(unsafe.Pointer(&c)))
	}
	fmt.Print("\r \r")
	os.Exit(0)
}

func main() {
	r, err := tea.NewProgram(initModel()).StartReturningModel()
	if err != nil {
		panic(err)
	}
	if m := r.(model); m.ok {
		fmt.Print("\n")
		pastecmd(build(m))
	}
}
