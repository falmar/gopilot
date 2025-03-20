package gopilot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/mafredri/cdp/devtool"
)

type Browser interface {
	Open(ctx context.Context, in *BrowserOpenInput) error
	NewPage(ctx context.Context, newTab bool) (Page, error)
	Close(ctx context.Context) error
}

type browser struct {
	config   *BrowserConfig
	logger   *slog.Logger
	instance *exec.Cmd
	datadir  string

	devtool *devtool.DevTools
	pages   []*page
}

func NewBrowser(cfg *BrowserConfig, logger *slog.Logger) Browser {
	return &browser{config: cfg, logger: logger, pages: make([]*page, 0)}
}

type BrowserOpenInput struct{}

func (b *browser) Open(ctx context.Context, in *BrowserOpenInput) error {
	tempDir, err := os.MkdirTemp("", "gopilot")
	if err != nil {
		return err
	}
	b.datadir = tempDir
	b.logger.Debug("created data dir", "path", b.datadir)

	b.instance = exec.Command(b.config.Path)
	b.instance.Env = b.config.Envs

	b.instance.Args = append(
		b.config.Args,
		fmt.Sprintf("--remote-debugging-port=%s", b.config.DebugPort),
		fmt.Sprintf("--user-data-dir=%s", tempDir),
	)

	dtChan := make(chan string)

	stderr, err := b.instance.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()
	go func() {
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "DevTools listening on") {
				dtChan <- line
			}

			b.logger.Debug("chromesdterr", "msg", line)
		}
	}()

	err = b.instance.Start()
	if err != nil {
		return err
	}

	dtString := strings.Split(<-dtChan, "DevTools listening on")
	if len(dtString) < 2 {
		return errors.New("unable to obtain dev tool url")
	}

	b.logger.Debug("listen on", "url", dtString[1])

	b.logger.Debug("creating devtool")
	b.devtool = devtool.New(fmt.Sprintf("http://127.0.0.1:%s", b.config.DebugPort))

	return nil
}

func (b *browser) NewPage(ctx context.Context, newTab bool) (Page, error) {
	p, err := newPage(
		ctx,
		b.devtool,
		b.logger,
		newTab,
	)
	if err != nil {
		return nil, err
	}

	b.pages = append(b.pages, p.(*page))

	return p, nil
}

func (b *browser) Close(ctx context.Context) error {
	b.logger.Debug("closing pages", "len", len(b.pages))
	for _, p := range b.pages {
		if p.closed {
			b.logger.Debug("page already closed", "target_id", p.target.ID)
			continue
		}

		b.logger.Debug("closing page", "target_id", p.target.ID)
		err := p.Close(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}

	b.devtool = nil
	b.pages = nil

	err := b.instance.Process.Signal(os.Interrupt)
	if err != nil {
		return err
	}

	defer func() {
		if _, err = os.Stat(b.datadir); !os.IsNotExist(err) {
			err = os.RemoveAll(b.datadir)
			if err != nil {
				b.logger.Warn("removing data dir error", "path", b.datadir, "error", err)
			} else {
				b.logger.Debug("removed data dir", "path", b.datadir)
			}
		}
	}()

	return b.instance.Wait()
}
