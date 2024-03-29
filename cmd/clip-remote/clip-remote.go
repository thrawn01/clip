package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/thrawn01/args"
	"github.com/thrawn01/clip"
)

var yellow = color.New(color.FgYellow).PrintfFunc()

func main() {
	refs := clip.BranchReferenceMap{}
	tracked := clip.TrackedBranchMap{}

	parser := args.NewParser(args.Name("clip-remote"),
		args.Desc("Clips remote branches that no longer are used locally"))
	parser.AddOption("--force").Alias("-f").IsTrue().
		Help("Don't ask before deleting remote branches")
	parser.AddOption("prefix").Default("").Alias("-p").
		Help("Attempt to prune only branches with this prefix." +
			" IE: '-p thrawn' will prune 'thrawn/dev' and 'thrawn/clip' branches")
	parser.AddArgument("remote").Default("origin").
		Help("The name of the remote to clip branches from")

	opts := parser.ParseSimple(nil)

	// Get which remote to clip
	remote := opts.String("remote")

	// List remote and local branches
	if err := clip.ListBranchRefs(refs); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// List tracked local branches
	if err := clip.ListTrackedBranches(tracked); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Find remote branches that do not have local branches and are not tracked
	branches, ok := refs[remote]
	if !ok {
		fmt.Fprintf(os.Stderr, "No such remote named '%s'", remote)
		os.Exit(1)
	}

	for _, branch := range branches {
		if branch.Name == "HEAD" {
			continue
		}

		// Does this branch exist locally?
		if clip.ExistsLocally(branch, refs) {
			continue
		}

		// Is this branch tracking a remote branch?
		if clip.IsTracked(branch, remote, tracked) {
			continue
		}

		if !clip.ExistsRemotely(branch, remote, refs) {
			continue
		}

		// Is this branch name prefixed with something?
		if prefix := opts.String("prefix"); prefix != "" {
			if !strings.HasPrefix(branch.Name, prefix) {
				continue
			}
		}

		if !opts.Bool("force") {
			// Ask if we should delete this remote branch
			msg := "Delete Remote Branch '%s/%s'"
			if !clip.YesNo(clip.Opts{Default: "Y"}, msg, remote, branch.Name) {
				continue
			}
		}

		yellow("Deleting %s/%s..\n", remote, branch.Name)
		// Delete remote branch
		if err := exec.Command("git", "push", remote, "--delete", branch.Name).Run(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
	os.Exit(0)
}
