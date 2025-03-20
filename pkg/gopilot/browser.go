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
	"time"

	"github.com/mafredri/cdp/devtool"
)

type Browser interface {
	Open(ctx context.Context, in *BrowserOpenInput) error
	NewPage(ctx context.Context, in *BrowserNewPageInput) (*BrowserNewPageOutput, error)
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

	b.logger.Debug("waiting for devtool url message")
	var waitErrorChan = make(chan error) // TODO: close this channel
	go func() {
		waitErrorChan <- b.instance.Wait()
	}()

	// TODO: find a better way to know exec command hangs or go defunct
	waitDuration := time.Second * 5
	var devtoolsURL string
	select {
	case err := <-waitErrorChan:
		return fmt.Errorf("exec wait exited unexpectedtly or too soon: %w", err)

	case <-time.NewTimer(waitDuration).C:
		return fmt.Errorf("duration %s exceeded waiting for devtool url", waitDuration)

	// successful case
	case dtMessage := <-dtChan:
		dtSplit := strings.Split(dtMessage, "DevTools listening on")
		if len(dtSplit) < 2 {
			return errors.New("unable to obtain dev tool url")
		}
		devtoolsURL = dtSplit[1]
	}

	b.logger.Debug("creating devtool", "url", devtoolsURL)
	b.devtool = devtool.New(fmt.Sprintf("http://127.0.0.1:%s", b.config.DebugPort))

	return nil
}

type BrowserNewPageInput struct {
	NewTab bool
}
type BrowserNewPageOutput struct {
	Page Page
}

func (b *browser) NewPage(ctx context.Context, in *BrowserNewPageInput) (*BrowserNewPageOutput, error) {
	p, err := newPage(
		ctx,
		b.devtool,
		b.logger,
		in.NewTab,
	)
	if err != nil {
		return nil, err
	}

	b.pages = append(b.pages, p.(*page))

	return &BrowserNewPageOutput{Page: p}, nil
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
