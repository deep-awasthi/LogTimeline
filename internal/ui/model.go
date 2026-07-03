package ui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/deepawasthi/logtimeline/internal/config"
	exportpkg "github.com/deepawasthi/logtimeline/internal/export"
	"github.com/deepawasthi/logtimeline/internal/filters"
	"github.com/deepawasthi/logtimeline/internal/models"
	"github.com/deepawasthi/logtimeline/internal/parser"
	"github.com/deepawasthi/logtimeline/internal/search"
	"github.com/deepawasthi/logtimeline/internal/stats"
	"github.com/deepawasthi/logtimeline/internal/tail"
	"github.com/deepawasthi/logtimeline/internal/timeline"
)

type mode int

const (
	modeList mode = iota
	modeDetail
	modeSearch
	modeFilter
	modeStats
)

type Model struct {
	timeline models.Timeline
	report   parser.Report
	cfg      config.UIConfig
	keys     keyMap
	help     help.Model
	spinner  spinner.Model
	search   textinput.Model
	filter   textinput.Model
	mode     mode
	cursor   int
	offset   int
	width    int
	height   int
	query    string
	filterBy string
	message  string
	live     bool
	store    *timeline.Store
	detector *parser.Detector
	source   *tail.Source
	lines    <-chan string
	errs     <-chan error
	cancel   context.CancelFunc
	err      error
}

func NewModel(tl models.Timeline, report parser.Report, cfg config.UIConfig) Model {
	return baseModel(cfg, tl, report)
}

func NewLiveModel(store *timeline.Store, detector *parser.Detector, source *tail.Source, cfg config.UIConfig) Model {
	ctx, cancel := context.WithCancel(context.Background())
	lines, errs := source.Lines(ctx)
	model := baseModel(cfg, store.Timeline(), parser.Report{Format: parser.FormatText})
	model.live = true
	model.store = store
	model.detector = detector
	model.source = source
	model.lines = lines
	model.errs = errs
	model.cancel = cancel
	return model
}

func baseModel(cfg config.UIConfig, tl models.Timeline, report parser.Report) Model {
	searchBox := textinput.New()
	searchBox.Prompt = "/ "
	searchBox.Placeholder = "search"
	filterBox := textinput.New()
	filterBox.Prompt = "f "
	filterBox.Placeholder = "ERROR, SQL, Kafka, Redis, REST, Exceptions"
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return Model{
		timeline: tl,
		report:   report,
		cfg:      cfg,
		keys:     defaultKeys(),
		help:     help.New(),
		spinner:  sp,
		search:   searchBox,
		filter:   filterBox,
	}
}

func (m Model) Init() tea.Cmd {
	if m.live {
		return tea.Batch(m.spinner.Tick, m.listenLive())
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case liveLineMsg:
		event, ok := m.detector.Parse(string(msg), int64(m.timeline.Events+1))
		if ok {
			m.store.Add(event)
			m.timeline = m.store.Timeline()
		}
		cmds = append(cmds, m.listenLive())
	case liveErrMsg:
		m.err = error(msg)
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case tea.KeyMsg:
		if m.mode == modeSearch {
			if key.Matches(msg, m.keys.enter) {
				m.query = m.search.Value()
				m.mode = modeList
				m.cursor, m.offset = 0, 0
				break
			}
			if key.Matches(msg, m.keys.esc) {
				m.mode = modeList
				break
			}
			var cmd tea.Cmd
			m.search, cmd = m.search.Update(msg)
			cmds = append(cmds, cmd)
			break
		}
		if m.mode == modeFilter {
			if key.Matches(msg, m.keys.enter) {
				m.filterBy = m.filter.Value()
				m.mode = modeList
				m.cursor, m.offset = 0, 0
				break
			}
			if key.Matches(msg, m.keys.esc) {
				m.mode = modeList
				break
			}
			var cmd tea.Cmd
			m.filter, cmd = m.filter.Update(msg)
			cmds = append(cmds, cmd)
			break
		}
		switch {
		case key.Matches(msg, m.keys.quit):
			if m.cancel != nil {
				m.cancel()
			}
			return m, tea.Quit
		case key.Matches(msg, m.keys.down):
			m.move(1)
		case key.Matches(msg, m.keys.up):
			m.move(-1)
		case key.Matches(msg, m.keys.enter):
			if m.mode == modeList && len(m.visible()) > 0 {
				m.mode = modeDetail
			}
		case key.Matches(msg, m.keys.esc):
			m.mode = modeList
		case key.Matches(msg, m.keys.search):
			m.search.Focus()
			m.search.SetValue(m.query)
			m.mode = modeSearch
		case key.Matches(msg, m.keys.filter):
			m.filter.Focus()
			m.filter.SetValue(m.filterBy)
			m.mode = modeFilter
		case key.Matches(msg, m.keys.stats):
			m.mode = modeStats
		case key.Matches(msg, m.keys.export):
			if err := exportpkg.Write("timeline.json", models.Timeline{Requests: m.visible(), Events: m.timeline.Events}, exportpkg.FormatJSON); err != nil {
				m.err = err
			} else {
				m.message = "exported timeline.json"
			}
		case key.Matches(msg, m.keys.refresh):
			m.timeline = m.timeline.Sorted()
		}
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 {
		m.width = 100
	}
	if m.height == 0 {
		m.height = 30
	}
	bodyHeight := m.height - 4
	if bodyHeight < 8 {
		bodyHeight = 8
	}
	var body string
	switch m.mode {
	case modeDetail:
		body = m.detailView(bodyHeight)
	case modeSearch:
		body = panelStyle.Width(m.width).Height(bodyHeight).Render(m.search.View())
	case modeFilter:
		body = panelStyle.Width(m.width).Height(bodyHeight).Render(m.filter.View())
	case modeStats:
		body = panelStyle.Width(m.width).Height(bodyHeight).Render(stats.Compute(m.visible()).String())
	default:
		body = m.listView(bodyHeight)
	}
	status := m.statusBar()
	return lipgloss.JoinVertical(lipgloss.Left, headerStyle.Width(m.width).Render("LogTimeline"), body, status, helpStyle.Width(m.width).Render(m.help.View(m.keys)))
}

func (m Model) listView(height int) string {
	requests := m.visible()
	if len(requests) == 0 {
		return panelStyle.Width(m.width).Height(height).Render("No requests match the current search or filter.")
	}
	rows := make([]string, 0, height)
	end := m.offset + height
	if end > len(requests) {
		end = len(requests)
	}
	for i := m.offset; i < end; i++ {
		req := requests[i]
		cursor := " "
		style := rowStyle
		if i == m.cursor {
			cursor = ">"
			style = selectedRowStyle
		}
		rows = append(rows, style.Render(fmt.Sprintf("%s %-18s %-7s %-8s %3d  %s", cursor, truncate(req.ID, 18), req.Status, models.FormatDuration(req.Duration), len(req.Events), truncate(req.Summary(), max(20, m.width-60)))))
	}
	return panelStyle.Width(m.width).Height(height).Render(strings.Join(rows, "\n"))
}

func (m Model) detailView(height int) string {
	requests := m.visible()
	if len(requests) == 0 {
		return panelStyle.Width(m.width).Height(height).Render("No request selected.")
	}
	req := requests[m.cursor]
	lines := []string{
		titleStyle.Render("Request ID"),
		req.ID,
		"",
		fmt.Sprintf("Status   %s", req.Status),
		fmt.Sprintf("Duration %s", models.FormatDuration(req.Duration)),
		fmt.Sprintf("Group    %s", req.GroupBy),
		"",
	}
	for i, event := range req.Events {
		if i > 0 {
			lines = append(lines, arrowStyle.Render("↓"))
		}
		lines = append(lines, fmt.Sprintf("%s  %-5s %-9s %s", event.Timestamp.Format("15:04:05"), event.Level, event.Kind, event.Message))
		if event.Error != nil && event.Error.Type != "" {
			lines = append(lines, errorStyle.Render("  "+event.Error.Type+" "+event.Error.RootCause))
		}
	}
	if len(lines) > height {
		lines = lines[:height]
	}
	return panelStyle.Width(m.width).Height(height).Render(strings.Join(lines, "\n"))
}

func (m Model) statusBar() string {
	parts := []string{
		fmt.Sprintf("%d requests", len(m.visible())),
		fmt.Sprintf("%d events", m.timeline.Events),
		"format " + string(m.report.Format),
	}
	if m.query != "" {
		parts = append(parts, "search "+m.query)
	}
	if m.filterBy != "" {
		parts = append(parts, "filter "+m.filterBy)
	}
	if m.live {
		parts = append(parts, m.spinner.View()+" live")
	}
	if m.message != "" {
		parts = append(parts, m.message)
	}
	if m.err != nil {
		parts = append(parts, "error "+m.err.Error())
	}
	return statusStyle.Width(m.width).Render(strings.Join(parts, " | "))
}

func (m Model) visible() []models.Request {
	requests := m.timeline.Requests
	if m.query != "" {
		requests = search.Search(requests, m.query)
	}
	if m.filterBy != "" {
		requests = filters.Apply(requests, m.filterBy)
	}
	return requests
}

func (m *Model) move(delta int) {
	requests := m.visible()
	if len(requests) == 0 {
		m.cursor, m.offset = 0, 0
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(requests) {
		m.cursor = len(requests) - 1
	}
	pageSize := m.height - 6
	if pageSize < 5 {
		pageSize = m.cfg.PageSize
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+pageSize {
		m.offset = m.cursor - pageSize + 1
	}
}

type liveLineMsg string
type liveErrMsg error

func (m *Model) listenLive() tea.Cmd {
	return func() tea.Msg {
		select {
		case line, ok := <-m.lines:
			if !ok {
				return nil
			}
			return liveLineMsg(line)
		case err := <-m.errs:
			if err == nil {
				return nil
			}
			return liveErrMsg(err)
		}
	}
}

func truncate(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	if limit <= 3 {
		return value[:limit]
	}
	return value[:limit-3] + "..."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
