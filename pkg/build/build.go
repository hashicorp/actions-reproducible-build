package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hashicorp/actions-go-build/internal/zipper"
	"github.com/hashicorp/actions-go-build/pkg/digest"
	"github.com/hashicorp/composite-action-framework-go/pkg/fs"
	"github.com/hashicorp/composite-action-framework-go/pkg/json"
)

// Build represents the build of a single binary.
// It could be a primary build or a verification build, this Build doesn't
// need to know.
type Build interface {
	Run() Result
	Env() []string
	Config() Config
	CachedResult() (Result, bool, error)
}

func resolveBashPath(path string) (string, error) {
	if path == "" {
		path = "bash"
	}
	return exec.LookPath(path)
}

func New(cfg Config, options ...Option) (Build, error) {
	s, err := newSettings(options)
	if err != nil {
		return &build{}, err
	}
	return &build{
		settings: s,
		config:   cfg,
	}, nil
}

type build struct {
	settings Settings
	config   Config
}

func (b *build) Config() Config {
	return b.config
}

func (b *build) log(f string, a ...any) {
	b.settings.logFunc(f, a...)
}

func (b *build) CachedResult() (Result, bool, error) {
	var r Result
	path := b.config.buildResultCachePath()
	exists, err := fs.FileExists(path)
	if err != nil {
		return r, false, err
	}
	if !exists {
		return r, false, nil
	}
	r, err = json.ReadFile[Result](path)
	return r, err == nil, err
}

func (b *build) Run() Result {
	c := b.config
	r := NewRecorder(b, b.log)
	b.log("Starting build process.")

	b.log("Beginning build, rooted at %q", b.config.Paths.WorkDir)

	var productRevisionTimestamp time.Time
	r.AddStep("validating inputs", func() error {
		var err error
		productRevisionTimestamp, err = c.Product.RevisionTimestamp()
		return err
	})

	r.AddStep("creating output directories", b.createDirectories)

	r.AddStep("running build instructions", b.runInstructions)
	r.AddStep("asserting executable written", b.assertExecutableWritten)
	r.AddStep("writing executable digest", func() error {
		if err := r.RecordBin(c.Paths.BinPath); err != nil {
			return err
		}
		return b.writeDigest(c.Paths.BinPath, "bin_digest")
	})
	r.AddStep("setting mtimes", func() error {
		return fs.SetMtimes(c.Paths.TargetDir, productRevisionTimestamp)
	})
	r.AddStep("creating zip file", func() error {
		return zipper.ZipToFile(c.Paths.TargetDir, c.Paths.ZipPath, r.logFunc)
	})
	r.AddStep("writing zip digest", func() error {
		if err := r.RecordZip(c.Paths.ZipPath); err != nil {
			return err
		}
		return b.writeDigest(c.Paths.ZipPath, "zip_digest")
	})

	return r.Run()
}

func (b *build) createDirectories() error {
	c := b.config
	b.log("Creating output directories.")
	return fs.Mkdirs(c.Paths.TargetDir, c.Paths.ZipDir(), c.Paths.MetaDir)
}

func (b *build) assertExecutableWritten() error {
	binExists, err := b.executableWasWritten()
	if err != nil {
		return err
	}
	if !binExists {
		return fmt.Errorf("no file written to BIN_PATH %q", b.config.Paths.BinPath)
	}
	return nil
}

func (b *build) executableWasWritten() (bool, error) {
	return fs.FileExists(b.config.Paths.BinPath)
}

func (b *build) writeDigest(of, named string) error {
	sha, err := digest.FileSHA256Hex(of)
	if err != nil {
		return err
	}

	digestPath := filepath.Join(b.config.Paths.MetaDir, named)

	return fs.WriteFile(digestPath, sha)
}

func (b *build) newCommand(name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(b.settings.context, name, args...)
	cmd.Dir = b.config.Paths.WorkDir
	cmd.Stdout = b.settings.stdout
	cmd.Stderr = b.settings.stderr
	return cmd
}

func (b *build) runCommand(name string, args ...string) error {
	return b.newCommand(name, args...).Run()
}

func (b *build) runInstructions() error {
	path, err := b.writeInstructions()
	if err != nil {
		return err
	}

	b.listInstructions()

	b.log("Running build instructions with environment:")
	env := b.Env()
	for _, e := range b.Env() {
		b.log(e)
	}
	c := b.newCommand(b.settings.bash, path)
	c.Env = os.Environ()
	c.Env = append(c.Env, env...)
	return c.Run()
}

// writeInstructions writes the build instructions to a temporary file
// and returns its path, or an error if writing fails.
func (b *build) writeInstructions() (path string, err error) {
	b.log("Writing build instructions to temp file.")
	return fs.WriteTempFile("actions-go-build.instructions", b.config.Parameters.Instructions)
}

func (b *build) listInstructions() {
	b.log("Listing build instructions...")
	b.log(b.config.Parameters.Instructions)
}
