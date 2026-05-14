package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"github.com/b2jant/bricktap/internal/adapters"
	"github.com/b2jant/bricktap/internal/core"
	"github.com/b2jant/bricktap/internal/dialects"
	"github.com/b2jant/bricktap/internal/parser"
	"github.com/b2jant/bricktap/internal/scanner"
)

// State enumeration for our Bubble Tea application
type state int

const (
	stateForm state = iota
	stateGenerating
	stateDone
)

type model struct {
	state        state
	form         *huh.Form
	spinner      spinner.Model
	framework    string
	dialect      string
	filesCreated int
	err          error
}

func initialModel() model {
	// Define the interactive form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose Target Framework").
				Options(
					huh.NewOption("dbt", "dbt"),
				).
				Value(&framework),
			huh.NewSelect[string]().
				Title("Target Data Warehouse?").
				Options(
					huh.NewOption("Snowflake", "snowflake"),
					huh.NewOption("BigQuery", "bigquery"),
					huh.NewOption("Postgres", "postgres"),
				).
				Value(&dialect),
		),
	)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		state:   stateForm,
		form:    form,
		spinner: s,
	}
}

// Global variables for form state bindings
var framework string
var dialect string

func (m model) Init() tea.Cmd {
	return m.form.Init()
}

// Define a message for when generation is complete
type generationCompleteMsg struct {
	filesGenerated int
	err            error
}

func startGeneration(selectedFramework string, selectedDialect string) tea.Cmd {
	return func() tea.Msg {
		inputDir := "./semantic_models"
		outputDir := "./models"

		// 1. Scan for all YAML models
		files, err := scanner.Scan(inputDir, ".yaml")
		if err != nil {
			return generationCompleteMsg{err: fmt.Errorf("failed to scan directory: %w", err)}
		}

		// Configure the SQL Dialect
		var d dialects.Dialect
		if selectedDialect == "snowflake" {
			d = dialects.NewSnowflakeDialect()
		} else {
			d = &dialects.BaseDialect{} // Default ANSI
		}

		// Configure the Generator Framework
		var gen adapters.Generator
		if selectedFramework == "dbt" {
			gen = adapters.NewDbtAdapter()
		} else {
			return generationCompleteMsg{err: fmt.Errorf("framework %s not yet implemented", selectedFramework)}
		}

		// Define some dummy global rules for testing the MVP
		rules := core.GlobalRules{
			TypeCasting: map[string]string{
				"string": "NULLIF(TRIM({column}), '')",
			},
		}

		// 2. Parse and Generate each file
		successCount := 0
		for _, fileInfo := range files {
			parsedModel, err := parser.ParseModel(fileInfo.AbsolutePath, rules)
			if err != nil {
				// We skip failing models for now to continue processing
				continue
			}

			if err := gen.Generate(*parsedModel, fileInfo, d, outputDir); err == nil {
				successCount++
			}
		}

		return generationCompleteMsg{filesGenerated: successCount, err: nil}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

	case generationCompleteMsg:
		m.state = stateDone
		m.err = msg.err
		m.filesCreated = msg.filesGenerated
		return m, tea.Quit
	}

	switch m.state {
	case stateForm:
		form, cmd := m.form.Update(msg)
		if f, ok := form.(*huh.Form); ok {
			m.form = f
		}
		if m.form.State == huh.StateCompleted {
			m.state = stateGenerating
			m.framework = framework
			m.dialect = dialect
			return m, tea.Batch(m.spinner.Tick, startGeneration(m.framework, m.dialect))
		}
		return m, cmd

	case stateGenerating:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateForm:
		return m.form.View()
	case stateGenerating:
		return fmt.Sprintf("\n %s Generating %s models for %s...\n", m.spinner.View(), m.framework, m.dialect)
	case stateDone:
		if m.err != nil {
			return fmt.Sprintf("\n❌ Error generating models: %v\n", m.err)
		}

		// Render the final Lipgloss summary box
		style := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2).
			Foreground(lipgloss.Color("228"))

		content := fmt.Sprintf("✨ Successfully generated %d models!\n\nFramework: %s\nDialect: %s\nInput: ./semantic_models\nOutput: ./models", m.filesCreated, m.framework, m.dialect)
		return "\n" + style.Render(content) + "\n"
	}
	return ""
}

// Start initializes the Bubble Tea application
func Start() error {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	return err
}
