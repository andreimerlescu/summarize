package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/teilomillet/gollm"
)

// --- STYLING ---
var (
	// Styles for chat messages
	senderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("5"))            // User (Purple)
	botStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))            // AI (Cyan)
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true) // Error messages

	// A slight border for the chat viewport
	viewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")). // Gray
			Padding(1)
)

func StartChat(buf *bytes.Buffer) {
	// Create and run the Bubble Tea program.
	// tea.WithAltScreen() provides a full-window TUI experience.
	// CORRECTED: Pass aiPtr.llm directly, not its address.
	p := tea.NewProgram(initialModel(NewAI(), buf.String()), tea.WithAltScreen(), tea.WithMouseCellMotion())

	finalModel, err := p.Run()
	if err != nil {
		log.Fatalf("❌ Oh no, there's been an error: %v", err)
	}

	if m, ok := finalModel.(model); ok && len(m.messages) > 1 {
		// More than 1 message means there was a conversation (initial message + at least one more).

		// Create a timestamped filename.
		timestamp := time.Now().Format("2006-01-02_15-04-05")
		filename := fmt.Sprintf("chatlog_%s.md", timestamp)

		var output bytes.Buffer
		output.WriteString("# Summarize Chat Log " + timestamp + "\n\n")
		for i := 0; i < len(m.messages); i++ {
			message := m.messages[i]
			output.WriteString(message)
			output.WriteString("\n")
		}

		// Write the chat history to the file.
		if writeErr := os.WriteFile(filepath.Join(*figs.String(kOutputDir), filename), output.Bytes(), 0644); writeErr != nil {
			fmt.Printf("\n❌ Could not save chat log: %v\n", writeErr)
		} else {
			fmt.Printf("\n📝 Chat log saved to %s\n", filename)
		}
	}
}

// --- BUBBLETEA MESSAGES ---
// We use custom messages to communicate between our async LLM calls and the UI.

// aiResponseMsg is sent when the AI has successfully generated a response.
type aiResponseMsg string

// errorMsg is sent when an error occurs during the AI call.
type errorMsg struct{ err error }

// --- BUBBLETEA MODEL ---
// The model is the single source of truth for the state of your application.
type model struct {
	// CORRECTED: The llm field is now the interface type, not a pointer to it.
	llm          gollm.LLM
	viewport     viewport.Model
	textarea     textarea.Model
	messages     []string
	summary      string
	isGenerating bool
	err          error
	ctx          context.Context
	chatHistory  []string
}

// initialModel creates the starting state of our application.
// CORRECTED: The llm parameter is now the interface type.
func initialModel(llm gollm.LLM, summary string) model {
	// Configure the text area for user input.
	ta := textarea.New()
	ta.Placeholder = "Send a message... (press Enter to send, Esc to quit)"
	ta.Focus()
	ta.Prompt = "┃ "
	ta.SetHeight(1)
	// Remove the default behavior of Enter creating a new line.
	ta.KeyMap.InsertNewline.SetEnabled(false)

	// The viewport is the scrolling area for the chat history.
	vp := viewport.New(0, 0) // Width and height are set dynamically

	if len(summary) == 0 {
		panic("no summary")
	}

	msg := fmt.Sprintf("%s %d bytes!", "Welcome to Summarize AI Chat! We've analyzed your project workspace and are ready to chat with you about ", len(summary))

	return model{
		llm:          llm,
		textarea:     ta,
		viewport:     vp,
		summary:      summary,
		messages:     []string{msg},
		chatHistory:  []string{},
		isGenerating: false,
		err:          nil,
		ctx:          context.Background(),
	}
}

// generateResponseCmd is a Bubble Tea command that calls the LLM in a goroutine.
// This prevents the UI from blocking while waiting for the AI.
func (m model) generateResponseCmd() tea.Cmd {
	return func() tea.Msg {
		userInput := m.textarea.Value()
		m.chatHistory = append(m.chatHistory, userInput)

		var wc strings.Builder
		breaker := "---ARM-GO-SUMMARIZE-BREAK-POINT---"
		if len(m.messages) > 0 {
			wc.WriteString("You are now continuing this conversation. This is the chat log: ")
			for i := 0; i < len(m.messages); i++ {
				v := m.messages[i]
				x := fmt.Sprintf("line %d: %s\n", i+1, v)
				wc.WriteString(x)
			}
			wc.WriteString("\n")
			wc.WriteString("The summarized project is:\n")
			parts := strings.Split(m.summary, breaker)
			if len(parts) == 2 {
				oldPrefix := strings.Clone(parts[0])
				oldSummary := strings.Clone(parts[1])
				newSummary := oldPrefix + wc.String() + oldSummary
				m.summary = newSummary
				wc.Reset()
			}
			wc.WriteString(m.summary)
			wc.WriteString("\n")
		} else {
			wc.WriteString("Your name is Summarize in this engagement. This is a comprehensive one page contents of " +
				"entire directory (recursively) of a specific subset of files by extension choice and a strings.Contains() avoid list" +
				"that is used to generate the following summary.\n\n" +
				"You are communicating with the user and shall refer to them as Commander. You are speaking to them in a " +
				"golang bubbletea TUI chat terminal that is ")
			wc.WriteString(strconv.Itoa(m.viewport.Width))
			wc.WriteString(" (int) width and ")
			wc.WriteString(strconv.Itoa(m.viewport.Height))
			wc.WriteString(" (int) height with ")
			wc.WriteString(strconv.Itoa(m.viewport.VisibleLineCount()))
			wc.WriteString(" (int) visible lines in the viewport. Your responses should singularly fit in the terminal" +
				"window. Be aware that your response will be formatted using wordwrap.String(<message>, m.viewport.Width) in the Bubbletea TUI, so ensure your message fits within the viewport width. " +
				"\n\n" +
				"The user expects you to be professional and keep focused on the language that you detect from the summary.\n" +
				"Be prepared to answer questions about security, best practices, and security concerns that you have over" +
				"the code. However, do not get distracted. Always follow the lead of the DevOps engineer. Do not be afraid to" +
				"offend. Your brutal honesty is welcome here and iron sharpens iron. Here is the summary now:\n" +
				breaker + "\n")
			oldSummary := strings.Clone(m.summary)
			newSummary := oldSummary + wc.String()
			m.summary = newSummary
			wc.Reset()
			wc.WriteString(m.summary)
			wc.WriteString("\n")
		}

		var systemPrompt strings.Builder
		systemPrompt.WriteString("This is a summary of a project that you are to respond to user prompts with the contents of this project.\n\n")
		systemPrompt.WriteString(m.summary)
		systemPrompt.WriteString("\n")

		var (
			response = ""
			err      = errors.New("failed to ask agent for a response")
			prompt   = gollm.NewPrompt(userInput,
				gollm.WithContext(strings.Join(m.chatHistory, "\n")),
				gollm.WithSystemPrompt(systemPrompt.String(), gollm.CacheTypeEphemeral),
				gollm.WithMaxLength(7777),
				gollm.WithDirectives("Be concise and offer complete solutions",
					"Act as Commander Data from the USS Starship Enterprise acting as an AI Agent assisting the user",
					"Refer to the user as Commander",
					"Speak as if you were on a Military Base as a member of the USS Starship Enterprise",
					"Speak as if you are on duty with fellow crew mates",
					"When replying to followup requests, build on your previous answer",
					"When a mistake is identified by the user, use the full previous response to modify and return",
					"Do not be afraid to offend and always give an honest answer in as few words as possible",
				),
				gollm.WithOutput(
					fmt.Sprintf("%s %d wide %d tall.", "Do not apply any formatting to the output"+
						" text except for line breaks and spaces. Commands and codes should be indented by 4 spaces "+
						"on the left and right side of the line and the text will render inside of a Golang BubbleTea"+
						"TUI window that is ", m.viewport.Width-5, m.viewport.Height-5),
				),
			)
		)
		response, err = m.llm.Generate(m.ctx, prompt)
		if err != nil {
			return errorMsg{err} // On error, return an error message.
		}
		response = response + "\n\n"

		return aiResponseMsg(response) // On success, return the AI's response.
	}
}

// --- BUBBLETEA LIFECYCLE ---

// Init is called once when the program starts. It can return an initial command.
func (m model) Init() tea.Cmd {
	return textarea.Blink // Start with a blinking cursor in the textarea.
}

// Update is the core of the application. It's called whenever a message (event) occurs.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		vpCmd tea.Cmd
	)

	// Handle updates for the textarea and viewport components.
	m.textarea, taCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	// Handle key presses
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			// Don't send if the AI is already working or input is empty.
			if m.isGenerating || m.textarea.Value() == "" {
				return m, nil
			}

			// Add the user's message to the history and set the generating flag.
			m.messages = append(m.messages, senderStyle.Render("You: ")+m.textarea.Value())
			m.isGenerating = true
			m.err = nil // Clear any previous error.

			// Create the command to call the LLM and reset the input.
			cmd := m.generateResponseCmd()
			m.textarea.Reset()
			m.viewport.SetContent(wordwrap.String(strings.Join(m.messages, "\n"), m.viewport.Width))
			m.viewport.GotoBottom() // Scroll to the latest message.

			return m, cmd
		}

	// Handle window resizing
	case tea.WindowSizeMsg:
		// Adjust the layout to the new window size.
		viewportStyle.Width(msg.Width - 2)   // Subtract border width
		viewportStyle.Height(msg.Height - 4) // Subtract textarea, help text, and border
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height - 4
		m.textarea.SetWidth(msg.Width)
		m.viewport.SetContent(wordwrap.String(strings.Join(m.messages, "\n"), m.viewport.Width)) // Re-render content

	// Handle the AI's response
	case aiResponseMsg:
		m.isGenerating = false
		m.messages = append(m.messages, botStyle.Render("Summarize AI: ")+string(msg))
		m.viewport.SetContent(wordwrap.String(strings.Join(m.messages, "\n"), m.viewport.Width))
		m.viewport.GotoBottom()

	// Handle any errors from the AI call
	case errorMsg:
		m.isGenerating = false
		m.err = msg.err
	}

	return m, tea.Batch(taCmd, vpCmd) // Return any commands from the components.
}

// View renders the UI. It's called after every Update.
func (m model) View() string {
	var bottomLine string
	if m.isGenerating {
		bottomLine = "🤔 Thinking..."
	} else if m.err != nil {
		bottomLine = errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	} else {
		bottomLine = m.textarea.View()
	}

	// Join the viewport and the bottom line (textarea or status) vertically.
	return lipgloss.JoinVertical(
		lipgloss.Left,
		viewportStyle.Render(m.viewport.View()),
		bottomLine,
	)
}
