package main

import (
	"fmt"
	"os"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	appNameStyle = lipgloss.NewStyle().Background(lipgloss.Color("110")).Padding(0, 1)
	faintStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Faint(true)
)

const (
	layoutPage uint = iota
	authPage
	homePage
)

type Form struct {
	email   textinput.Model
	confirm textinput.Model
	pass    textinput.Model
}

type model struct {
	page    uint
	choices []string
	cursor  int
	form    Form
}

func (m model) Init() tea.Cmd {
	return nil
}

func initialModel() model {
	email := textinput.New()
	email.Placeholder = "email"
	email.Focus() // курсор стартует здесь
	email.SetWidth(30)

	pass := textinput.New()
	pass.Placeholder = "password"
	pass.EchoMode = textinput.EchoPassword
	pass.SetWidth(30)

	confirm := textinput.New()
	confirm.Placeholder = "confirm"
	confirm.Focus()
	confirm.SetWidth(30)

	return model{
		page:    layoutPage,
		choices: []string{"register", "login"},
		form: Form{
			email:   email,
			pass:    pass,
			confirm: confirm,
		},
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			m.page = layoutPage
			return m, nil

		case "tab":
			if m.page == authPage {
				if m.form.email.Focused() {
					m.form.email.Blur()
					m.form.pass.Focus()
				} else {
					m.form.pass.Blur()
					m.form.email.Focus()
				}
			}
			return m, nil
		}

		if m.page == layoutPage {
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
				return m, nil
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
				return m, nil
			case "enter", "space":
				m.form.email.SetValue("")
				m.form.pass.SetValue("")
				m.page = authPage
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	if m.form.email.Focused() {
		m.form.email, cmd = m.form.email.Update(msg)
	} else if m.form.pass.Focused() {
		m.form.pass, cmd = m.form.pass.Update(msg)
	}
	return m, cmd
}

func (m model) View() tea.View {
	s := appNameStyle.Render("Go Keeper") + "\n\n"

	if m.page == layoutPage {
		for i, choice := range m.choices {

			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
	}

	if m.page == authPage {
		if m.choices[m.cursor] == "register" {
			s += "Registration:\n"
		}

		if m.choices[m.cursor] == "login" {
			s += "Login:\n"
		}
		s += "  " + m.form.email.View() + "\n"
		s += "  " + m.form.pass.View() + "\n"
	}

	s += faintStyle.Render("\n\nq - quit\nesc - discard\ntab - change input focus\n")
	return tea.NewView(s)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
