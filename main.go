package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

const authCnfFile = "~/.ssh/authorized_forced_commands.yml"

type authCommands struct {
	Commands []struct {
		Description string   `json:"description"`
		Env         []string `json:"env"`
		Path        string   `json:"path"`
	} `json:"commands"`
	Type string `json:"type"`
}

func checkPermissions(cnfFile string) error {
	user, err := user.Current()
	if err != nil {
		return errors.New("can't get current user")
	}

	uid, _ := strconv.Atoi(user.Uid)
	gid, _ := strconv.Atoi(user.Gid)

	fileStat, err := os.Stat(cnfFile)
	if err != nil {
		return errors.New("can't stat cnf file")
	}

	if fileStat.Mode()&(1<<2) != 0 {
		return errors.New("cnf file should not be accesible by others group")
	}

	fstat := fileStat.Sys().(*syscall.Stat_t)
	if fstat == nil {
		return errors.New("can't get cnf file ownership")
	}

	if uid != int(fstat.Uid) {
		return errors.New("user must be cnf owner")
	}

	if gid != int(fstat.Gid) {
		return errors.New("user group must be cnf group")
	}

	return nil
}

func main() {
	var authCommands authCommands
	var commandIndex int

	isAllowed := false

	if runtime.GOOS == "windows" {
		fmt.Fprintln(os.Stderr, "windows not supported")
		os.Exit(1)
	}

	sshOriginalCommand, exits := os.LookupEnv("SSH_ORIGINAL_COMMAND")
	if !exits || sshOriginalCommand == "" {
		fmt.Fprintln(os.Stderr, "SSH_ORIGINAL_COMMAND not set")
		os.Exit(1)
	}

	expandedCnfPath, err := homedir.Expand(authCnfFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "can't expand cnf path: ", err)
		os.Exit(1)
	}

	err = checkPermissions(expandedCnfPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	cnfFileReader, err := ioutil.ReadFile(expandedCnfPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err, authCnfFile)
		os.Exit(1)
	}

	err = yaml.Unmarshal(cnfFileReader, &authCommands)
	if err != nil {
		fmt.Fprintln(os.Stderr, "conf file not valid")
		os.Exit(1)
	}

	commandFields := strings.Fields(sshOriginalCommand)
	command := commandFields[0]
	commandArgs := commandFields[1:]

	for commandIndex = range authCommands.Commands {
		if authCommands.Commands[commandIndex].Path == command {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		fmt.Fprintln(os.Stderr, "command not allowed: ", command)
		os.Exit(1)
	}

	cmd := exec.Command(command, commandArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if len(authCommands.Commands[commandIndex].Env) > 0 {
		cmd.Env = append(os.Environ(), authCommands.Commands[commandIndex].Env...)
	}

	if err = cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			fmt.Fprintln(os.Stderr, "execution error: ", err)
			os.Exit(exitError.ExitCode())
		}
	}
}
