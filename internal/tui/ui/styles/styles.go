package styles

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	ColorPurple      = lipgloss.Color("#f8baff")
	ColorCyan        = lipgloss.Color("#99faff")
	ColorGray        = lipgloss.Color("#AEAEAE")
	ColorDarkGray    = lipgloss.Color("#222222ff")
	ColorBg          = lipgloss.Color("#15141a")
	ColorError       = lipgloss.Color("#FF0000")
	ColorErrorBg     = lipgloss.Color("#1a0a0a")
	ColorErrorBorder = lipgloss.Color("#960018")
	ColorGreen       = lipgloss.Color("#00FF70")
	ColorDarkGreen   = lipgloss.Color("#5AC88A")
)

var (
	ChatInputStyle = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Left).
			Foreground(lipgloss.Color("#ffffff")).
			BorderForeground(ColorDarkGray).
			Border(lipgloss.NormalBorder()).
			Padding(0, 1, 2)

	PurpleStyle = lipgloss.NewStyle().
			Foreground(ColorPurple)

	AgentContentStyle = PurpleStyle

	UserContentStyle = lipgloss.NewStyle().
				Foreground(ColorCyan).Margin(1, 0).Padding(1).Background(ColorDarkGray)

	AgentChatViewStyle = lipgloss.NewStyle().
				PaddingLeft(3)

	AgentPromptStyle = ChatInputStyle.
				BorderBottom(false).
				BorderRight(false).
				BorderLeft(false)

	ThinkingTokenStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#555555")).
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

	TodoIdInputStyle = ChatInputStyle.
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
				Background(lipgloss.Color("#29273b")).
				Foreground(lipgloss.Color("#bebcc4"))

	ErrorInChatStyle = lipgloss.NewStyle().
				Padding(1, 1, 1).
				MarginTop(1).
				MarginBottom(1).
				Foreground(ColorError).
				Background(ColorErrorBg)

	ListRowStyle = lipgloss.NewStyle().
			Foreground(ColorGray).
			Border(lipgloss.RoundedBorder()).
			BorderLeft(true).
			BorderRight(false).
			BorderTop(false).
			BorderBottom(false).
			Padding(0, 1).
			Margin(1, 1, 0).
			BorderForeground(ColorPurple)

	ModeStyle = lipgloss.NewStyle().
			Background(ColorGray).
			Foreground(lipgloss.Color("#000000")).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(ColorError).
			Background(ColorErrorBg)
)

type ListTheme struct{}
type Theme struct {
	ListTheme ListTheme
}
