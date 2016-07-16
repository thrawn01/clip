package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/thrawn01/clip/pkg"
)

func run(buf *string, name string, args ...string) error {
	var err error
	var output []byte
	if output, err = exec.Command(name, args...).Output(); err != nil {
		return errors.Wrap(err, "error running 'git config'")
	} else {
		*buf = string(output)
		return nil
	}
}

func trackedBranches(result pkg.TrackedBranchMap) error {
	var output string
	// Using git config list all the branch entries
	if err := run(&output, "git", "config", "--get-regexp", "^branch\\."); err != nil {
		return err
	}
	return pkg.ParseTrackedBranches(output, result)
}

func main() {
	tracked := pkg.TrackedBranchMap{}

	// List Tracked Branches
	if err := trackedBranches(tracked); err != nil {
		fmt.Sprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Printf("%+v\n", tracked)

	// List All Branches

	// Organize all the available information on our branches

	// Display this information for the user

	// TODO: Remove Hello World\n
	fmt.Printf("Hello World\n")
}
