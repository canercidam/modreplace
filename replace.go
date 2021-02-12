package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/mod/modfile"
)

// ClonedRepo contains cloned repository info.
type ClonedRepo struct {
	Path     string
	Version  string
	Revision string
}

// Replace replace puts replacements to the directory and updates go.mod file.
func Replace(paths []string, file *modfile.File) {
	var cloneRepos []*ClonedRepo
	for _, path := range paths {
		if version, revision, ok := replace(path, file); ok {
			cloneRepos = append(cloneRepos, &ClonedRepo{
				Path:     path,
				Version:  version,
				Revision: revision,
			})
		}
	}

	exec.CommandContext(
		context.Background(),
		"rm", "-rf", "./modreplace",
	).Run()
	for _, cloned := range cloneRepos {
		clone(cloned)
	}
}

func replace(path string, file *modfile.File) (string, string, bool) {
	var foundRequired *modfile.Require
	for _, required := range file.Require {
		if required.Mod.Path == path {
			foundRequired = required
			break
		}
	}
	if foundRequired == nil {
		log.Printf("%s does not exist in go.mod - ignoring", path)
		return "", "", false
	}
	version := foundRequired.Mod.Version

	var foundReplacement *modfile.Replace
	fmt.Println(file.Replace)
	for _, replace := range file.Replace {
		fmt.Println(replace.Old.Path, path)
		if replace.Old.Path == path {
			foundReplacement = replace
			break
		}
	}
	if foundReplacement == nil {
		if err := file.AddReplace(path, "", replacementPath(path), ""); err != nil {
			panic(err)
		}
	}

	var revision string
	switch {
	case len(version) >= 34:
		revision = strings.Split(version, "-")[2][:7]
		version = ""
	case strings.Contains(version, "+incompatible"):
		version = strings.Split(version, "+")[0]
	}

	return version, revision, true
}

func clone(repo *ClonedRepo) {
	path := repo.Path
	repoPath := replacementPath(path)
	dotGitPath := path + "/.git"
	os.MkdirAll(replacementPath(path), os.ModePerm)

	cloneCmd := []string{"git", "clone"}
	if len(repo.Version) > 0 {
		cloneCmd = append(cloneCmd, "-b", repo.Version)
	}
	cloneCmd = append(cloneCmd, fmt.Sprintf("https://%s", path), repoPath)

	err := exec.CommandContext(context.Background(), cloneCmd[0], cloneCmd[1:]...).Run()
	if err != nil {
		log.Panicf("failed to clone the repo %s: %v", repoPath, err)
	}

	if len(repo.Revision) > 0 {
		rev := repo.Revision
		command := exec.CommandContext(
			context.Background(),
			"git", "checkout", rev,
		)
		command.Dir = repoPath
		if err := command.Run(); err != nil {
			log.Panicf("failed to checkout %s to %s: %v", path, rev, err)
		}
	}

	exec.CommandContext(
		context.Background(),
		"rm", "-rf", dotGitPath,
	).Run()
}

func replacementPath(path string) string {
	return fmt.Sprintf("./modreplace/%s", path)
}
