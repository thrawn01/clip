package clip

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type TrackedBranch struct {
	Remote string
	Merge  string
}

type TrackedBranchMap map[string]*TrackedBranch

type Branch struct {
	Name string
	Ref  string
	Sha  string
}

type BranchMap map[string]*Branch
type BranchReferenceMap map[string]BranchMap

type BranchDetail struct {
	Name    string
	Ref     string
	Sha     string
	Remotes []*Branch
	Tracked *TrackedBranch
}

type BranchDetailMap map[string]*BranchDetail

func NewBranch(name, ref, sha string) *Branch {
	return &Branch{
		Name: name,
		Ref:  strings.Trim(ref, " "),
		Sha:  strings.Trim(sha, " "),
	}
}

func NewBranchDetail(branch *Branch) *BranchDetail {
	return &BranchDetail{
		Name: branch.Name,
		Ref:  branch.Ref,
		Sha:  branch.Sha,
	}
}

func ListTrackedBranches(result TrackedBranchMap) error {
	var output string
	// Using git config list all the tracked branch entries
	if err := Run(&output, "git", "config", "--get-regexp", "^branch\\."); err != nil {
		return err
	}
	return ParseTrackedBranches(result, output)
}

// ParseTrackedBranches parses the output of `git config` and return a structure that looks like
//
//	tracked := map[string]*TrackedBranch {
//		"master": &TrackedBranch{ Remote: "origin", Merge: "refs/head/master"},
//		"fix-me-local": &TrackedBranch{ Remote: "upstream", Merge: "refs/head/fix-version"},
//	}
//
func ParseTrackedBranches(result TrackedBranchMap, input string) error {
	regexBranch, _ := regexp.Compile(`branch\.(.*?)\.remote (.+)`)
	regexMerge, _ := regexp.Compile(`branch\.(.*?)\.merge ((refs\/)?heads\/)?(.+)`)

	for _, line := range strings.Split(input, "\n") {
		branch := regexBranch.FindStringSubmatch(line)
		if len(branch) != 0 {
			if tracked, ok := result[branch[1]]; ok {
				tracked.Remote = branch[2]
			} else {
				result[branch[1]] = &TrackedBranch{Remote: branch[2]}
			}
		}
		merge := regexMerge.FindStringSubmatch(line)
		if len(merge) != 0 {
			name := strings.Split(line, " ")[1]
			if tracked, ok := result[merge[1]]; ok {
				tracked.Merge = name
			} else {
				result[merge[1]] = &TrackedBranch{Merge: name}
			}
		}
	}
	return nil
}

func ListBranchRefs(result map[string]BranchMap) error {
	var output string
	// Using git show-ref
	if err := Run(&output, "git", "show-ref"); err != nil {
		return err
	}
	return ParseBranchRefs(result, output)
}

// ParseBranchRefs parses the output of `git show-ref` and return a structure that looks like
//
// 	all := map[string]BranchMap {
//		"local": map[string]*Branch {
//			"master": &Branch{
//			 	Name: "master",
//			 	Sha: "2dc90a39c09e52045a483fc8b58e45da386fb149",
//			 	Ref: "remotes/origin/HEAD",
//			},
//		}
//		"origin": map[string]*Branch {
//			"master": &Branch{
//			 	Name: "master",
//			 	Sha: "2dc90a39c09e52045a483fc8b58e45da386fb149",
//				Ref: "remotes/origin/HEAD",
//			},
//	}
//
func ParseBranchRefs(all map[string]BranchMap, input string) error {
	regexLocal, _ := regexp.Compile(`^heads\/(.+)$`)
	regexRemote, _ := regexp.Compile(`^remotes\/(.+?)\/(.+)$`)

	for _, line := range strings.Split(input, "\n") {
		ref := strings.Split(line, "refs/")
		if len(ref) != 2 {
			continue
		}
		// Is a local Branch
		match := regexLocal.FindStringSubmatch(ref[1])
		if len(match) != 0 {
			if local, ok := all["local"]; ok {
				local[match[1]] = NewBranch(match[1], ref[1], ref[0])
			} else {
				all["local"] = BranchMap{}
				all["local"][match[1]] = NewBranch(match[1], ref[1], ref[0])
			}
		}
		// Is a Remote Branch
		match = regexRemote.FindStringSubmatch(ref[1])
		if len(match) != 0 {
			if remote, ok := all[match[1]]; ok {
				remote[match[2]] = NewBranch(match[2], ref[1], ref[0])
			} else {
				all[match[1]] = BranchMap{}
				all[match[1]][match[2]] = NewBranch(match[2], ref[1], ref[0])
			}
		}
	}
	return nil
}

func FindTrackedBranches(result *BranchDetail, refs BranchReferenceMap, tracked TrackedBranchMap) error {
	// If this branch is listed as a tracked branch
	if trackedBranch, ok := tracked[result.Name]; ok {
		result.Tracked = trackedBranch

		// If the remote tracked branch name differs from the local branch name
		if !strings.HasSuffix(trackedBranch.Merge, result.Name) {
			remote, err := GetRemoteBranchName(trackedBranch.Merge)
			if err != nil {
				return errors.Wrap(err, "FindTrackedBranches()")
			}
			// Add this branch to our list of remotes
			result.Remotes = append(result.Remotes, refs[trackedBranch.Remote][remote])
		}
	}
	return nil
}

func FindRemoteBranches(detail *BranchDetail, refs BranchReferenceMap, tracked TrackedBranchMap) error {
	for remote, branches := range refs {
		// Only interested in remote branches
		if remote == "local" {
			continue
		}
		// If we find a branch with the same name on any of the remotes, assume they are the same
		if branch, ok := branches[detail.Name]; ok {
			detail.Remotes = append(detail.Remotes, branch)
		}
	}
	return nil
}

func MergeBranchDetail(result BranchDetailMap, refs BranchReferenceMap, tracked TrackedBranchMap) error {
	for remote, branches := range refs {
		// Only interested in local branches
		if remote != "local" {
			continue
		}

		for name, branch := range branches {
			detail := NewBranchDetail(branch)

			if err := FindTrackedBranches(detail, refs, tracked); err != nil {
				return err
			}
			if err := FindRemoteBranches(detail, refs, tracked); err != nil {
				return err
			}

			// Since the main branch could be main or master or something else
			if name == "main" || name == "master" || name == "trunk" {
				name = "_trunk_"
			}

			result[name] = detail
		}
	}
	return nil
}

func CommitsBetween(commits *[]string, begin, end string) error {
	if begin == end {
		return nil
	}
	var output string
	if err := Run(&output, "git", "log", "--pretty='%h'", fmt.Sprintf("%s..%s", begin, end)); err != nil {
		return errors.Wrap(err, "CommitsBetween()")
	}
	output = strings.TrimSpace(output)
	if output == "" {
		return nil
	}
	*commits = strings.Split(output, "\n")
	return nil
}

func Run(buf *string, name string, args ...string) error {
	var err error
	var output []byte
	//fmt.Printf("Run: '%s %s'\n", name, args)
	if output, err = exec.Command(name, args...).Output(); err != nil {
		return errors.Wrapf(err, "error running '%s %s'", name, args)
	} else {
		*buf = string(output)
		return nil
	}
}

func ExistsLocally(needle *Branch, refs BranchReferenceMap) bool {
	for name, _ := range refs["local"] {
		if needle.Name == name {
			return true
		}
	}
	return false
}

func ExistsRemotely(needle *Branch, remote string, refs BranchReferenceMap) bool {
	for name, _ := range refs[remote] {
		if needle.Name == name {
			return true
		}
	}
	return false
}

func IsTracked(needle *Branch, remote string, tracked TrackedBranchMap) bool {
	for name, branch := range tracked {
		// If tracked branch shares our name and it's tracking the same remote we are interested in
		if needle.Name == name && branch.Remote == remote {
			return true
		}
	}
	return false
}

func GetRemoteBranchName(merge string) (string, error) {
	regexRemoteName, _ := regexp.Compile(`((refs\/)?heads\/)?(.+)`)
	// Get the remote name
	match := regexRemoteName.FindStringSubmatch(merge)
	if len(match) != 4 {
		return "", errors.New(fmt.Sprintf("Failed to extract tracked branch's"+
			" remote name from '%s'", merge))
	}
	return match[3], nil
}

type Opts struct {
	Default  string
	Validate func(string) bool
}

func readInput(opts Opts, msg string, args ...interface{}) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s > ", fmt.Sprintf(msg, args...))
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return opts.Default
	}
	return input
}

func YesNo(opts Opts, msg string, args ...interface{}) bool {
	isYes, _ := regexp.Compile(`^(Y|y)$`)
	isNo, _ := regexp.Compile(`^(N|n)$`)

	for {
		input := readInput(opts, fmt.Sprintf("%s (Y/N)", msg), args...)
		if isYes.MatchString(input) {
			return true
		}
		if isNo.MatchString(input) {
			return false
		}
	}
}
