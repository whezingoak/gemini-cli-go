package terminal

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/whezingoak/gemini-cli-go/pkg/api"
)

type ChatModel struct {
	viewport        viewport.Model
	textarea        textarea.Model
	session         *api.ChatSession
	messages        []string // Rendered HTML/ANSI messages
	err             error
	width, height   int

	// Streaming state
	streaming       bool
	currentResponse string
	sub             chan string // Channel for receiving chunks
}

type errMsg error
type chunkMsg string
type streamDoneMsg struct{}

func StartChat(client *api.Client) error {
	p := tea.NewProgram(initialModel(client))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}

func initialModel(client *api.Client) ChatModel {
	session := client.StartChat()

	// Load GEMINI.md if it exists
	if content, err := os.ReadFile("GEMINI.md"); err == nil {
		session.AddContext(fmt.Sprintf("Project Context (GEMINI.md):\n%s", string(content)))
	}

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

			// Handle commands
			if strings.HasPrefix(v, "/") {
				return m.handleCommand(v)
			}

			if m.streaming {
				return m, nil // Ignore input while streaming
			}

			userMsg := fmt.Sprintf("**You:** %s", v)
			renderedUserMsg, _ := renderMarkdown(userMsg, m.viewport.Width)
			m.messages = append(m.messages, renderedUserMsg)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))

			m.textarea.Reset()
			m.viewport.GotoBottom()

			m.streaming = true
			m.currentResponse = ""
			m.sub = make(chan string)

			return m, tea.Batch(
				startStreaming(m.session, v, m.sub),
				waitForChunk(m.sub),
			)
		}

	case errMsg:
		m.err = msg
		m.streaming = false
		errorMsg := fmt.Sprintf("**Error:** %s", msg.Error())
		renderedError, _ := renderMarkdown(errorMsg, m.viewport.Width)
		m.messages = append(m.messages, renderedError)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		return m, nil

	case chunkMsg:
		m.currentResponse += string(msg)

		geminiMsg := fmt.Sprintf("**Gemini:**\n%s", m.currentResponse)
		rendered, _ := renderMarkdown(geminiMsg, m.viewport.Width)

		fullContent := strings.Join(m.messages, "\n") + "\n" + rendered
		m.viewport.SetContent(fullContent)
		m.viewport.GotoBottom()

		return m, waitForChunk(m.sub)

	case streamDoneMsg:
		m.streaming = false
		// Finalize
		geminiMsg := fmt.Sprintf("**Gemini:**\n%s", m.currentResponse)
		rendered, _ := renderMarkdown(geminiMsg, m.viewport.Width)
		m.messages = append(m.messages, rendered)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
		m.currentResponse = "" // Clear current response
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

func startStreaming(session *api.ChatSession, prompt string, sub chan string) tea.Cmd {
	return func() tea.Msg {
		_, err := session.SendMessageStream(context.Background(), prompt, func(chunk string) {
			sub <- chunk
		})
		if err != nil {
			return errMsg(err)
		}
		close(sub)
		return nil
	}
}

func waitForChunk(sub chan string) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-sub
		if !ok {
			return streamDoneMsg{}
		}
		return chunkMsg(chunk)
	}
}

func (m ChatModel) handleCommand(cmd string) (ChatModel, tea.Cmd) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return m, nil
	}

	switch parts[0] {
	case "/clear":
		m.messages = []string{}
		m.viewport.SetContent("Chat cleared.")
		m.textarea.Reset()
		return m, nil
	case "/exit", "/quit":
		return m, tea.Quit
	case "/help":
		helpMsg, _ := renderMarkdown(`
**Available Commands:**
- **/add <file>**: Add file content to chat context
- **/clear**: Clear chat history
- **/exit**: Exit chat
- **/help**: Show this help message
`, m.viewport.Width)
		m.messages = append(m.messages, helpMsg)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.textarea.Reset()
		return m, nil
	case "/add":
		if len(parts) < 2 {
			errMsg := "**Error:** Usage: `/add <file>`"
			rendered, _ := renderMarkdown(errMsg, m.viewport.Width)
			m.messages = append(m.messages, rendered)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			return m, nil
		}

		filepath := parts[1]
		content, err := os.ReadFile(filepath)
		if err != nil {
			errMsg := fmt.Sprintf("**Error:** Failed to read file `%s`: %s", filepath, err)
			rendered, _ := renderMarkdown(errMsg, m.viewport.Width)
			m.messages = append(m.messages, rendered)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			return m, nil
		}

		// Add to context
		m.session.AddContext(fmt.Sprintf("File: %s\nContent:\n```\n%s\n```", filepath, string(content)))

		successMsg := fmt.Sprintf("**System:** Added `%s` to context.", filepath)
		rendered, _ := renderMarkdown(successMsg, m.viewport.Width)
		m.messages = append(m.messages, rendered)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.textarea.Reset()
		return m, nil

	default:
		errMsg := fmt.Sprintf("**Error:** Unknown command `%s`", parts[0])
		rendered, _ := renderMarkdown(errMsg, m.viewport.Width)
		m.messages = append(m.messages, rendered)
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.textarea.Reset()
		return m, nil
	}
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
