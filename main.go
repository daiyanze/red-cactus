package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

var (
	color             = termenv.ColorProfile().Color
	selectedItemStyle = termenv.Style{}.Background(color("237")).Styled
	dirItemStyle      = termenv.Style{}.Foreground(color("33")).Styled
)

type item struct {
	name  string
	isDir bool
    prefix string ""
}

type model struct {
	items          []item
	cursor         int
	curDir         string
	curCommand     string
	commandModeOn  bool
	ready          bool
	viewport       viewport.Model
}

const footerHeight int = 2

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.Model{Height: msg.Height - footerHeight, Width: msg.Width}
			m.ready = true
		} else {
			m.viewport.Height = msg.Height - footerHeight
			m.viewport.Width = msg.Width
		}
	case tea.KeyMsg:
		keyStr := msg.String()
		if keyStr == ":" && !m.commandModeOn {
			m.commandModeOn = true
		}

		if m.commandModeOn {
			m.curCommand += keyStr

			if keyStr == "ctrl+c" || keyStr == "esc" {
				m.curCommand = ""
				m.commandModeOn = false
			}
			break
		}

		switch keyStr {
		case "ctrl+c", "q":
			return m, tea.Quit
        case " ":
            return m, tea.Quit
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
                if (m.cursor <= m.viewport.YOffset + 3) {
                    m.viewport.LineUp(1)
                }
			} else if m.cursor == 0 {
				m.cursor = len(m.items) - 1
				m.viewport.GotoBottom()
			}
		case "j", "down":
			if m.cursor < len(m.items)-1 {
				m.cursor++
                if (m.cursor >= m.viewport.YOffset + m.viewport.Height - 3) {
                    m.viewport.LineDown(1)
                }
			} else if m.cursor == len(m.items)-1 {
				m.cursor = 0
				m.viewport.GotoTop()
			}
		case "enter":
			if !m.items[m.cursor].isDir {
				break
			}

			// If the selected is the current directory, change curDir to it's parent dir
			path := filepath.Join(m.curDir, m.items[m.cursor].name)
			if m.cursor == 0 {
				m.curDir = filepath.Join(m.curDir, "../")
				path = m.curDir
			} else {
				m.curDir = path
			}

			files := fetchDir(path)
			m.items = []item{
				{
					name:  m.curDir,
					isDir: true,
				},
			}
			for _, file := range files {
				m.items = append(m.items,
					item{
						name:  file.Name(),
						isDir: file.IsDir(),
					},
				)
			}

			m.cursor = 0
		}
	}

	s := ""
	for i, item := range m.items {
		space := " "
		if i != 0 {
			space = "   "
		}
		str := fmt.Sprintf("%s%s%s\n", space, item.name, strings.Repeat(" ", m.viewport.Width))

		if item.isDir {
			str = dirItemStyle(str)
		}

		if m.cursor == i {
			s += selectedItemStyle(str)
		} else {
			s += str
		}
	}

	m.viewport.SetContent(s)

    return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("%s\n%s", m.contentView(), m.bottomView())
}

func (m model) contentView() string {
    return m.viewport.View()
}

func (m model) bottomView() string {
	return termenv.String("command"+ m.curCommand).Foreground(color("241")).String()
}

func fetchDir(path string) []fs.FileInfo {
	files, err := ioutil.ReadDir(path)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return files
}

func main() {
	curDir, errGetwd := os.Getwd()

	if errGetwd != nil {
		log.Fatal(errGetwd)
	}

	files := fetchDir(curDir)

	initModel := model{
		items: []item{
			{
				name:  curDir,
				isDir: true,
			},
		},
		cursor: 0,
		curDir: curDir,
	}

	for _, file := range files {
		initModel.items = append(initModel.items,
			item{
				name:  file.Name(),
				isDir: file.IsDir(),
			},
		)
	}

	p := tea.NewProgram(initModel)
	p.EnterAltScreen()
	defer p.ExitAltScreen()

	if errProgram := p.Start(); errProgram != nil {
		log.Fatal(errProgram)
		os.Exit(1)
	}
}
