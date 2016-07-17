package main

import (
	"fmt"
	"os"

	"sort"

	"github.com/pkg/errors"
	"github.com/thrawn01/clip/pkg"
)

func listTrackedBranches(result pkg.TrackedBranchMap) error {
	var output string
	// Using git config list all the tracked branch entries
	if err := pkg.Run(&output, "git", "config", "--get-regexp", "^branch\\."); err != nil {
		return err
	}
	return pkg.ParseTrackedBranches(result, output)
}

func listBranchRefs(result map[string]pkg.BranchMap) error {
	var output string
	// Using git show-ref
	if err := pkg.Run(&output, "git", "show-ref"); err != nil {
		return err
	}
	return pkg.ParseBranchRefs(result, output)
}

func aheadBehind(output *string, master, branch string) error {
	var ahead, behind []string
	if err := pkg.CommitsBetween(&ahead, master, branch); err != nil {
		return errors.Wrap(err, "aheadBehind() - ahead")
	}
	if err := pkg.CommitsBetween(&behind, branch, master); err != nil {
		return errors.Wrap(err, "aheadBehind() - ahead")
	}
	*output = fmt.Sprintf(" (%d/%d)", len(ahead), len(behind))
	return nil
}

func sortBranches(details pkg.BranchDetailMap) []string {
	var sortedBranches []string
	for key := range details {
		sortedBranches = append(sortedBranches, key)
	}
	sort.Strings(sortedBranches)
	return sortedBranches
}

func printRemotes(branch *pkg.BranchDetail) {
	for _, remote := range branch.Remotes {
		var commits []string
		fmt.Printf("     %s ", remote.Ref)
		// Commits Behind
		if err := pkg.CommitsBetween(&commits, remote.Sha, branch.Sha); err != nil {
			fmt.Sprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if len(commits) != 0 {
			fmt.Printf("\033[32mis %d commits behind\033[0m\n", len(commits))
			continue
		}
		// Commits Ahead
		if err := pkg.CommitsBetween(&commits, branch.Sha, remote.Sha); err != nil {
			fmt.Sprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if len(commits) != 0 {
			fmt.Printf("\033[31mis %d commits ahead\033[0m\n", len(commits))
			continue
		}
		fmt.Println("")
	}
}

func main() {
	trackedBranches := pkg.TrackedBranchMap{}
	branchRefs := pkg.BranchReferenceMap{}
	details := pkg.BranchDetailMap{}

	// List Tracked Branches
	if err := listTrackedBranches(trackedBranches); err != nil {
		fmt.Sprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// List All Branches organized by remote
	if err := listBranchRefs(branchRefs); err != nil {
		fmt.Sprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Collect all the branch information so it's simple to display
	if err := pkg.MergeBranchDetail(details, branchRefs, trackedBranches); err != nil {
		fmt.Sprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Display a sorted list of branch information to the user
	for _, name := range sortBranches(details) {
		branch := details[name]
		var follow, tracked string

		if branch.Tracked != nil {
			tracked = fmt.Sprintf(" [%s]", branch.Tracked.Remote)
		}
		if name != "master" {
			if err := aheadBehind(&follow, details["master"].Sha, branch.Sha); err != nil {
				fmt.Sprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}
		// Print the branch name and the remote it's tracking
		fmt.Printf("\033[33m%s\033[0m%s%s\n", name, follow, tracked)
		// Print all the remotes associated with this branch
		printRemotes(branch)
	}
}
