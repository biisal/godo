package styles

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

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
│ send /clear → clear chats        │
└──────────────────────────────────┘
`

var (
	ChatInputStyle = lipgloss.NewStyle().
			AlignHorizontal(lipgloss.Left).
			Foreground(lipgloss.Color("#ffffff")).
			BorderForeground(lipgloss.Color("#666666")).
			Border(lipgloss.NormalBorder()).
			Padding(0, 1, 2)

	PurpleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f8baff"))

	AgentContentStyle = PurpleStyle
	UserContentStyle  = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#99faff"))

	DescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#89b4fa")).
			Padding(0, 1)
)

type ListTheme struct {
}
type Theme struct {
	ListTheme ListTheme
}

func (t Theme) GetInstructionStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#29273b")).Foreground(lipgloss.Color("#bebcc4"))
}

func (t Theme) GetBackground() lipgloss.Color {
	return lipgloss.Color("#15141a")
}

func (t Theme) GetErrorInChatStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF8699")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#960018")).
		BorderRight(false).
		BorderLeft(true).
		BorderBottom(false).
		BorderTop(false).
		Padding(1, 1, 1).
		MarginTop(1).
		MarginBottom(1).
		Background(lipgloss.Color("#1a0a0a")).
		BorderForeground(lipgloss.Color("#960018"))
}
func (t Theme) GetUserContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#99faff")).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#99faff")).
		BorderRight(false).
		BorderLeft(true).
		BorderBottom(false).
		BorderTop(false).
		Padding(1, 1, 1).
		MarginTop(1).
		MarginBottom(1).
		Background(lipgloss.Color("#0B0C0D")).
		BorderForeground(lipgloss.Color("#99faff"))
}
func (t Theme) GetAgentContentStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#f8baff")).
		BorderForeground(lipgloss.Color("#f8baff"))
}
func (t Theme) GetBorderColor() lipgloss.Color {
	return lipgloss.Color("#666666")
}

func (t Theme) GetTestRedColor() lipgloss.Color {
	return lipgloss.Color("#FF0000")
}
func (t Theme) GetPurpleColoe() lipgloss.Color {
	return lipgloss.Color("#f8baff")
}
func (t Theme) GetGrayColor() lipgloss.Color {
	return lipgloss.Color("#AEAEAE")
}
func (t *Theme) GetGreenColor() lipgloss.Color {
	return lipgloss.Color("#00FF70")
}
func (t *Theme) GetDarkGreenColor() lipgloss.Color {
	return lipgloss.Color("#5AC88A")
}
func (t Theme) ListCustomDelegate(width int) list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	bg := t.GetBackground()
	defaltStyle := lipgloss.NewStyle().Background(bg).Padding(0, 2).Width(width)
	// Override title and description
	d.Styles.NormalTitle = defaltStyle
	d.Styles.NormalDesc = defaltStyle
	d.Styles.SelectedTitle = defaltStyle.Bold(true).Foreground(lipgloss.Color("#ffffff"))
	d.Styles.SelectedDesc = defaltStyle

	return d
}

func (lt *ListTheme) RowStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AEAEAE")).
		Border(lipgloss.RoundedBorder()).
		BorderLeft(true).
		BorderRight(false).
		BorderTop(false).
		BorderBottom(false).
		Padding(0, 1).
		Margin(1, 1, 0).
		BorderForeground(lipgloss.Color("#f8baff"))
}

func (lt *ListTheme) SelectedColor() lipgloss.Color {
	return lipgloss.Color("#f8baff")
}

func (t *Theme) TitleBackround() lipgloss.Color {
	return lipgloss.Color("#f8baff")

}
func (t *Theme) ModeStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(t.GetGrayColor()).
		Foreground(lipgloss.Color("#000000")).
		Padding(0, 1)
}

func (t *Theme) GetTitleInputStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#29273b")).Foreground(lipgloss.Color("#bebcc4"))
}
func (t *Theme) GetDescriptionInputStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#29273b")).Foreground(lipgloss.Color("#bebcc4"))
}

func (t *Theme) GetEorrorStyle() lipgloss.Style {
	return lipgloss.NewStyle().Padding(0, 1).
		Foreground(lipgloss.Color("#FF0000")).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF0000"))
}
