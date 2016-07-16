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
