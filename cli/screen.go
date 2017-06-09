package cli

import (
	"bytes"
	"errors"
	"strings"

	"github.com/b4b4r07/history/config"
	"github.com/b4b4r07/history/history"
)

type Screen struct {
	Lines   []string
	Records []history.Record
}

func NewScreen() (s *Screen, err error) {
	c := config.Conf.Screen
	v := config.Conf.Var
	var (
		lines   []string
		records history.Records
	)

	h, err := history.Load()
	if err != nil {
		return
	}

	h.Records.Sort()
	h.Records.Reverse()
	h.Records.Unique()

	if v.Query != "" {
		h.Records.Contains(v.Query)
	}

	columns := []string{}
	if v.Columns != "" {
		columns = strings.Split(v.Columns, ",")
		if len(columns) > 0 {
			config.Conf.Screen.Columns = columns
		}
	}

	if idx := config.IndexCommandColumns(); idx == -1 {
		if len(config.Conf.Screen.Columns) > 0 {
			// Other elements are specified although {{.Command}} is not specified in column
			err = errors.New("Error: {{.Command}} tepmplete should be contained in columns")
			return
		}
	}

	for _, record := range h.Records {
		if c.FilterDir && v.Dir != record.Dir {
			continue
		}
		if c.FilterBranch && v.Branch != record.Branch {
			continue
		}
		if c.FilterHostname && v.Hostname != record.Hostname {
			continue
		}
		lines = append(lines, record.Render())
		records = append(records, record)
	}

	return &Screen{
		Lines:   lines,
		Records: records,
	}, nil
}

type Line struct {
	history.Record
}

type Lines []Line

func (s *Screen) parseLine(line string) (*Line, error) {
	l := strings.Split(line, "\t")
	var record history.Record
	idx := config.IndexCommandColumns()
	if idx == -1 {
		// default
		idx = 0
	}
	for _, record = range s.Records {
		if record.Command == l[idx] {
			break
		}
	}
	return &Line{record}, nil
}

func (l *Lines) Filter(fn func(Line) bool) *Lines {
	lines := make(Lines, 0)
	for _, line := range *l {
		if fn(line) {
			lines = append(lines, line)
		}
	}
	return &lines
}

func (s *Screen) Select() (lines Lines, err error) {
	if len(s.Lines) == 0 {
		err = errors.New("no text to display")
		return
	}
	selectcmd := config.Conf.Core.SelectCmd
	if selectcmd == "" {
		err = errors.New("no selectcmd specified")
		return
	}

	text := strings.NewReader(strings.Join(s.Lines, "\n"))
	var buf bytes.Buffer
	err = runFilter(selectcmd, text, &buf)
	if err != nil {
		return
	}

	if buf.Len() == 0 {
		err = errors.New("no lines selected")
		return
	}

	selectedLines := strings.Split(buf.String(), "\n")
	for _, line := range selectedLines {
		if line == "" {
			continue
		}
		parsedLine, err := s.parseLine(line)
		if err != nil {
			return lines, err
		}
		lines = append(lines, *parsedLine)
	}

	if len(lines) == 0 {
		err = errors.New("no lines selected")
		return
	}

	return
}
