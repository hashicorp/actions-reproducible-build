package build

import (
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/actions-go-build/pkg/crt"
	"github.com/hashicorp/actions-go-build/pkg/digest"
)

type TempDirs struct {
	cacheKey
	kind string
}

type cacheKey struct {
	product    crt.Product
	parameters Parameters
	tool       crt.Tool
}

// TempDirFunc is the function used by this package to get the system temp dir.
// You can override this for testing purposes to get platform-independent paths.
var TempDirFunc = os.TempDir

// CacheKeyFunc can be overridden by tests to generate stable strings.
var CacheKeyFunc = digest.CompoundID

func (ck cacheKey) Key() string { return CacheKeyFunc(ck.product, ck.parameters, ck.tool) }

func newDirsFromConfig(c Config, verification bool) TempDirs {
	if verification {
		return NewVerificationDirs(c.Product, c.Parameters, c.Tool)
	}
	return NewPrimaryDirs(c.Product, c.Parameters, c.Tool)
}

func NewPrimaryDirs(p crt.Product, params Parameters, t crt.Tool) TempDirs {
	return NewTempDirs("primary", p, params, t)
}

func NewVerificationDirs(p crt.Product, params Parameters, t crt.Tool) TempDirs {
	return NewTempDirs("verification", p, params, t)
}

type tempDirs struct {
	Primary, Verification TempDirs
}

func NewTempDirs(kind string, p crt.Product, params Parameters, t crt.Tool) TempDirs {
	assertSourceHash(p)
	key := cacheKey{p, params, t}
	return TempDirs{kind: kind, cacheKey: key}
}

func assertSourceHash(p crt.Product) {
	if p.SourceHash != "" {
		return
	}
	// It's the maintainers' jobs to make sure we don't hit this panic.
	// It's here to avoid writing undiscoverable files to the cache.
	if (p == crt.Product{}) {
		log.Panicf("SourceHash is empty; Product is empty.")
	}
	log.Panicf("SourceHash is empty; Product is nonempty: % #v", p)
}

func (d TempDirs) RemoteBuildRoot(extension ...string) string {
	return d.cacheDir("source", extension...)
}

func (d TempDirs) SourceDownloadDir() string {
	return d.cacheDir("sourcearchive")
}

func (d TempDirs) BuildResultCacheDir(extension ...string) string {
	return d.cacheDir("buildresult", extension...)
}

func (d TempDirs) VerificationResultCachePath(configID, zipName string) string {
	return d.cacheDir("verificationresult", configID, zipName+".json")
}

func (d TempDirs) cacheDir(kind string, extension ...string) string {
	return d.tempDirPath(prefix(extension, "cache", kind, d.product.Repository, d.product.Name, d.product.SourceHash)...)
}

func (d TempDirs) tempDirPath(elem ...string) string {
	return prefixPath(elem, TempDirFunc(), d.tool.Name, d.tool.Version, d.tool.Revision, d.kind, d.Key())
}

func productIDSegments(p crt.Product) []string {
	return []string{p.Repository, p.Name, p.Version.Full}
}

func prefix(slice []string, with ...string) []string { return append(with, slice...) }

func prefixPath(slice []string, with ...string) string {
	return filepath.Join(prefix(slice, with...)...)
}
