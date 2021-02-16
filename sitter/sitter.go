package sitter

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"

	"github.com/armon/circbuf"
	"github.com/carlmjohnson/flagext"
)

const AppName = "kristy"

func CLI(args []string) error {
	var app appEnv
	err := app.ParseArgs(args)
	if err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err = app.Exec(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

func (app *appEnv) ParseArgs(args []string) error {
	fl := flag.NewFlagSet(AppName, flag.ContinueOnError)
	// Likely sufficient for anyone's purposes
	fl.Int64Var(&app.bufSize, "buffer-size", 640*1<<10, "byte size for output buffer")
	app.Logger = log.New(nil, AppName+" ", log.LstdFlags)
	flagext.LoggerVar(
		fl, app.Logger, "silent", flagext.LogSilent, "don't log debug output")
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), `kristy - babysit

Usage:

	kristy [options] <command to babysit>

Options:
`)
		fl.PrintDefaults()
		fmt.Fprintln(fl.Output(), "")
	}
	if err := fl.Parse(args); err != nil {
		return err
	}
	if err := flagext.ParseEnv(fl, AppName); err != nil {
		return err
	}
	if err := flagext.MustHaveArgs(fl, 1, -1); err != nil {
		return err
	}
	app.cmd = fl.Args()
	return nil
}

type appEnv struct {
	bufSize int64
	cmd     []string
	*log.Logger
	// TK Slack, Healthcheck, Sentry
}

func (app *appEnv) Exec(ctx context.Context) (err error) {
	app.Println("starting")
	defer app.Println("done")
	// Tell HC we started
	// Errors to Sentry then Slack

	cmd := exec.CommandContext(ctx, app.cmd[0], app.cmd[1:]...)
	buf, _ := circbuf.NewBuffer(app.bufSize)
	cmd.Stdout = io.MultiWriter(buf, os.Stdout)
	cmd.Stderr = io.MultiWriter(buf, os.Stderr)
	// TODO: Handle errors
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("could not start process: %w", err)
	}

	if err = cmd.Wait(); err != nil {
		// report failure to Sentry
		// fallback to HC then Slack
		return fmt.Errorf("process failed: %w", err)
	}
	// Tell HC we succeeded,
	// errors to Sentry then Slack
	return err
}
