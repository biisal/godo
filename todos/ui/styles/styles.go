package styles

import "github.com/charmbracelet/lipgloss"

const Logo = `
  ______          __           ___                    __ 
 /_  __/___  ____/ /___       /   | ____ ____  ____  / /_
  / / / __ \/ __  / __ \     / /| |/ __ \/ _ \/ __ \/ __/
 / / / /_/ / /_/ / /_/ /    / ___ / /_/ /  __/ / / / /_  
/_/  \____/\__,_/\____/    /_/  |_\__, /\___/_/ /_/\__/  
                                 /____/                                                                                                                                                   
`

var (
	AgentContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#f8baff")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#f8baff")).Bold(true)
	UserContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#99faff")).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#99faff")).Bold(true)
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
