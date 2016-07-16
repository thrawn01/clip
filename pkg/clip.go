package pkg

import (
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

func NewBranch(name, ref, sha string) *Branch {
	return &Branch{
		Name: name,
		Ref:  strings.Trim(ref, " "),
		Sha:  strings.Trim(sha, " "),
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

func ParseBranches(input string, all map[string]BranchMap) error {
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
