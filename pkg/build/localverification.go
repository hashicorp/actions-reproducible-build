package build

import (
	"time"

	cp "github.com/otiai10/copy"
)

// LocalVerification is the local verification build. It is run inside a
// temporary copy of the primary build's root directory.
type LocalVerification struct {
	*core
	primaryRoot string
	startAfter  time.Time
}

func NewLocalVerification(primaryRoot string, startAfter time.Time, cfg Config, options ...Option) (Build, error) {
	core, err := newCore(cfg, options...)
	if err != nil {
		return nil, err
	}
	return &LocalVerification{
		core:        core,
		primaryRoot: primaryRoot,
	}, nil
}

func (lv *LocalVerification) Steps() []Step {

	var sleepTime time.Duration
	now := time.Now()
	if lv.startAfter.After(now) {
		sleepTime = lv.startAfter.Sub(now)
	}

	pre := []Step{
		newStep("copying primary build root dir to temp dir", func() error {
			pPath := lv.primaryRoot
			vPath := lv.Config().Paths.WorkDir
			return cp.Copy(pPath, vPath)
		}),
		newStep("waiting until the stagger time has elapsed", func() error {
			time.Sleep(sleepTime)
			return nil
		}),
	}

	return append(pre, lv.core.Steps()...)
}