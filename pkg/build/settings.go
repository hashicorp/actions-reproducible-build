package build

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/hashicorp/actions-go-build/internal/log"
)

// Settings contains settings for running the instructions.
// These are not to be confused with crt.BuildConfig, these settings
// are build-run specific and not part of the _definition_ of the build.
// Don't use this directly, use the With... functions to set
// settings when calling New.
type Settings struct {
	bash      string
	name      string
	context   context.Context
	logFunc   func(string, ...any)
	debugFunc func(string, ...any)
	stdout    io.Writer
	stderr    io.Writer
}

func (s *Settings) Log(f string, a ...any)   { s.logFunc(s.name+": "+f, a...) }
func (s *Settings) Debug(f string, a ...any) { s.debugFunc(s.name+": "+f, a...) }

func newSettings(name string, options []Option) (Settings, error) {
	out := &Settings{}
	for _, o := range options {
		o(out)
	}
	if err := out.setDefaults(); err != nil {
		return Settings{}, err
	}
	return *out, nil
}

func resolveBashPath(path string) (string, error) {
	if path == "" {
		path = "bash"
	}
	return exec.LookPath(path)
}

func (s *Settings) setDefaults() (err error) {
	s.bash, err = resolveBashPath(s.bash)
	if err != nil {
		return err
	}
	if s.context == nil {
		s.context = context.Background()
	}
	if s.logFunc == nil {
		s.logFunc = log.Verbose
	}
	if s.debugFunc == nil {
		s.logFunc = log.Debug
	}
	if s.stdout == nil {
		s.stdout = os.Stderr
	}
	if s.stderr == nil {
		s.stderr = os.Stderr
	}
	return nil
}

// Option represents a function that configures Settings.
type Option func(*Settings)

func WithContext(c context.Context) Option        { return func(s *Settings) { s.context = c } }
func WithLogfunc(f func(string, ...any)) Option   { return func(s *Settings) { s.logFunc = f } }
func WithDebugfunc(f func(string, ...any)) Option { return func(s *Settings) { s.debugFunc = f } }
func WithStdout(w io.Writer) Option               { return func(s *Settings) { s.stdout = w } }
func WithStderr(w io.Writer) Option               { return func(s *Settings) { s.stderr = w } }
