package commands

import (
	"fmt"

	"github.com/hashicorp/actions-go-build/pkg/commands/opts"
	"github.com/hashicorp/composite-action-framework-go/pkg/cli"
)

// Primary runs the primary build, in the current directory.
var Primary = cli.LeafCommand("primary", "run the primary build", func(b *opts.PrimaryBuild) error {
	result := b.Run()
	if _, err := fmt.Fprint(stdout, result); err != nil {
		return err
	}
	return result.Error()
})
