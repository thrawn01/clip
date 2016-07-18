package main

import (
	"fmt"
	"os"

	"os/exec"

	"github.com/thrawn01/args"
	"github.com/thrawn01/clip/pkg"
)

func main() {
	refs := pkg.BranchReferenceMap{}
	tracked := pkg.TrackedBranchMap{}

	parser := args.NewParser(args.Name("clip-remote"),
		args.Desc("Clips remote branches that no longer are used locally"))
	parser.AddOption("--force").Alias("-f").IsTrue().
		Help("Don't ask before deleting remote branches")
	parser.AddPositional("remote").Default("origin").
		Help("The name of the remote to clip branches from")

	opts := parser.ParseArgsSimple(nil)

	// Get which remote to clip
	remote := opts.String("remote")

	// List remote and local branches
	if err := pkg.ListBranchRefs(refs); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// List tracked local branches
	if err := pkg.ListTrackedBranches(tracked); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	// Find remote branches that do not have local branches and are not tracked
	branches, ok := refs[remote]
	if !ok {
		fmt.Fprintln(os.Stderr, "No such remote named '%s'", remote)
		os.Exit(1)
	}

	for _, branch := range branches {
		if branch.Name == "HEAD" {
			continue
		}
		// Does this branch exist locally?
		if pkg.ExistsLocally(branch, refs) {
			continue
		}

		// Is this branch tracking a remote branch?
		if pkg.IsTracked(branch, remote, tracked) {
			continue
		}

		if !pkg.ExistsRemotely(branch, remote, refs) {
			continue
		}

		if !opts.Bool("force") {
			// Ask if we should delete this remote branch
			msg := "Delete Remote Branch '%s/%s'"
			if !pkg.YesNo(pkg.Opts{Default: "Y"}, msg, remote, branch.Name) {
				continue
			}
		}

		fmt.Printf("\033[33mDeleteing %s/%s..\033[0m\n", remote, branch.Name)
		// Delete remote branch
		if err := exec.Command("git", "push", remote, "--delete", branch.Name).Run(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
	os.Exit(0)
}
