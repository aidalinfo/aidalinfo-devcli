package main

import (
	"aidalinfo-copilot/backend"
	"context"
	"fmt"

	"github.com/wailsapp/wails/v2/pkg/runtime"
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
	// Initialiser le contexte pour LogToFrontend
	backend.AppCtxForLogToFrontend = ctx
	// Force la fenêtre à se maximiser sur l'écran courant au démarrage
	runtime.WindowMaximise(ctx)
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
	branch, _ := backend.GetCurrentBranch(path)
	return branch
}

func (a *App) GetBranches(path string) []string {
	return backend.GetBranches(path)
}

func (a *App) GetLastCommits(submodules []string) ([]backend.Commit, error) {
	return backend.GetLastCommits(submodules)
}

// Backend Setup operations
func (a *App) InstallSubmodules(path string, branches []string) error {
	return backend.SubmoduleAction(path, branches...)
}

func (a *App) InstallNpmDependencies(path string, all bool) error {
	return backend.NpmAction(path, all)
}

func (a *App) UpdateGitSubmodules(path string, submodules []string) error {
	return backend.GitUpdateAction(path, submodules)
}

func (a *App) GetDefaultBranch() (string, error) {
	return backend.GetDefaultBranch()
}

// Additional backend operations for tagging and advanced features
func (a *App) CreateTag(repoPath, version, message string) error {
	return backend.CreateTag(repoPath, version, message)
}

func (a *App) GetLastTags(repoPath string) (backend.TagsResult, error) {
	vTags, rcTags, err := backend.GetLastTags(repoPath)
	return backend.TagsResult{VTags: vTags, RcTags: rcTags}, err
}
func (a *App) TagAction(version, message string) error {
	return backend.TagAction(version, message)
}

func (a *App) NpmUpdateAction(path string) error {
	return backend.NpmUpdateAction(path)
}

// Expose DownloadBackupWithCreds to frontend
func (a *App) DownloadBackupWithCreds(creds backend.S3Credentials, s3Path, destPath string) error {
	return backend.DownloadBackupWithCreds(a.ctx, creds, s3Path, destPath)
}

// Update operations
func (a *App) GetCurrentVersion() string {
	return backend.GetCurrentVersion()
}

func (a *App) CheckForUpdates() (*backend.UpdateInfo, error) {
	return backend.CheckForUpdates()
}

func (a *App) PerformUpdate(downloadURL string) error {
	tmpFile, err := backend.DownloadUpdate(downloadURL)
	if err != nil {
		return err
	}
	
	err = backend.PerformUpdate(tmpFile)
	if err != nil {
		return err
	}
	
	runtime.EventsEmit(a.ctx, "update:complete")
	return nil
}

// Expose BackupInfo to frontend
func (a *App) ListBackupsWithCreds(creds backend.S3Credentials, s3Dir string) ([]backend.BackupInfo, error) {
	return backend.ListBackupsWithCreds(a.ctx, creds, s3Dir)
}

// Expose RestoreMongoBackup to frontend
func (a *App) RestoreMongoBackup(creds backend.S3Credentials, s3Path, mongoHost, mongoPort, mongoUser, mongoPassword string) error {
	return backend.RestoreMongoBackup(a.ctx, creds, s3Path, mongoHost, mongoPort, mongoUser, mongoPassword)
}

// Expose RestoreS3Backup to frontend
// wailsjs/go/main/App.d.ts doit être régénéré pour :
// export function RestoreS3Backup(cloudCreds: backend.S3Credentials, localCreds: backend.S3Credentials, s3Path: string, s3Host: string, s3Port: string, s3Region: string, s3UseHttps: boolean): Promise<void>;
func (a *App) RestoreS3Backup(cloudCreds backend.S3Credentials, localCreds backend.S3Credentials, s3Path, s3Host, s3Port, s3Region string, s3UseHttps bool) error {
	return backend.RestoreS3Backup(a.ctx, cloudCreds, localCreds, s3Path, s3Host, s3Port, s3Region, s3UseHttps)
}
