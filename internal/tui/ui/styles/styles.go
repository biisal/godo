// Package styles contains all the styles for the UI
package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type ThemeColors struct {
	Background          lipgloss.TerminalColor
	Foreground          lipgloss.TerminalColor
	Muted               lipgloss.TerminalColor
	MutedForeground     lipgloss.TerminalColor
	Border              lipgloss.TerminalColor
	InputBorder         lipgloss.TerminalColor
	Primary             lipgloss.TerminalColor
	Secondary           lipgloss.TerminalColor
	Accent              lipgloss.TerminalColor
	Success             lipgloss.TerminalColor
	SuccessForeground   lipgloss.TerminalColor
	Destructive         lipgloss.TerminalColor
	DestructiveBg       lipgloss.TerminalColor
	DestructiveBorder   lipgloss.TerminalColor
	PrimaryForeground   lipgloss.TerminalColor
	SecondaryForeground lipgloss.TerminalColor
	AccentForeground    lipgloss.TerminalColor
}

var (
	lightTheme = ThemeColors{
		Background:          lipgloss.NoColor{},
		Foreground:          lipgloss.NoColor{},
		Muted:               lipgloss.Color("7"),
		MutedForeground:     lipgloss.Color("8"),
		Border:              lipgloss.Color("8"),
		InputBorder:         lipgloss.Color("8"),
		Primary:             lipgloss.Color("4"),
		Secondary:           lipgloss.Color("255"),
		Accent:              lipgloss.Color("5"),
		Success:             lipgloss.Color("2"),
		SuccessForeground:   lipgloss.NoColor{},
		Destructive:         lipgloss.Color("1"),
		DestructiveBg:       lipgloss.NoColor{},
		DestructiveBorder:   lipgloss.Color("1"),
		PrimaryForeground:   lipgloss.NoColor{},
		SecondaryForeground: lipgloss.NoColor{},
		AccentForeground:    lipgloss.NoColor{},
	}
	darkTheme = ThemeColors{
		Background:          lipgloss.NoColor{},
		Foreground:          lipgloss.NoColor{},
		Muted:               lipgloss.Color("8"),
		MutedForeground:     lipgloss.Color("7"),
		Border:              lipgloss.Color("8"),
		InputBorder:         lipgloss.Color("8"),
		Primary:             lipgloss.Color("13"),
		Secondary:           lipgloss.Color("234"),
		Accent:              lipgloss.Color("5"),
		Success:             lipgloss.Color("10"),
		SuccessForeground:   lipgloss.NoColor{},
		Destructive:         lipgloss.Color("9"),
		DestructiveBg:       lipgloss.NoColor{},
		DestructiveBorder:   lipgloss.Color("1"),
		PrimaryForeground:   lipgloss.NoColor{},
		SecondaryForeground: lipgloss.NoColor{},
		AccentForeground:    lipgloss.NoColor{},
	}
	colors = ThemeBuilder()
)

func ThemeBuilder() ThemeColors {
	if termenv.HasDarkBackground() {
		return darkTheme
	}
	return lightTheme
}

// Colors returns the active semantic theme colors.
func Colors() ThemeColors { return colors }

var (
	ChatInputStyle = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Left).
			Foreground(colors.Foreground).
			BorderForeground(colors.Accent).
			Background(colors.Secondary).
			Border(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(false).
			BorderTop(false)

	PrimaryStyle = lipgloss.NewStyle().
			Foreground(colors.Primary)

	AgentContentStyle = lipgloss.NewStyle().Foreground(colors.Accent)

	UserContentStyle = lipgloss.NewStyle().
				Foreground(colors.SecondaryForeground).Margin(1, 0).Padding(1).Background(colors.Secondary)

	AgentChatViewStyle = lipgloss.NewStyle().
				PaddingLeft(0)

	AgentPromptStyle = ChatInputStyle.Margin(1, 0)

	ThinkingTokenStyle = lipgloss.NewStyle().
				Foreground(colors.MutedForeground).
				Italic(true)

	HalfWidthLeftStyle = lipgloss.NewStyle().
				AlignHorizontal(lipgloss.Left)

	HalfWidthRightStyle = lipgloss.NewStyle().
				AlignHorizontal(lipgloss.Right)

	CenteredTitleStyle = lipgloss.NewStyle().
				Align(lipgloss.Center)

	TodoInputStyle = ChatInputStyle.
			Height(5).
			BorderTop(true).
			Border(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderRight(false).
			BorderLeft(false)

	TodoIDInputStyle = ChatInputStyle.
				Height(1).
				BorderTop(true).
				BorderBottom(false).
				BorderRight(false).
				BorderLeft(false)

	TodoDescStyle = lipgloss.NewStyle().
			PaddingBottom(1)

	TodoListStyle = lipgloss.NewStyle().
			Padding(1)

	TodoDescViewportStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderLeft(true).
				BorderBottom(false).
				Padding(1, 1, 0, 1).
				BorderRight(false).
				BorderTop(false)

	InstructionStyle = lipgloss.NewStyle().
				Padding(0, 1).
				Foreground(colors.MutedForeground)

	ErrorInChatStyle = lipgloss.NewStyle().
				Padding(1, 1, 1).
				MarginTop(1).
				MarginBottom(1).
				Foreground(colors.Destructive).
				Background(colors.DestructiveBg)

	ListRowStyle = lipgloss.NewStyle().
			Foreground(colors.Foreground).
			Border(lipgloss.RoundedBorder()).
			BorderLeft(true).
			BorderRight(false).
			BorderTop(false).
			BorderBottom(false).
			Padding(0, 1).
			Margin(1, 1, 0).
			BorderForeground(colors.Primary)

	ModeStyle = lipgloss.NewStyle().
			Background(colors.Secondary).
			Foreground(colors.SecondaryForeground).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(colors.Destructive).
			Background(colors.DestructiveBg)

	ShellSidePanelStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colors.Border).
				Padding(1).
				MarginLeft(1)

	ShellOutputStyle = lipgloss.NewStyle().
				Foreground(colors.Success)
)

type (
	ListTheme struct{}
	Theme     struct {
		ListTheme ListTheme
	}
)
