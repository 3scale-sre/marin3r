package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	. "github.com/dave/jennifer/jen"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

const (
	sourceModuleName  string = "github.com/envoyproxy/go-control-plane"
	envoyModuleName   string = "github.com/envoyproxy/go-control-plane/envoy"
	contribModuleName string = "github.com/envoyproxy/go-control-plane/contrib"
)

var (
	apiVersion  string
	gomodFile   string
	packagFile  string
	packageName string
)

var cmd = &cobra.Command{
	Use: "gen-pkg-envoy-proto",
	Run: func(cmd *cobra.Command, args []string) {
		generate()
	},
}

func init() {
	// flags
	cmd.Flags().StringVar(&apiVersion, "api-version", "", "Generate imports for proto messages for the specific API version")
	cmd.MarkFlagRequired("api-version")
	cmd.Flags().StringVar(&packagFile, "package-file", "", "Name of the generated package file")
	cmd.MarkFlagRequired("package-file")
	cmd.Flags().StringVar(&gomodFile, "gomod-file", "go.mod", "Location of the go.mod file to extrac go-control-plane release from")
	cmd.Flags().StringVar(&packageName, "package-name", "envoy", "Package name of the generated file")
}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type moduleRevision struct {
	path       string
	ref        string
	tagPrefix  string
	dirPrefix  string
	pseudo     bool
	present    bool
}

func generate() {
	revs := inspectRevisions()

	repo, err := git.Clone(memory.NewStorage(), nil, &git.CloneOptions{
		URL:  "https://" + sourceModuleName,
		Tags: git.AllTags,
	})
	checkIfError(err)

	index := map[string]struct{}{}
	for _, rev := range revs {
		if !rev.present {
			continue
		}

		log.Printf("generating %s packages from %s@%s", rev.dirPrefix, rev.path, rev.ref)
		for _, pkg := range listProtoPackages(repo, rev) {
			index[pkg] = struct{}{}
		}
	}

	list := make([]string, 0, len(index))
	for pkg := range index {
		list = append(list, pkg)
	}
	sort.Strings(list)

	writePackageFile(packagFile, list)
}

func inspectRevisions() []moduleRevision {
	body, err := os.ReadFile(gomodFile)
	checkIfError(err)

	gomod, err := modfile.Parse(gomodFile, body, nil)
	checkIfError(err)

	versions := map[string]string{}
	for _, m := range gomod.Require {
		versions[m.Mod.Path] = m.Mod.Version
	}

	envoy := moduleRevision{
		path:      envoyModuleName,
		tagPrefix: "envoy/",
		dirPrefix: "envoy/",
	}
	contrib := moduleRevision{
		path:      contribModuleName,
		tagPrefix: "contrib/",
		dirPrefix: "contrib/",
	}
	root := moduleRevision{
		path:      sourceModuleName,
		dirPrefix: "",
	}

	if v, ok := versions[envoyModuleName]; ok {
		envoy.present = true
		envoy.ref, envoy.pseudo = parseModuleRef(v, envoy.tagPrefix)
	}
	if v, ok := versions[contribModuleName]; ok {
		contrib.present = true
		contrib.ref, contrib.pseudo = parseModuleRef(v, contrib.tagPrefix)
	}
	if v, ok := versions[sourceModuleName]; ok {
		root.present = true
		root.ref, root.pseudo = parseModuleRef(v, "")
	}

	if !envoy.present && !root.present {
		log.Fatalf("%s module not found in modules file", sourceModuleName)
	}

	revs := []moduleRevision{}
	if envoy.present {
		revs = append(revs, envoy)
	} else {
		revs = append(revs, moduleRevision{
			path:      sourceModuleName,
			ref:       root.ref,
			pseudo:    root.pseudo,
			present:   true,
			dirPrefix: "envoy/",
		})
	}
	if contrib.present {
		revs = append(revs, contrib)
	} else if root.present {
		revs = append(revs, moduleRevision{
			path:      sourceModuleName,
			ref:       root.ref,
			pseudo:    root.pseudo,
			present:   true,
			dirPrefix: "contrib/",
		})
	}

	return revs
}

func parseModuleRef(version, tagPrefix string) (string, bool) {
	exp := regexp.MustCompile(`.*-([0-9a-f]{12})$`)
	if match := exp.FindStringSubmatch(version); len(match) != 0 {
		return match[1], true
	}

	return tagPrefix + version, false
}

func listProtoPackages(repo *git.Repository, rev moduleRevision) []string {
	var revision plumbing.Revision
	if rev.pseudo {
		revision = plumbing.Revision(rev.ref)
	} else {
		revision = plumbing.Revision(plumbing.NewTagReferenceName(rev.ref))
	}

	log.Printf("revision: %s", revision.String())
	h, err := repo.ResolveRevision(revision)
	checkIfError(err)
	commit, err := repo.CommitObject(*h)
	checkIfError(err)
	tree, err := commit.Tree()
	checkIfError(err)

	index := map[string]struct{}{}
	err = tree.Files().ForEach(func(f *object.File) error {
		dir, filename := filepath.Split(f.Name)
		dir = filepath.Clean(dir)
		if !strings.HasPrefix(dir, rev.dirPrefix) {
			return nil
		}
		if strings.HasSuffix(filename, ".pb.go") && strings.HasSuffix(dir, "/"+apiVersion) {
			index[sourceModuleName+"/"+dir] = struct{}{}
		}
		return nil
	})
	checkIfError(err)

	list := make([]string, 0, len(index))
	for pkg := range index {
		list = append(list, pkg)
	}
	sort.Strings(list)

	return list
}

func writePackageFile(packagePath string, importList []string) {
	f, err := os.OpenFile(packagePath, os.O_CREATE|os.O_RDWR, 0644)
	checkIfError(err)
	defer f.Close()

	// Reset file contents before writing
	f.Truncate(0)
	f.Seek(0, 0)

	pkg := NewFile(packageName)
	pkg.Anon(importList...)

	w := bufio.NewWriter(f)
	err = pkg.Render(w)
	checkIfError(err)
	checkIfError(w.Flush())
}

func checkIfError(err error) {
	if err == nil {
		return
	}

	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("error: %s", err))
	os.Exit(1)
}
