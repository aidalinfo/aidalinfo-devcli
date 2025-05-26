package main

import (
	"context"
	"fmt"
	"aidalinfo-cli/backend"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// Backend Git operations
func (a *App) ListSubmodules(path string) ([]string, error) {
	return backend.ListSubmodule(path)
}

func (a *App) CleanSubmodules(submodules []string) ([]string, error) {
	return backend.CleanSubmodule(submodules)
}

func (a *App) GitStatus(submodule string) string {
	return backend.GitStatus(submodule)
}

func (a *App) GetCurrentBranch(path string) string {
	return backend.GetCurrentBranch(path)
}

func (a *App) GetBranches(path string) []string {
	return backend.GetBranches(path)
}

func (a *App) GetLastCommits(submodules []string) ([]backend.Commit, error) {
	return backend.GetLastCommits(submodules)
}

// Backend Setup operations
func (a *App) InstallSubmodules(branches ...string) error {
	return backend.SubmoduleAction(branches...)
}

func (a *App) InstallNpmDependencies(all bool) error {
	return backend.NpmAction(all)
}

func (a *App) UpdateGitSubmodules(submodules []string) error {
	return backend.GitUpdateAction(submodules)
}

func (a *App) GetDefaultBranch() (string, error) {
	return backend.GetDefaultBranch()
}

// Additional backend operations for tagging and advanced features
func (a *App) CreateTag(repoPath, version, message string) error {
	return backend.CreateTag(repoPath, version, message)
}

func (a *App) GetLastTags(repoPath string) ([]string, []string, error) {
	return backend.GetLastTags(repoPath)
}

func (a *App) TagAction(version, message string) error {
	return backend.TagAction(version, message)
}

func (a *App) NpmUpdateAction() error {
	return backend.NpmUpdateAction()
}
