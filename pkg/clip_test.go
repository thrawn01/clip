package pkg_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thrawn01/clip/pkg"
)

var gitConfig string = `branch.master.remote origin
branch.master.merge refs/heads/master
branch.re-fix-version.remote origin
branch.re-fix-version.merge refs/heads/re-fix-version
branch.base-and-flake-fix.remote origin
branch.base-and-flake-fix.merge refs/heads/base-and-flake-fix
`

var gitShowRef string = `5f813e2f5a9cd6335e36797dd3428a7632d52102 refs/heads/base-and-flake-fix
2dc90a39c09e52045a483fc8b58e45da386fb149 refs/heads/master
228ea1897661759a46541676e6de0cc6bc0bddfc refs/heads/re-fix-version
2dc90a39c09e52045a483fc8b58e45da386fb149 refs/remotes/origin/HEAD
5f813e2f5a9cd6335e36797dd3428a7632d52102 refs/remotes/origin/base-and-flake-fix
ac0ff092a6bd193fe73660a8f0302e5ed32911dc refs/remotes/origin/fix-version
2dc90a39c09e52045a483fc8b58e45da386fb149 refs/remotes/origin/master
228ea1897661759a46541676e6de0cc6bc0bddfc refs/remotes/origin/re-fix-version
ac0ff092a6bd193fe73660a8f0302e5ed32911dc refs/remotes/upstream/fix-version
77160475db9c4608ae4acf17fd1eb3e5b2195b2a refs/tags/v1.2.2
2dc90a39c09e52045a483fc8b58e45da386fb149 refs/remotes/upstream/master
`

func TestClip(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Args Parser")
}

var _ = Describe("pkg.clip", func() {
	Describe("ParseTrackedBranches()", func() {
		It("Should parse tracked git branches and merge branches", func() {
			tracked := pkg.TrackedBranchMap{}
			err := pkg.ParseTrackedBranches(gitConfig, tracked)
			Expect(err).To(BeNil())
			Expect(tracked["master"].Remote).To(Equal("origin"))
			Expect(tracked["re-fix-version"].Merge).To(Equal("refs/heads/re-fix-version"))
			Expect(tracked["base-and-flake-fix"].Merge).To(Equal("refs/heads/base-and-flake-fix"))
		})
	})
	Describe("ParseBranches()", func() {
		var branches map[string]pkg.BranchMap
		BeforeEach(func() {
			branches = make(map[string]pkg.BranchMap)
			err := pkg.ParseBranches(gitShowRef, branches)
			Expect(err).To(BeNil())
		})

		It("Should parse local 3 local branches", func() {
			local := branches["local"]
			Expect(len(local)).To(Equal(3))

			// Should be a master
			master := local["master"]
			Expect(master.Name).To(Equal("master"))
			Expect(master.Ref).To(Equal("heads/master"))
			Expect(master.Sha).To(Equal("2dc90a39c09e52045a483fc8b58e45da386fb149"))

			// Should be a re-fix-version
			fix := local["re-fix-version"]
			Expect(fix.Name).To(Equal("re-fix-version"))
			Expect(fix.Ref).To(Equal("heads/re-fix-version"))
			Expect(fix.Sha).To(Equal("228ea1897661759a46541676e6de0cc6bc0bddfc"))
		})
		It("Should parse 5 origin branches", func() {
			origin := branches["origin"]
			Expect(len(origin)).To(Equal(5))
		})
		It("Should parse 2 upstream branches", func() {
			upstream := branches["origin"]
			Expect(len(upstream)).To(Equal(5))
		})
	})
})
