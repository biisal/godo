package styles

import "github.com/charmbracelet/lipgloss"

const Logo = `
   godo     ..----.._    _
            .' .--.    "-.(O)_
'-.__.-'"'=:|  ,  _)_ \__ . c\'-..
             ''------'---''---'-"`

const ChatInstructions = `
┌──────────── Controls ────────────┐
│ Ctrl + A     → Back to Todo Mode │
│ Page Up      → Scroll Up         │
│ Page Down    → Scroll Down       │
│ Esc / Ctrl+C → Quit              │
└──────────────────────────────────┘
`

var (
	ChatInputStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#130F1A")). // Gray background
			Foreground(lipgloss.Color("#ffffff")). // White text
			Padding(0, 1, 2)
	PurpleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f8baff")).Background(lipgloss.Color("#130F1A"))

	AgentContentStyle = PurpleStyle.Background(lipgloss.Color("#130F1A"))
	UserContentStyle  = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#99faff")).Background(lipgloss.Color("#130F1A"))

	DescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89b4fa")).
			Padding(0, 1)
	TitleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	GreenStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00"))
	HelpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#A593E0")).Underline(true).Align(lipgloss.Center)
	RedStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
	SelectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	ItemStyle         = lipgloss.NewStyle()
	DocStyle          = lipgloss.NewStyle().Margin(1, 2)
	FadeStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	FocusedStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFFF")).Bold(true)
	BlurredStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#888")).Italic(true).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4F6892"))
	BoxStyle          = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#89b4fa")).
				Margin(0, 0)
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF0000")).Padding(0, 1)
	EventStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF9E")).
			Bold(true).Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#00FF9E")).Padding(0, 1)
)

type Theme struct {
}

func (t Theme) GetInstructionStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#29273b")).Foreground(lipgloss.Color("#bebcc4"))
}

func (t Theme) GetBackground() lipgloss.Color {
	return lipgloss.Color("#15141a")
}

func (t Theme) GetUserContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#99faff")).
		Background(lipgloss.Color("#130F1A")).
		Border(lipgloss.RoundedBorder()).
		BorderLeft(true).
		BorderRight(false).
		BorderTop(false).
		BorderBottom(false).
		Padding(0, 1).
		Margin(1, 1, 0).
		BorderForeground(lipgloss.Color("#99faff"))
}
func (t Theme) GetAgentContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f8baff")).
		Background(lipgloss.Color("#130F1A")).
		Border(lipgloss.RoundedBorder()).
		BorderLeft(true).
		BorderRight(false).
		BorderTop(false).
		Padding(0, 1).
		Margin(1, 1, 2).
		BorderBottom(false).
		BorderForeground(lipgloss.Color("#f8baff"))
}
func (t Theme) GetBorderColor() lipgloss.Color {
	return lipgloss.Color("#89b4fa")
}

func (t Theme) GetTestRedColor() lipgloss.Color {
	return lipgloss.Color("#FF0000")
}
