package pkg_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/thrawn01/clip/pkg"
)

var input string = `branch.master.remote origin
branch.master.merge refs/heads/master
branch.re-fix-version.remote origin
branch.re-fix-version.merge refs/heads/re-fix-version
branch.base-and-flake-fix.remote origin
branch.base-and-flake-fix.merge refs/heads/base-and-flake-fix`

func TestClip(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Args Parser")
}

var _ = Describe("pkg.clip", func() {

	Describe("ParseTrackedBranches()", func() {
		It("Should parse tracked git branches and merge branches", func() {
			tracked := pkg.TrackedBranchMap{}
			err := pkg.ParseTrackedBranches(input, tracked)
			Expect(err).To(BeNil())
			Expect(tracked["master"].Remote).To(Equal("origin"))
			Expect(tracked["re-fix-version"].Merge).To(Equal("refs/heads/re-fix-version"))
		})
	})
})
