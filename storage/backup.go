package storage

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
)

// GITCONFIG is the default contents of the .git/gitconfig file
const GITCONFIG = `
[core]
	repositoryformatversion = 0
	filemode = true
	bare = false
	logallrefupdates = true
[user]
	email = mcmanager@local
	name = mcmanager
`

// GITIGNORE is the default contents of the .gitignore file
const GITIGNORE = `
*~
logs/
`

// gitAvailable returns if the git command is available or not
func gitAvailable() bool {
	_, err := exec.LookPath("git")
	if err == nil {
		return true
	}
	return false
}

// SetupServerBackup initializes a git repo inside an instance dir
func SetupServerBackup(serverID string) error {
	if serverID == "" {
		return fmt.Errorf("cannot setup backups without a server ID")
	}

	var err error
	for err == nil {
		err = gitInit(serverID)
		err = gitConfigs(serverID)
		err = GitCommit(serverID, "initial")
		break
	}
	return err
}

func gitInit(serverID string) error {
	if !gitAvailable() {
		return fmt.Errorf("git not available")
	}

	var cmd = exec.Command("git", "init")
	cmd.Dir = filepath.Join(SERVERDIR, serverID)
	return cmd.Run()
}

func gitConfigs(serverID string) error {
	var err error
	for err == nil {
		err = WriteServerFile(serverID, ".gitignore", []byte(GITIGNORE))
		err = WriteServerFile(serverID, ".git/config", []byte(GITCONFIG))
		break
	}

	return err
}

// GitCommit adds all files in the instance dir and commits them
func GitCommit(serverID, message string) error {
	if !gitAvailable() {
		return fmt.Errorf("git not available")
	}
	var instanceDir = filepath.Join(SERVERDIR, serverID)

	var add = exec.Command("git", "add", "-A")
	add.Dir = instanceDir

	var commit = exec.Command("git", "commit", "-m", message)
	commit.Dir = instanceDir

	var err error
	var out []byte
	for err == nil {
		out, err = add.CombinedOutput()
		out, err = commit.CombinedOutput()
		break
	}

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if status.ExitStatus() == 1 {
					err = nil
				} else {
					fmt.Printf("GitCommit error output: \n%s\n", out)
				}
			}
		}
	}

	return err
}
