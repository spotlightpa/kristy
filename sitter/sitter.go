package sitter

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"

	"github.com/armon/circbuf"
	"github.com/carlmjohnson/errutil"
	"github.com/carlmjohnson/exitcode"
	"github.com/carlmjohnson/flagext"
	"github.com/carlmjohnson/slackhook"
	"github.com/spotlightpa/kristy/healthcheckio"
)

const appName = "kristy"

// CLI for the kristy application
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
	fl := flag.NewFlagSet(appName, flag.ContinueOnError)
	app.Logger = log.New(nil, appName+" ", log.LstdFlags)
	flagext.LoggerVar(
		fl, app.Logger, "silent", flagext.LogSilent, "don't log debug output")
	fl.DurationVar(&app.cl.Timeout, "timeout", 10*time.Second, "timeout for HTTP requests")
	fl.Func("slack", "Slack hook `URL`", app.setSlackHook)
	fl.Func("healthcheck", "`UUID` for HealthCheck.io job", app.setHC)
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), `kristy - a baby-sitter for your cron jobs

Kristy tells HealthCheck.io how your cronjobs are doing. If it can't reach
HealthCheck.io, it falls back to warning Slack that something went wrong.

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
	if err := flagext.ParseEnv(fl, appName); err != nil {
		return err
	}
	if err := flagext.MustHave(fl,
		"healthcheck", "slack",
	); err != nil {
		return err
	}
	if err := flagext.MustHaveArgs(fl, 1, -1); err != nil {
		return err
	}
	app.cmd = fl.Args()
	return nil
}

type appEnv struct {
	cmd []string
	cl  http.Client
	sc  *slackhook.Client
	hc  *healthcheckio.Client
	*log.Logger
}

func (app *appEnv) setSlackHook(s string) error {
	app.sc = slackhook.New(s, &app.cl)
	return nil
}

func (app *appEnv) setHC(s string) error {
	app.hc = healthcheckio.New(s, &app.cl)
	return nil
}

const (
	kb = 1 << 10
	// Ought to be enough for anyone
	maxBuf       = 640 * kb
	maxHCBuff    = 10 * kb
	maxSlackBuff = 40 * kb
)

func (app *appEnv) Exec(ctx context.Context) (err error) {
	app.Println("starting")
	defer app.Println("done")
	// Tell HC we started
	errStart := make(chan error, 1)
	go func() { errStart <- app.hc.Start(ctx) }()
	// Run the command
	stdout, stderr, cmderr := app.work(ctx)
	// Tell HC how that went
	code := exitcode.Get(cmderr)
	msg := makeMessage(stdout, stderr, maxHCBuff)
	statusErr := app.hc.Status(ctx, code, msg)
	// If the HC commands didn't work, fallback to Slack
	if err = errutil.Merge(<-errStart, statusErr); err != nil {
		slackerr := app.sc.PostCtx(ctx, slackhook.Message{
			Text: "Could not report job to Healthcheck.io",
			Attachments: []slackhook.Attachment{
				{
					Title: fmt.Sprintf("problem running cron job %s", app.cmd[0]),
					Color: "#f00",
					Fields: []slackhook.Field{
						{
							Title: "Job output",
							Value: string(makeMessage(stdout, stderr, maxSlackBuff)),
						}}}}})
		err = errutil.Merge(err, slackerr)
	}
	// Report if anything failed
	return errutil.Merge(err, cmderr)
}

func makeMessage(stdout, stderr []byte, limit int) []byte {
	limit -= len("-- stdout --\n\n-- stderr --\n\n")
	if len(stdout)+len(stderr) > limit {
		if len(stdout) < limit/2 {
			stderr = stderr[len(stderr)-limit+len(stdout):]
		} else if len(stderr) < limit/2 {
			stdout = stdout[len(stdout)-limit+len(stderr):]
		} else {
			stderr = stderr[len(stderr)-limit/2:]
			stdout = stdout[len(stdout)-limit/2:]
		}
	}
	return []byte(fmt.Sprintf("-- stdout --\n%s\n-- stderr --\n%s\n", stdout, stderr))
}

func (app *appEnv) work(ctx context.Context) (stdout, stderr []byte, err error) {
	// Errors to Sentry then Slack
	cmd := exec.CommandContext(ctx, app.cmd[0], app.cmd[1:]...)
	bufO, _ := circbuf.NewBuffer(maxBuf)
	cmd.Stdout = io.MultiWriter(bufO, os.Stdout)
	bufE, _ := circbuf.NewBuffer(maxBuf)
	cmd.Stderr = io.MultiWriter(bufE, os.Stderr)
	defer func() {
		stdout = bufO.Bytes()
		stderr = bufE.Bytes()
	}()

	if err = cmd.Start(); err != nil {
		err = fmt.Errorf("could not start process: %w", err)
		return
	}

	if err = cmd.Wait(); err != nil {
		// report failure to Sentry
		// fallback to HC then Slack
		err = fmt.Errorf("process failed: %w", err)
		return
	}
	return
}