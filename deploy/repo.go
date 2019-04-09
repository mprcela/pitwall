package deploy

import (
	"fmt"
	"os"

	"github.com/minus5/svckit/log"

	"code.gitea.io/git"
)

// NewRepo clones repository, pulls changes
func NewRepo(root, from string) (Repo, error) {
	r := Repo{
		root: root,
		from: from,
	}
	if err := r.Clone(); err != nil {
		return r, err
	}
	if err := r.Pull(); err != nil {
		return r, err
	}
	return r, nil
}

// Repo repository structure
type Repo struct {
	root string
	from string
}

// Pull repository
func (r Repo) Pull() error {
	log.S("from", r.from).S("to", r.root).Info("pull")
	return git.Pull(r.root, git.PullRemoteOptions{
		All:    true,
		Rebase: true,
	})
}

// Push to repository
func (r Repo) Push() error {
	log.S("repo", r.root).Info("git push")
	return git.Push(r.root, git.PushOptions{Remote: "origin", Branch: "master"})

}

// Commit to repository
func (r Repo) Commit(msg string, files ...string) error {
	log.S("repo", r.root).S("files", fmt.Sprintf("%v", files)).Info("git commit")
	if err := git.AddChanges(r.root, false, files...); err != nil {
		return err
	}
	err := git.CommitChanges(r.root, git.CommitChangesOptions{
		Message: msg,
	})
	if err != nil {
		return err
	}
	return r.Push()
}

// Clone repository
func (r Repo) Clone() error {
	if fi, err := os.Stat(r.root); err == nil && fi.IsDir() {
		return nil
	}
	log.S("repo", r.root).Info("git clone")
	return git.Clone(r.from, r.root, git.CloneRepoOptions{})
}
