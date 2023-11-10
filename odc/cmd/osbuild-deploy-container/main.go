package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/osbuild/images/pkg/arch"
	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/manifest"
	"github.com/osbuild/images/pkg/ostree"
	"github.com/osbuild/images/pkg/rpmmd"
)

var reposStr string

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func check(err error) {
	if err != nil {
		fail(err.Error())
	}
}

type BuildConfig struct {
	Name      string               `json:"name"`
	OSTree    *ostree.ImageOptions `json:"ostree,omitempty"`
	Blueprint *blueprint.Blueprint `json:"blueprint,omitempty"`
	Depends   interface{}          `json:"depends,omitempty"` // ignored
}

// Parse embedded repositories and return repo configs for the given
// architecture.
func loadRepos(archName string) []rpmmd.RepoConfig {
	var repoData map[string][]rpmmd.RepoConfig
	err := json.Unmarshal([]byte(reposStr), &repoData)
	if err != nil {
		fail(fmt.Sprintf("error loading repositories: %s", err))
	}
	archRepos, ok := repoData[archName]
	if !ok {
		fail(fmt.Sprintf("no repositories defined for %s", archName))
	}
	return archRepos
}

func loadConfig(path string) BuildConfig {
	fp, err := os.Open(path)
	check(err)
	defer fp.Close()

	dec := json.NewDecoder(fp)
	dec.DisallowUnknownFields()
	var conf BuildConfig

	check(dec.Decode(&conf))
	if dec.More() {
		fail(fmt.Sprintf("multiple configuration objects or extra data found in %q", path))
	}
	return conf
}

func makeManifest(config BuildConfig, repos []rpmmd.RepoConfig, architecture arch.Arch, seedArg int64, cacheRoot string) (manifest.OSBuildManifest, error) {
	return manifest.OSBuildManifest{}, nil
}

func main() {
	hostArch := arch.Current()
	repos := loadRepos(hostArch.String())

	var outputDir, osbuildStore, rpmCacheRoot, configFile, imgref string
	flag.StringVar(&outputDir, "output", ".", "artifact output directory")
	flag.StringVar(&osbuildStore, "store", ".osbuild", "osbuild store for intermediate pipeline trees")
	flag.StringVar(&rpmCacheRoot, "rpmmd", "/var/cache/osbuild/rpmmd", "rpm metadata cache directory")
	flag.StringVar(&configFile, "config", "", "build config file")
	flag.StringVar(&imgref, "imgref", "", "container image to deploy")

	flag.Parse()

	if err := os.MkdirAll(outputDir, 0777); err != nil {
		fail(fmt.Sprintf("failed to create target directory: %s", err.Error()))
	}

	config := BuildConfig{
		Name: "empty",
	}
	if configFile != "" {
		config = loadConfig(configFile)
	}

	seedArg := int64(0)

	fmt.Printf("Generating manifest for %s: ", config.Name)
	mf, err := makeManifest(config, repos, hostArch, seedArg, rpmCacheRoot)
	if err != nil {
		check(err)
	}
	fmt.Printf("%+v\n", mf)

}