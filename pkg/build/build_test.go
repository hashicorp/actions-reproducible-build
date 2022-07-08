package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/hashicorp/actions-go-build/internal/fs"
	"github.com/hashicorp/actions-go-build/internal/get"
	"github.com/hashicorp/actions-go-build/pkg/crt"
)

func must1[T any](t *testing.T, do func() (T, error)) T {
	res, err := do()
	if err != nil {
		t.Fatal(err)
	}
	return res
}

func TestBuild_Run_ok(t *testing.T) {
	dir := testTempDir(t)

	testBuild := must1(t, func() (Build, error) { return New(standardConfig(dir)) })
	b := testBuild.(*build)
	b.createTestProductRepo(t)
	t.Logf("Test dir: %q", dir)
	if err := b.Run(); err != nil {
		t.Fatal(err)
	}
}

const mainDotGo = `
	package main

	import "fmt"

	func main() {
		fmt.Println("hello, world")
	}
`

const goDotMod = `module github.com/dadgarcorp/lockbox

go 1.18
`

// createTestProductRepo creates a test repo and returns its path.
func (b *build) createTestProductRepo(t *testing.T) {
	b.writeTestFile(t, "main.go", mainDotGo)
	b.writeTestFile(t, "go.mod", goDotMod)
	repo, err := get.Init(b.config.WorkDir)
	if err != nil {
		t.Fatal(err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	wt.Add(".")
	wt.Commit("initial commmit", &git.CommitOptions{})
}

// must is a quick way to fail a test depending on if an error is nil or not.
func must(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func (b *build) runTestCommand(t *testing.T, name string, args ...string) {
	must(t, b.runCommand(name, args...))
}

func (b *build) writeTestFile(t *testing.T, name, contents string) {
	name = filepath.Join(b.config.WorkDir, name)
	must(t, fs.WriteFile(name, contents))
}

func testTempDir(t *testing.T) string {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "_")
	f, err := os.MkdirTemp("", name+".*")
	must(t, err)
	must(t, os.Chmod(f, os.ModePerm))
	return f
}

func standardCommitTime() (ts time.Time, rfc3339 string) {
	ts = time.Date(2022, 7, 4, 11, 33, 33, 0, time.UTC)
	rfc3339 = ts.Format(time.RFC3339)
	return
}

func standardConfig(workDir string) crt.BuildConfig {
	_, revisionTimestampRFC3339 := standardCommitTime()
	return crt.BuildConfig{
		Product: crt.Product{
			Repository:   "dadgarcorp/lockbox",
			Name:         "lockbox",
			Version:      "1.2.3",
			Revision:     "cabba9e",
			RevisionTime: revisionTimestampRFC3339,
		},
		ProductVersionMeta: "",
		WorkDir:            workDir,
		TargetDir:          filepath.Join(workDir, "dist"),
		BinPath:            filepath.Join(workDir, "dist", "lockbox"),
		ZipDir:             filepath.Join(workDir, "out"),
		ZipPath:            filepath.Join(workDir, "out", "lockbox_1.2.3_amd64.zip"),
		MetaDir:            filepath.Join(workDir, "meta"),
		Instructions:       `echo -n "Building '$BIN_PATH'..." && go build -o $BIN_PATH && echo "Done!"`,
		TargetOS:           "linux",
		TargetArch:         "amd64",
	}
}