package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/thrawn01/clip"
)

var (
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).PrintfFunc()
	green  = color.New(color.FgGreen).PrintfFunc()
)

func aheadBehind(output *string, master, branch string) error {
	var ahead, behind []string
	if err := clip.CommitsBetween(&ahead, master, branch); err != nil {
		return errors.Wrap(err, "aheadBehind() - ahead")
	}
	if err := clip.CommitsBetween(&behind, branch, master); err != nil {
		return errors.Wrap(err, "aheadBehind() - ahead")
	}
	*output = fmt.Sprintf(" (%d/%d)", len(ahead), len(behind))
	return nil
}

func sortBranches(details clip.BranchDetailMap) []string {
	var sortedBranches []string
	for key := range details {
		sortedBranches = append(sortedBranches, key)
	}
	sort.Strings(sortedBranches)
	return sortedBranches
}

func printRemotes(branch *clip.BranchDetail) {
	for _, remote := range branch.Remotes {
		if remote == nil {
			continue
		}

		var commits []string
		fmt.Printf("     %s ", remote.Ref)
		// Commits Behind
		if err := clip.CommitsBetween(&commits, remote.Sha, branch.Sha); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if len(commits) != 0 {
			green("is %d commits behind\n", len(commits))
			continue
		}
		// Commits Ahead
		if err := clip.CommitsBetween(&commits, branch.Sha, remote.Sha); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if len(commits) != 0 {
			red("is %d commits ahead\n", len(commits))
			continue
		}
		fmt.Println("")
	}
}

func main() {
	trackedBranches := clip.TrackedBranchMap{}
	branchRefs := clip.BranchReferenceMap{}
	details := clip.BranchDetailMap{}

	// List Tracked Branches
	if err := clip.ListTrackedBranches(trackedBranches); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// List All Branches organized by remote
	if err := clip.ListBranchRefs(branchRefs); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Collect all the branch information so it's simple to display
	if err := clip.MergeBranchDetail(details, branchRefs, trackedBranches); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Display a sorted list of branch information to the user
	for _, name := range sortBranches(details) {
		branch := details[name]
		var follow, tracked string

		if branch.Tracked != nil {
			tracked = fmt.Sprintf(" [%s]", branch.Tracked.Remote)
		}
		if name != "_trunk_" {
			if err := aheadBehind(&follow, details["_trunk_"].Sha, branch.Sha); err != nil {
				fmt.Fprintln(os.Stderr, err.Error())
				os.Exit(1)
			}
		}
		// Print the branch name and the remote it's tracking
		fmt.Printf("%s%s%s\n", yellow(branch.Name), follow, tracked)
		// Print all the remotes associated with this branch
		printRemotes(branch)
	}
}
