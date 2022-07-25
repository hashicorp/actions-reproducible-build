package commands

import (
	"log"

	"github.com/hashicorp/actions-go-build/pkg/build"
	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
	cp "github.com/otiai10/copy"
)

// Verification runs the verification build, first copying the primary build
// directory to the verification build root.
var Verification = cli.LeafCommand("verification", "run the verification build", func(c *opts.VerificationBuildOpts) error {
	result, err := runVerificationBuild(c.PrimaryBuildRoot, c.VerificationBuildRoot, c.Build)
	if err != nil {
		return err
	}

	if err := writeJSON(stdout, result); err != nil {
		return err
	}
	return result.Error()
})

func runVerificationBuild(primaryBuildRoot, verificationBuildRoot string, verificationBuild build.Build) (crt.BuildResult, error) {
	log.Printf("Running verification build")
	log.Printf("Copying %s to %s", primaryBuildRoot, verificationBuildRoot)
	if err := cp.Copy(primaryBuildRoot, verificationBuildRoot); err != nil {
		return crt.BuildResult{}, err
	}
	return verificationBuild.Run(), nil
}
