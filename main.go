package main

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type replace struct {
	search       *regexp.Regexp
	repl         string
	toLowerIndex int
}

func main() {

	drives, err := getDrives()
	if err != nil {
		println("Unable to get the drive list.")
	}

	wslDrives := filterWslDrives(drives)
	notWslDrives := filterNotWslDrives(drives)

	replaces := []replace{
		{
			regexp.MustCompile("^[" + strings.Join(wslDrives, "") + "]:(\\\\.*)$"),
			`$1`,
			-1,
		},
		{
			regexp.MustCompile(`^\\\\wsl\$[a-zA-z0-9 _-]*(\\.*)$`),
			`$1`,
			-1,
		},
		{
			regexp.MustCompile("^([" + strings.Join(notWslDrives, "") + "]):(\\\\.*)$"),
			`/mnt/$1$2`,
			5,
		},
	}

	//command := strings.Join(os.Args[1:], " ")
	wd, err := os.Getwd()
	if err != nil {
		println("Unable to get the working directory.")
		os.Exit(1)
	}

	for _, replace := range replaces {
		wd = string(replace.search.ReplaceAll([]byte(wd), []byte(replace.repl)))
		if replace.toLowerIndex > -1 {
			wd = wd[0:replace.toLowerIndex] + strings.ToLower(string(wd[replace.toLowerIndex])) + wd[replace.toLowerIndex+1:]
		}
	}

	wd = strings.ReplaceAll(wd, "\\", "/")

	command := []string{"--", "cd", wd, ";", "git"}
	command = append(command, os.Args[1:]...)
	cmd := exec.Command("wsl", command...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		println("Unable to execute: " + strings.Join(command, " "))
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		os.Exit(1)
	}

	os.Exit(0)
}

func getDrives() ([]string, error) {
	cmd := exec.Command("wmic", "logicaldisk", "get", "name")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	var drives []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, ":") {
			drives = append(drives, string(line[0]))
		}
	}

	return drives, nil
}

func filterWslDrives(drives []string) []string {
	var wslDrives []string
	for _, drive := range drives {
		// check if the drive corresponding to a WSL filesystem
		// assuming that contains /etc/wsl.conf
		if _, err := os.Stat(drive + ":\\etc\\wsl.conf"); !os.IsNotExist(err) {
			wslDrives = append(wslDrives, drive)
		}
	}

	return wslDrives
}

func filterNotWslDrives(drives []string) []string {
	var notWslDrives []string
	for _, drive := range drives {
		// check if the drive do not corresponding to a WSL filesystem
		// assuming that not contains /etc/wsl.conf
		if _, err := os.Stat(drive + ":\\etc\\wsl.conf"); os.IsNotExist(err) {
			notWslDrives = append(notWslDrives, drive)
		}
	}

	return notWslDrives
}
