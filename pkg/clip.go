package pkg

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
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
		Ref: strings.Trim(ref, " "),
		Sha: strings.Trim(sha, " "),
	}
}

func NewBranchDetail(branch *Branch) *BranchDetail {
	return &BranchDetail{
		Name: branch.Name,
		Ref:  branch.Ref,
		Sha:  branch.Sha,
	}
}

func ParseTrackedBranches(input string, result TrackedBranchMap) error {
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

func ParseBranchRefs(input string, all map[string]BranchMap) error {
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
	regexRemoteName, _ := regexp.Compile(`((refs\/)?heads\/)?(.+)`)

	// If this branch is listed as a tracked branch
	if trackedBranch, ok := tracked[result.Name]; ok {
		result.Tracked = trackedBranch

		// If the remote tracked branch name differs from the local branch name
		if !strings.HasSuffix(trackedBranch.Merge, result.Name) {
			// Get the remote name
			match := regexRemoteName.FindStringSubmatch(trackedBranch.Merge)
			if len(match) != 3 {
				return errors.New(fmt.Sprintf("Failed to extract tracked branch's"+
					" remote name from '%s'", trackedBranch.Merge))
			}
			// Add this branch to our list of remotes
			result.Remotes = append(result.Remotes, refs[trackedBranch.Remote][match[3]])
		}
	}
	return nil
}

func FindRemoteBranches(result *BranchDetail, refs BranchReferenceMap, tracked TrackedBranchMap) error {
	return nil
}

func MergeBranchDetail(refs BranchReferenceMap, tracked TrackedBranchMap, result BranchDetailMap) error {

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
			result[name] = detail
		}
	}
	return nil
}
