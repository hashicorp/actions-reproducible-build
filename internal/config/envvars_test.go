package config

import (
	"os"
	"testing"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/testhelpers/assert"
	"github.com/hashicorp/composite-action-framework-go/pkg/testhelpers/goldenfile"
)

func TestConfig_ExportToGitHubEnv_ok(t *testing.T) {
	goldenfile.Do(t, func(got *os.File) {
		os.Setenv("GITHUB_ENV", got.Name())
		c := standardConfig()
		c.ExportToGitHubEnv()
	})
}

func TestConfig_BuildConfig_ok(t *testing.T) {
	cases := []struct {
		desc   string
		config Config
		root   string
		want   crt.BuildConfig
	}{
		{
			"root",
			testConfig(),
			"/",
			testBuildConfig(),
		},
		{
			"root/blah",
			testConfig(),
			"/blah",
			testBuildConfig(func(bc *crt.BuildConfig) {
				bc.Paths.WorkDir = "/blah"
				bc.Paths.TargetDir = "/blah/dist"
				bc.Paths.BinPath = "/blah/dist/lockbox"
				bc.Paths.ZipPath = "/blah/out/lockbox_1.2.3_linux_amd64.zip"
				bc.Paths.MetaDir = "/blah/meta"
			}),
		},
		{
			"root/blah (overridden zip name)",
			testConfig(func(c *Config) {
				c.ZipName = "blargle.zip"
			}),
			"/blah",
			testBuildConfig(func(bc *crt.BuildConfig) {
				bc.Paths.WorkDir = "/blah"
				bc.Paths.TargetDir = "/blah/dist"
				bc.Paths.BinPath = "/blah/dist/lockbox"
				bc.Paths.ZipPath = "/blah/out/blargle.zip"
				bc.Paths.MetaDir = "/blah/meta"
			}),
		},
	}

	for _, c := range cases {
		desc, config, root, want := c.desc, c.config, c.root, c.want
		t.Run(desc, func(t *testing.T) {
			got, err := config.buildConfig(root)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, got, want)
		})
	}
}

func standardBuildconfig() crt.BuildConfig {
	return crt.BuildConfig{
		Product:    standardProduct(),
		Parameters: standardParameters(),
		Paths: crt.BuildPaths{
			WorkDir:   "/",
			TargetDir: "/dist",
			BinPath:   "/dist/lockbox",
			ZipPath:   "/out/lockbox_1.2.3_linux_amd64.zip",
			MetaDir:   "/meta",
		},
	}
}

func testBuildConfig(modifiers ...func(*crt.BuildConfig)) crt.BuildConfig {
	return applyModifiers(standardBuildconfig(), modifiers...)
}