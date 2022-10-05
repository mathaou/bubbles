package table

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

type tableModel struct {
	XOffset int
	YOffset int
	Width   int
	Height  int
}

// Model defines a state for the table widget.
type Model struct {
	KeyMap KeyMap
	view   tableModel
	cols   []Column
	// and array of arrays
	rows       []Row
	row        int
	col        int
	focus      bool
	selectCell bool
	styles     Styles
}

// Row represents one line in the table. Each index is a cell.
type Row []string

// Column defines the table structure.
type Column struct {
	Title string
	Width int
}

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the menu menu.
type KeyMap struct {
	LineUp           key.Binding
	LineDown         key.Binding
	LineRight        key.Binding
	LineLeft         key.Binding
	PageUp           key.Binding
	PageDown         key.Binding
	HalfPageUp       key.Binding
	HalfPageDown     key.Binding
	GotoTop          key.Binding
	GotoBottom       key.Binding
	ToggleCellSelect key.Binding
}

// DefaultKeyMap returns a default set of keybindings.
func DefaultKeyMap() KeyMap {
	const spacebar = " "
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		LineLeft: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("left/h", "move cells or col left")),
		LineRight: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("right/l", "move cell or cols right")),
		PageUp: key.NewBinding(
			key.WithKeys("b", "pgup"),
			key.WithHelp("b/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("f", "pgdown", spacebar),
			key.WithHelp("f/pgdn", "page down"),
		),
		HalfPageUp: key.NewBinding(
			key.WithKeys("u", "ctrl+u"),
			key.WithHelp("u", "½ page up"),
		),
		HalfPageDown: key.NewBinding(
			key.WithKeys("d", "ctrl+d"),
			key.WithHelp("d", "½ page down"),
		),
		GotoTop: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GotoBottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),
		ToggleCellSelect: key.NewBinding(
			key.WithKeys("t", "ctrl+t"),
			key.WithHelp("t", "toggle cell select")),
	}
}

// Styles contains style definitions for this list component. By default, these
// values are generated by DefaultStyles.
type Styles struct {
	Header       lipgloss.Style
	Cell         lipgloss.Style
	Selected     lipgloss.Style
	SelectedCell lipgloss.Style
}

// DefaultStyles returns a set of default style definitions for this table.
func DefaultStyles() Styles {
	return Styles{
		Selected:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")),
		Header:       lipgloss.NewStyle().Bold(true).Padding(0, 1),
		Cell:         lipgloss.NewStyle().Padding(0, 1),
		SelectedCell: lipgloss.NewStyle().Padding(0, 1).Bold(true).Foreground(lipgloss.Color("212")),
	}
}

// SetStyles sets the table styles.
func (m *Model) SetStyles(s Styles) {
	m.styles = s
}

// Option is used to set options in New. For example:
//
//	table := New(WithColumns([]Column{{Title: "ID", Width: 10}}))
type Option func(*Model)

// New creates a new model for the table widget.
func New(opts ...Option) Model {
	m := Model{
		row: 0,
		col: 0,
		view: tableModel{
			XOffset: 0,
			YOffset: 0,
			Width:   20,
			Height:  20,
		},
		KeyMap: DefaultKeyMap(),
		styles: DefaultStyles(),
	}

	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// WithColumns sets the table columns (headers).
func WithColumns(cols []Column) Option {
	return func(m *Model) {
		m.cols = cols
	}
}

// WithRows sets the table rows (data).
func WithRows(rows []Row) Option {
	return func(m *Model) {
		m.rows = rows
	}
}

// WithHeight sets the height of the table.
func WithHeight(h int) Option {
	return func(m *Model) {
		m.view.Height = h
	}
}

// WithWidth sets the width of the table.
func WithWidth(w int) Option {
	return func(m *Model) {
		m.view.Width = w
	}
}

// WithFocused sets the focus state of the table.
func WithFocused(f bool) Option {
	return func(m *Model) {
		m.focus = f
	}
}

// WithStyles sets the table styles.
func WithStyles(s Styles) Option {
	return func(m *Model) {
		m.styles = s
	}
}

// WithKeyMap sets the key map.
func WithKeyMap(km KeyMap) Option {
	return func(m *Model) {
		m.KeyMap = km
	}
}

// Update is the Bubble Tea update loop.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}

	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	// TODO: mouse support is only easy to do when in ALT mode
	// there may be a PR in the future to add both at the same time as an option
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.LineUp):
			m.MoveUp(1)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.LineLeft):
			m.MoveLeft(1)
		case key.Matches(msg, m.KeyMap.LineRight):
			m.MoveRight(1)
		case key.Matches(msg, m.KeyMap.PageUp):
			m.MoveUp(m.view.Height)
		case key.Matches(msg, m.KeyMap.PageDown):
			m.MoveDown(m.view.Height)
		case key.Matches(msg, m.KeyMap.HalfPageUp):
			m.MoveUp(m.view.Height / 2)
		case key.Matches(msg, m.KeyMap.HalfPageDown):
			m.MoveDown(m.view.Height / 2)
		case key.Matches(msg, m.KeyMap.LineDown):
			m.MoveDown(1)
		case key.Matches(msg, m.KeyMap.GotoTop):
			m.GotoTop()
		case key.Matches(msg, m.KeyMap.GotoBottom):
			m.GotoBottom()
		case key.Matches(msg, m.KeyMap.ToggleCellSelect):
			m.ToggleCellSelect()
		}
	}

	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Focused returns the focus state of the table.
func (m *Model) Focused() bool {
	return m.focus
}

// Focus focusses the table, allowing the user to move around the rows and
// interact.
func (m *Model) Focus() {
	m.focus = true
}

// Blur blurs the table, preventing selection or movement.
func (m *Model) Blur() {
	m.focus = false
}

// View renders the component.
func (m *Model) View() string {
	builder := strings.Builder{}
	completedSections := make(chan bool, 2)
	var (
		header string
		body   string
	)

	go func(h *string) {
		*h = m.headersView()
		completedSections <- true
	}(&header)

	go func(b *string) {
		*b = m.bodyView()
		completedSections <- true
	}(&body)

	<-completedSections
	<-completedSections

	builder.WriteString(header + "\n")
	builder.WriteString(body)
	return builder.String()
}

func (m *Model) ToggleCellSelect() {
	m.selectCell = !m.selectCell
}

// SelectedRow returns the selected row.
// You can cast it to your own implementation.
func (m *Model) SelectedRow() Row {
	return m.rows[m.row]
}

func (m *Model) SelectedCell() string {
	if m.selectCell {
		return m.rows[m.row][m.col]
	}

	return ""
}

// SetRows set a new rows state.
func (m *Model) SetRows(r []Row) {
	m.rows = r
}

// SetWidth sets the width of the viewport of the table.
func (m *Model) SetWidth(w int) {
	m.view.Width = w
}

// SetHeight sets the height of the viewport of the table.
func (m *Model) SetHeight(h int) {
	m.view.Height = h
}

// Height returns the viewport height of the table.
func (m *Model) Height() int {
	return m.view.Height
}

// Width returns the viewport width of the table.
func (m *Model) Width() int {
	return m.view.Width
}

// RowIndex returns the index of the selected row.
func (m *Model) Cursor() int {
	return clamp(m.row+m.view.YOffset, 0, len(m.rows)-1)
}

func (m *Model) RowIndex() int {
	return clamp(m.row+m.view.YOffset, 0, len(m.rows)-1)
}

func (m *Model) SetRowIndex(n int) {
	m.row = clamp(n, 0, len(m.rows)-1)
}

// SetCursor sets the cursor position in the table.
func (m *Model) SetCursor(n int) {
	m.row = clamp(n, 0, len(m.rows)-1)
}

func (m *Model) ColIndex() int {
	return clamp(m.col+m.view.XOffset, 0, len(m.rows[0])-1)
}

func (m *Model) SetColIndex(n int) {
	m.col = clamp(n, 0, len(m.rows[0])-1)
}

// MoveUp moves the selection up by any number of row.
// It can not go above the first row.
func (m *Model) MoveUp(n int) {
	m.row = clamp(m.row-n, 0, len(m.rows)-1)

	if m.row < m.view.YOffset {
		m.view.YOffset = m.row
	}
}

func (m *Model) MoveLeft(n int) {
	if m.selectCell {
		m.col = clamp(m.col-n, 0, len(m.rows[0])-1)

		if m.col < m.view.XOffset {
			m.view.XOffset = m.col
		}
	} else {
		if m.view.XOffset > 0 {
			m.view.XOffset -= 1
		}
	}
}

// MoveDown moves the selection down by any number of row.
// It can not go below the last row.
func (m *Model) MoveDown(n int) {
	m.row = clamp(m.row+n, 0, len(m.rows)-1)

	if m.row > (m.view.YOffset + (m.view.Height - 1)) {
		m.view.YOffset = m.row - (m.view.Height - 1)
	}
}

func (m *Model) MoveRight(n int) {
	if m.selectCell {
		// rather big assumption that all rows will have same number of elements
		m.col = clamp(m.col+n, 0, len(m.rows[0])-1)

		if m.col > (m.view.XOffset + (m.view.Width - 1)) {
			m.view.XOffset = m.col - (m.view.Width - 1)
		}
	} else {
		m.view.XOffset = clamp(m.view.XOffset+n, 0, len(m.rows[0])-m.view.Width)
	}
}

// GotoTop moves the selection to the first row.
func (m *Model) GotoTop() {
	m.MoveUp(m.row + m.view.YOffset)
}

// GotoBottom moves the selection to the last row.
func (m *Model) GotoBottom() {
	m.MoveDown(len(m.rows))
}

// FromValues create the table rows from a simple string. It uses `\n` by
// default for getting all the rows and the given separator for the fields on
// each row.
func (m *Model) FromValues(value, separator string) {
	var rows []Row
	for _, line := range strings.Split(value, "\n") {
		r := Row{}
		for _, field := range strings.Split(line, separator) {
			r = append(r, field)
		}
		rows = append(rows, r)
	}

	m.SetRows(rows)
}

func (m *Model) headersView() string {
	var s = make([]string, len(m.cols))

	cell := 0
	for _, col := range m.cols[m.view.XOffset:clamp(m.view.XOffset+m.view.Width, 0, len(m.rows[0]))] {
		style := lipgloss.NewStyle().Width(col.Width).MaxWidth(col.Width).Inline(true)
		renderedCell := style.Render(runewidth.Truncate(col.Title, col.Width, "…"))
		s[cell] = m.styles.Header.Render(renderedCell)
		cell++
	}

	return lipgloss.JoinHorizontal(lipgloss.Left, s...)
}

func (m *Model) bodyView() string {
	builder := strings.Builder{}
	for i := m.view.YOffset; i < m.view.YOffset+m.view.Height; i++ {
		builder.WriteString(m.renderRow(i) + "\n")
	}

	return builder.String()
}

func (m *Model) renderRow(rowID int) string {
	var s = make([]string, len(m.cols))
	cell := 0
	for i, value := range m.rows[rowID][m.view.XOffset:clamp(m.view.XOffset+m.view.Width, 0, len(m.rows[0]))] {
		width := m.cols[i+m.view.XOffset].Width
		style := lipgloss.NewStyle().Width(width).MaxWidth(width).Inline(true)
		var renderedCell string
		if rowID == m.row && m.col == cell+m.view.XOffset && m.selectCell {
			renderedCell = m.styles.Selected.Padding(0, 1).Render(style.Render(runewidth.Truncate(value, width, "…")))
		} else {
			renderedCell = m.styles.Cell.Render(style.Render(runewidth.Truncate(value, width, "…")))
		}

		s[cell] = renderedCell
		cell++
	}

	row := lipgloss.JoinHorizontal(lipgloss.Left, s...)

	if rowID == m.row && !m.selectCell {
		return m.styles.Selected.Render(row)
	}

	return row
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func clamp(v, low, high int) int {
	return min(max(v, low), high)
}
