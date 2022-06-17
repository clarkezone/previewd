package localrepomanager

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/go-git/go-git/v5/plumbing"

	clarkezoneLog "github.com/clarkezone/previewd/pkg/log"
)

type gitlayer struct {
	repo *git.Repository
	wt   *git.Worktree
	pat  string
}

func clone(repo string, localfolder string) (*gitlayer, error) {
	clarkezoneLog.Debugf("gitlayer::clone repo:%v localfolder:%v", repo, localfolder)
	gl := &gitlayer{}

	var clo = &git.CloneOptions{
		URL:      repo,
		Progress: os.Stdout,
	}

	err := doClone(gl, localfolder, clo)
	if err != nil {
		clarkezoneLog.Errorf("gitlayer::clone doclone failed with %v", err)
		return nil, err
	}

	return gl, nil
}

func doClone(gl *gitlayer, localfolder string, clo *git.CloneOptions) error {
	re, err := git.PlainClone(localfolder, false, clo)
	if err != nil {
		clarkezoneLog.Errorf("Plainclone %v\n", err.Error())
		return err
	}
	gl.repo = re

	wt, err := gl.repo.Worktree()
	if err != nil {
		clarkezoneLog.Errorf("Get worktree %v\n", err.Error())
		return err
	}
	gl.wt = wt

	if err != nil {
		return err
	}
	return nil
}

// func secureClone(repo string, localfolder string, pw string) (*gitlayer, error) {
// 	gl := &gitlayer{}
//
// 	gl.pat = pw
// 	var clo = &git.CloneOptions{
// 		URL:      repo,
// 		Progress: os.Stdout,
// 		Auth: &http.BasicAuth{
// 			Username: "abc123", // yes, this can be anything except an empty string
// 			Password: pw,
// 		},
// 	}
//
// 	err := doClone(gl, localfolder, clo)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return gl, nil
// }

func (gl *gitlayer) checkout(branch string) error {
	remote, err := gl.repo.Remote("origin")
	if err != nil {
		fmt.Printf("Get remote %v\n", err.Error())
		return err
	}

	var feo *git.FetchOptions

	if gl.pat == "" {
		feo = &git.FetchOptions{Force: true}
	} else {
		feo = &git.FetchOptions{
			Auth: &http.BasicAuth{
				Username: "abc123", // yes, this can be anything except an empty string
				Password: gl.pat,
			},
		}
	}

	err = remote.Fetch(feo)
	if err != nil && err.Error() != "already up-to-date" {
		fmt.Printf("Fetch failed %v\n", err.Error())
		return err
	}

	nm := plumbing.NewRemoteReferenceName(remote.Config().Name, branch)

	fmt.Printf("Checking out new branch %v with force", nm)
	err = gl.wt.Checkout(&git.CheckoutOptions{Branch: nm, Force: true})

	if err != nil {
		fmt.Printf("Checkout new branch failed %v\n", err.Error())
		return err
	}
	return nil
}

func (gl *gitlayer) pull(branch string) error {
	fmt.Printf("Pulling branch %v\n", branch)
	nm := plumbing.NewBranchReferenceName(branch)

	var feo *git.PullOptions

	if gl.pat == "" {
		feo = &git.PullOptions{ReferenceName: nm, Force: true}
	} else {
		feo = &git.PullOptions{
			ReferenceName: nm,
			Auth: &http.BasicAuth{
				Username: "abc123", // yes, this can be anything except an empty string
				Password: gl.pat,
			},
		}
	}

	err := gl.wt.Pull(feo)

	if err != nil && err.Error() != "already up-to-date" {
		fmt.Printf("Pull failed %v\n", err.Error())
		return err
	}
	return nil
}
