package clip_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thrawn01/clip"
)

var gitConfig string = `branch.master.remote origin
branch.master.merge refs/heads/master
branch.re-fix-version.remote origin
branch.re-fix-version.merge refs/heads/re-fix-version
branch.base-and-flake-fix.remote origin
branch.base-and-flake-fix.merge refs/heads/base-and-flake-fix
branch.fix-me-local.remote upstream
branch.fix-me-local.merge refs/heads/fix-version`

var gitShowRef string = `5f813e2f5a9cd6335e36797dd3428a7632d52102 refs/heads/base-and-flake-fix
1a55f87bb9542848d1b19c2bde3f1552426a6b99 refs/heads/fix-me-local
2dc90a39c09e52045a483fc8b58e45da386fb149 refs/heads/master
228ea1897661759a46541676e6de0cc6bc0bddfc refs/heads/re-fix-version
2dc90a39c09e52045a483fc8b58e45da386fb149 refs/remotes/origin/HEAD
02b58afd28673f8dcc28370a44a6c58877b8950d refs/remotes/origin/base-and-flake-fix
ac0ff092a6bd193fe73660a8f0302e5ed32911dc refs/remotes/origin/fix-version
2dc90a39c09e52045a483fc8b58e45da386fb149 refs/remotes/origin/master
228ea1897661759a46541676e6de0cc6bc0bddfc refs/remotes/origin/re-fix-version
ac0ff092a6bd193fe73660a8f0302e5ed32911dc refs/remotes/upstream/fix-version
2dc90a39c09e52045a483fc8b58e45da386fb149 refs/remotes/upstream/master
01dbc5ce8be93f8437e4ae91833a99e0666b5e5e refs/tags/1.1.0
2dc90a39c09e52045a483fc8b58e45da386fb149 refs/tags/1.3.0
77160475db9c4608ae4acf17fd1eb3e5b2195b2a refs/tags/v1.2.2
`

func TestClip(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Args Parser")
}

var _ = Describe("pkg.clip", func() {
	Describe("ParseTrackedBranches()", func() {
		It("Should parse tracked git branches and merge branches", func() {
			tracked := clip.TrackedBranchMap{}
			err := clip.ParseTrackedBranches(tracked, gitConfig)
			Expect(err).To(BeNil())
			Expect(tracked["master"].Remote).To(Equal("origin"))
			Expect(tracked["re-fix-version"].Merge).To(Equal("refs/heads/re-fix-version"))
			Expect(tracked["base-and-flake-fix"].Merge).To(Equal("refs/heads/base-and-flake-fix"))
		})
	})
	Describe("ParseBranches()", func() {
		var branches map[string]clip.BranchMap
		BeforeEach(func() {
			branches = clip.BranchReferenceMap{}
			err := clip.ParseBranchRefs(branches, gitShowRef)
			Expect(err).To(BeNil())
		})

		It("Should parse local 3 local branches", func() {
			local := branches["local"]
			Expect(len(local)).To(Equal(4))

			// Should be master
			master, ok := local["master"]
			Expect(ok).To(Equal(true))
			Expect(master.Name).To(Equal("master"))
			Expect(master.Ref).To(Equal("heads/master"))
			Expect(master.Sha).To(Equal("2dc90a39c09e52045a483fc8b58e45da386fb149"))

			// Should be re-fix-version
			fix, ok := local["re-fix-version"]
			Expect(ok).To(Equal(true))
			Expect(fix.Name).To(Equal("re-fix-version"))
			Expect(fix.Ref).To(Equal("heads/re-fix-version"))
			Expect(fix.Sha).To(Equal("228ea1897661759a46541676e6de0cc6bc0bddfc"))
		})
		It("Should parse 5 origin branches", func() {
			origin := branches["origin"]
			Expect(len(origin)).To(Equal(5))

			head, ok := origin["HEAD"]
			Expect(ok).To(Equal(true))
			Expect(head.Name).To(Equal("HEAD"))
			Expect(head.Ref).To(Equal("remotes/origin/HEAD"))
			Expect(head.Sha).To(Equal("2dc90a39c09e52045a483fc8b58e45da386fb149"))

			_, ok = origin["master"]
			Expect(ok).To(Equal(true))
			_, ok = origin["fix-version"]
			Expect(ok).To(Equal(true))
			_, ok = origin["base-and-flake-fix"]
			Expect(ok).To(Equal(true))
		})
		It("Should parse 2 upstream branches", func() {
			upstream := branches["upstream"]
			Expect(len(upstream)).To(Equal(2))

			master, ok := upstream["master"]
			Expect(ok).To(Equal(true))
			Expect(master.Name).To(Equal("master"))
			Expect(master.Ref).To(Equal("remotes/upstream/master"))
			Expect(master.Sha).To(Equal("2dc90a39c09e52045a483fc8b58e45da386fb149"))

			fix, ok := upstream["fix-version"]
			Expect(ok).To(Equal(true))
			Expect(fix.Name).To(Equal("fix-version"))
			Expect(fix.Ref).To(Equal("remotes/upstream/fix-version"))
			Expect(fix.Sha).To(Equal("ac0ff092a6bd193fe73660a8f0302e5ed32911dc"))
		})
	})
	Describe("FindTrackedBranches()", func() {
		var detail *clip.BranchDetail
		var tracked clip.TrackedBranchMap
		var refs clip.BranchReferenceMap

		BeforeEach(func() {
			tracked = clip.TrackedBranchMap{}
			refs = clip.BranchReferenceMap{}

			err := clip.ParseTrackedBranches(tracked, gitConfig)
			Expect(err).To(BeNil())
			err = clip.ParseBranchRefs(refs, gitShowRef)
			Expect(err).To(BeNil())
		})
		It("Should fill in branch details if the branch is tracking a remote branch", func() {
			detail = clip.NewBranchDetail(clip.NewBranch("re-fix-version", "", ""))

			err := clip.FindTrackedBranches(detail, refs, tracked)
			Expect(err).To(BeNil())
			Expect(detail.Name).To(Equal("re-fix-version"))
			Expect(detail.Tracked.Remote).To(Equal("origin"))
			Expect(detail.Tracked.Merge).To(Equal("refs/heads/re-fix-version"))
		})

		It("Should fill in branch details even if remote branch has a diff name", func() {
			detail = clip.NewBranchDetail(clip.NewBranch("fix-me-local", "", ""))

			err := clip.FindTrackedBranches(detail, refs, tracked)
			Expect(err).To(BeNil())
			Expect(detail.Name).To(Equal("fix-me-local"))
			Expect(detail.Tracked.Remote).To(Equal("upstream"))
			Expect(detail.Tracked.Merge).To(Equal("refs/heads/fix-version"))
		})
	})
	Describe("FindRemoteBranches()", func() {
		var detail *clip.BranchDetail
		var tracked clip.TrackedBranchMap
		var refs clip.BranchReferenceMap

		BeforeEach(func() {
			tracked = clip.TrackedBranchMap{}
			refs = clip.BranchReferenceMap{}

			err := clip.ParseTrackedBranches(tracked, gitConfig)
			Expect(err).To(BeNil())
			err = clip.ParseBranchRefs(refs, gitShowRef)
			Expect(err).To(BeNil())
		})
		It("Should add remotes to details if the local branch has the same name as a remote branch", func() {
			detail = clip.NewBranchDetail(clip.NewBranch("master", "", ""))

			err := clip.FindRemoteBranches(detail, refs, tracked)
			Expect(err).To(BeNil())
			remotes := detail.Remotes
			Expect(len(remotes)).To(Equal(2))
			Expect(remotes[0].Name).To(Equal("master"))
			Expect(remotes[1].Name).To(Equal("master"))
			Expect(remotes[0].Sha).To(Equal("2dc90a39c09e52045a483fc8b58e45da386fb149"))
			Expect(remotes[1].Sha).To(Equal("2dc90a39c09e52045a483fc8b58e45da386fb149"))
		})
	})
	Describe("MergeBranchDetail()", func() {
		It("Should merge tracked and reference branches into a BranchDetailMap{}", func() {
			details := clip.BranchDetailMap{}
			tracked := clip.TrackedBranchMap{}
			refs := clip.BranchReferenceMap{}

			err := clip.ParseTrackedBranches(tracked, gitConfig)
			Expect(err).To(BeNil())
			err = clip.ParseBranchRefs(refs, gitShowRef)
			Expect(err).To(BeNil())

			// Merge the branch detail
			err = clip.MergeBranchDetail(details, refs, tracked)

			Expect(err).To(BeNil())
			Expect(len(details)).To(Equal(4))

			fix := details["fix-me-local"]
			Expect(fix.Name).To(Equal("fix-me-local"))
			Expect(fix.Tracked.Remote).To(Equal("upstream"))

			master := details["master"]
			Expect(master.Name).To(Equal("master"))
			Expect(master.Tracked.Remote).To(Equal("origin"))
		})
	})
	Describe("CommitsBetween()", func() {
		It("Should return commits between begin and ending sha's", func() {
			commits := make([]string, 0)

			// These sha's exist within our own repository
			err := clip.CommitsBetween(&commits, "152b5832f1e6f06d3efed6e55657c997c41855ed",
				"c73462192ad2e0a690bad82659a7f7c7a1a8bc62")
			Expect(err).To(BeNil())
			Expect(len(commits)).To(Equal(3))
		})
		It("Should return zero commits if begin and ending sha's are the same", func() {
			commits := make([]string, 0)

			// These sha's exist within our own repository
			err := clip.CommitsBetween(&commits, "152b5832f1e6f06d3efed6e55657c997c41855ed",
				"152b5832f1e6f06d3efed6e55657c997c41855ed")
			Expect(err).To(BeNil())
			Expect(len(commits)).To(Equal(0))
		})
	})
	Describe("ExistsLocally()", func() {
		var refs clip.BranchReferenceMap

		BeforeEach(func() {
			refs = clip.BranchReferenceMap{}
			err := clip.ParseBranchRefs(refs, gitShowRef)
			Expect(err).To(BeNil())
		})

		It("Should return true if the branch exists locally", func() {
			result := clip.ExistsLocally(&clip.Branch{Name: "master"}, refs)
			Expect(result).To(Equal(true))
			result = clip.ExistsLocally(&clip.Branch{Name: "fix-me-local"}, refs)
			Expect(result).To(Equal(true))
		})
		It("Should return false if the branch doesn't exist locally", func() {
			result := clip.ExistsLocally(&clip.Branch{Name: "unknown-branch"}, refs)
			Expect(result).To(Equal(false))
		})
	})
	Describe("ExistsRemotely()", func() {
		var refs clip.BranchReferenceMap

		BeforeEach(func() {
			refs = clip.BranchReferenceMap{}
			err := clip.ParseBranchRefs(refs, gitShowRef)
			Expect(err).To(BeNil())
		})

		It("Should return true if the branch exists on the specified remote", func() {
			result := clip.ExistsRemotely(&clip.Branch{Name: "master"}, "origin", refs)
			Expect(result).To(Equal(true))
			result = clip.ExistsRemotely(&clip.Branch{Name: "fix-version"}, "upstream", refs)
			Expect(result).To(Equal(true))
		})
		It("Should return false if the branch doesn't exist on the specified remote", func() {
			result := clip.ExistsRemotely(&clip.Branch{Name: "fix-me-local"}, "origin", refs)
			Expect(result).To(Equal(false))
		})
	})
	Describe("IsTracked()", func() {
		var tracked clip.TrackedBranchMap

		BeforeEach(func() {
			tracked = clip.TrackedBranchMap{}
			err := clip.ParseTrackedBranches(tracked, gitConfig)
			Expect(err).To(BeNil())
		})

		It("Should return true if the branch is tracked remotely", func() {
			result := clip.IsTracked(&clip.Branch{Name: "master"}, "origin", tracked)
			Expect(result).To(Equal(true))
			result = clip.IsTracked(&clip.Branch{Name: "fix-me-local"}, "upstream", tracked)
			Expect(result).To(Equal(true))
		})
		It("Should return false if the branch is tracked remotely but is on wrong remote", func() {
			result := clip.IsTracked(&clip.Branch{Name: "fix-me-local"}, "origin", tracked)
			Expect(result).To(Equal(false))
		})
	})
})
