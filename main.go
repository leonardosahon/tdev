// Package main
// Handle multiple tmux sessions using config file for better dev environment
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Session struct {
	Name    string   `yaml:"name"`
	Root    string   `yaml:"root"`
	Windows []Window `yaml:"windows"`
}

type Window struct {
	Name  string `yaml:"name"`
	Path  string `yaml:"path"`
	Cmd   string `yaml:"cmd"`
	Panes []Pane `yaml:"panes"`
}

type Pane struct {
	Path string `yaml:"path"`
	Cmd  string `yaml:"cmd"`
	Hor  bool   `yaml:"horizontal,omitempty"`
}

var (
	rootDir string
	dryRun  bool
	runWall []string
)

func expand(path string) string {
	if strings.HasPrefix(path, "~/") {
		usr, _ := user.Current()
		return filepath.Join(usr.HomeDir, path[2:])
	}
	return path
}

func main() {
	args := os.Args[1:]
	argLen := len(args)

	fmt.Println("This utility assumes your tmux index starts at 1 and not 0")

	if argLen == 0 {
		fmt.Println("Usage: tdev [SESSION_FILE.(yaml|yml)]")
		return
	}

	if argLen > 1 && args[1] == "-d" {
		dryRun = true
	}

	runSession(args[0])

	if dryRun {
		log.Printf("Dry Run: %+v\n", runWall)
	}
}

func cmd(name string, arg ...string) *exec.Cmd {
	if dryRun {
		runWall = append(runWall, fmt.Sprintf("$> %v %+v\n", name, arg))
		return &exec.Cmd{
			Err: nil,
		}
	}

	return exec.Command(name, arg...)
}

func runSession(name string) {
	sess := loadSession(name)
	name = sess.Name

	// Check if tmux session exists
	if cmd("tmux", "has-session", "-t", name).Run() == nil {
		attach(name)
		return
	}

	rootDir = expand(sess.Root)

	if rootDir == "" {
		rootDir, _ = os.Getwd()
	}

	// Create session
	err := cmd("tmux", "new-session", "-d", "-s", name, "-c", rootDir, "-n", "main").Run()
	if err != nil {
		log.Fatalf("Session creation failed! %+v\n", err)
	}

	for k, w := range sess.Windows {
		i := k + 1

		if k == 0 {
			err := cmd(
				"tmux",
				"rename-window", "-t",
				fmt.Sprintf("%s:%d", name, i),
				w.Name,
			).Run()
			if err != nil {
				log.Fatalf("Failed to rename first window! %+v\n", err)
			}
		} else {
			path := expand(filepath.Join(rootDir, w.Path))

			if len(w.Panes) > 0 {
				path = expand(filepath.Join(rootDir, w.Panes[0].Path))
			}

			err := cmd(
				"tmux",
				"new-window", "-t",
				fmt.Sprintf("%s:%d", name, i),
				"-n",
				w.Name,
				"-c",
				path,
			).Run()
			if err != nil {
				log.Fatalf("Window creation failed! %+v\n", err)
			}

		}

		if len(w.Panes) > 0 {
			splitWindow(name, i, w)
			continue
		}

		injectCmds(name, i, 1, w.Cmd)
	}

	err = cmd("tmux", "select-window", "-t", fmt.Sprintf("%s:1", name)).Run()
	if err != nil {
		log.Fatalf("Select window failed! %+v\n", err)
	}

	attach(name)
}

func attach(name string) {
	cmd := cmd("tmux", "attach", "-t", name)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	_ = cmd.Run()
}

func splitWindow(session string, index int, w Window) {
	for i, p := range w.Panes {
		i++

		side := "-v"
		if p.Hor {
			side = "-h"
		}

		if i > 1 {
			err := cmd(
				"tmux", "split-window", side,
				"-t", fmt.Sprintf("%s:%d", session, index),
				"-c", expand(filepath.Join(rootDir, p.Path)),
			).Run()
			if err != nil {
				log.Fatalf("Split window failed! %+v\n", err)
			}

		}

		injectCmds(session, index, i, p.Cmd)
	}
}

func injectCmds(session string, win, pane int, cmdStr string) {
	if cmdStr == "" {
		return
	}

	tmuxTarget := fmt.Sprintf("%s:%d.%d", session, win, pane)

	err := cmd(
		"tmux", "send-keys", "-t", tmuxTarget, cmdStr, "C-m",
	).Run()
	if err != nil {
		log.Fatalf("Inject Cmds failed! %+v\n", err)
	}
}

func loadSession(name string) Session {
	data, err := os.ReadFile(name)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var sess Session
	if err := yaml.Unmarshal(data, &sess); err != nil {
		log.Fatalf("Error parsing Session YAML: %+v", err)
	}

	return sess
}
