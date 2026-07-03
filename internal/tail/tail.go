package tail

import (
	"bufio"
	"context"
	"io"
	"os"
	"time"

	"github.com/deepawasthi/logtimeline/internal/config"
)

type Source struct {
	Path string
	cfg  config.TailConfig
	stop chan struct{}
}

func NewSource(path string, cfg config.TailConfig) *Source {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 500 * time.Millisecond
	}
	return &Source{Path: path, cfg: cfg, stop: make(chan struct{})}
}

func (s *Source) Lines(ctx context.Context) (<-chan string, <-chan error) {
	lines := make(chan string, 1024)
	errs := make(chan error, 1)
	go func() {
		defer close(lines)
		defer close(errs)
		file, err := os.Open(s.Path)
		if err != nil {
			errs <- err
			return
		}
		defer file.Close()
		if s.cfg.FromEnd {
			if _, err := file.Seek(0, io.SeekEnd); err != nil {
				errs <- err
				return
			}
		}
		reader := bufio.NewReader(file)
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stop:
				return
			default:
			}
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				lines <- trimLine(line)
			}
			if err == nil {
				continue
			}
			if err != io.EOF {
				errs <- err
				return
			}
			time.Sleep(s.cfg.PollInterval)
		}
	}()
	return lines, errs
}

func (s *Source) Stop(ctx context.Context) error {
	select {
	case <-s.stop:
	default:
		close(s.stop)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func trimLine(line string) string {
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	return line
}
