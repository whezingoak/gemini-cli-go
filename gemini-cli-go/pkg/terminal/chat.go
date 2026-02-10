package terminal

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/whezingoak/gemini-cli-go/pkg/api"
)

type ChatModel struct {
	viewport    viewport.Model
	textarea    textarea.Model
	session     *api.ChatSession
	messages    []string
	err         error
	width, height int
}

type errMsg error
type responseMsg string

func StartChat(client *api.Client) error {
	p := tea.NewProgram(initialModel(client))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func initialModel(client *api.Client) ChatModel {
	session := client.StartChat()
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to Gemini CLI! Type a message and press Enter.`)

	return ChatModel{
		textarea: ta,
		viewport: vp,
		session:  session,
		messages: []string{},
		err:      nil,
	}
}

func (m ChatModel) Init() tea.Cmd {
	return textarea.Blink
}

func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			v := m.textarea.Value()
			if v == "" {
				return m, nil
			}

			userMsg := fmt.Sprintf("**You:** %s", v)
			renderedUserMsg, _ := renderMarkdown(userMsg, m.viewport.Width)
			m.messages = append(m.messages, renderedUserMsg)

			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()

			// Send to API
			return m, func() tea.Msg {
				resp, err := m.session.SendMessage(context.Background(), v)
				if err != nil {
					return errMsg(err)
				}
				if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
					return responseMsg(resp.Candidates[0].Content.Parts[0].Text)
				}
				return errMsg(fmt.Errorf("no response content"))
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		errorMsg := fmt.Sprintf("**Error:** %s", msg.Error())
		renderedError, _ := renderMarkdown(errorMsg, m.viewport.Width)
		m.messages = append(m.messages, renderedError)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		return m, nil

	case responseMsg:
		geminiMsg := fmt.Sprintf("**Gemini:**\n%s", string(msg))
		renderedGemini, _ := renderMarkdown(geminiMsg, m.viewport.Width)
		m.messages = append(m.messages, renderedGemini)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		return m, nil

	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if m.height > verticalMarginHeight {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
			m.textarea.SetWidth(msg.Width)
		}
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m ChatModel) headerView() string {
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFDF5")).
		Background(lipgloss.Color("#25A065")).
		Padding(0, 1).
		Render("Gemini CLI")
	line := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title))))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m ChatModel) footerView() string {
	return m.textarea.View()
}

func (m ChatModel) View() string {
	return fmt.Sprintf(
		"%s\n%s\n%s",
		m.headerView(),
		m.viewport.View(),
		m.footerView(),
	)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func renderMarkdown(text string, width int) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return text, err
	}
	return renderer.Render(text)
}
