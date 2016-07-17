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

func listTrackedBranches(result pkg.TrackedBranchMap) error {
	var output string
	// Using git config list all the tracked branch entries
	if err := run(&output, "git", "config", "--get-regexp", "^branch\\."); err != nil {
		return err
	}
	return pkg.ParseTrackedBranches(output, result)
}

func listBranchRefs(result map[string]pkg.BranchMap) error {
	var output string
	// Using git show-ref
	if err := run(&output, "git", "show-ref"); err != nil {
		return err
	}
	return pkg.ParseBranchRefs(output, result)
}

func main() {
	trackedBranches := pkg.TrackedBranchMap{}
	branchRefs := pkg.BranchReferenceMap{}
	branchDetails := pkg.BranchDetailMap{}

	// List Tracked Branches
	if err := listTrackedBranches(trackedBranches); err != nil {
		fmt.Sprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Printf("%+v\n", trackedBranches)

	// List All Branches organized by remote
	if err := listBranchRefs(branchRefs); err != nil {
		fmt.Sprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Collect all the branch information so it's simple to display
	if err := pkg.MergeBranchDetail(branchDetails, branchRefs, trackedBranches); err != nil {
		fmt.Sprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Display this information for the user

	// TODO: Remove Hello World\n
	fmt.Printf("Hello World\n")
}
