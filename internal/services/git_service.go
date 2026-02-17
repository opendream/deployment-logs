package services

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gossh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"golang.org/x/crypto/ssh"
)

type GitService struct{}

func buildAuth(sshKey string) (transport.AuthMethod, error) {
	if sshKey != "" {
		keys, err := gossh.NewPublicKeys("git", []byte(sshKey), "")
		if err != nil {
			return nil, fmt.Errorf("ssh key parse failed: %w", err)
		}
		keys.HostKeyCallback = ssh.InsecureIgnoreHostKey()
		return keys, nil
	}
	return nil, nil
}

func (s *GitService) CloneOrPull(repoURL, sshKey, reposDir, repoName, branch string) (*git.Repository, error) {
	dir := filepath.Join(reposDir, repoName)
	auth, err := buildAuth(sshKey)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		repo, err := git.PlainClone(dir, false, &git.CloneOptions{
			URL:           repoURL,
			Auth:          auth,
			ReferenceName: plumbing.NewBranchReferenceName(branch),
			SingleBranch:  true,
		})
		if err != nil {
			return nil, fmt.Errorf("clone failed: %w", err)
		}
		return repo, nil
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		os.RemoveAll(dir)
		return s.CloneOrPull(repoURL, sshKey, reposDir, repoName, branch)
	}

	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("worktree failed: %w", err)
	}

	// Fetch the branch
	err = repo.Fetch(&git.FetchOptions{
		Auth: auth,
		RefSpecs: []config.RefSpec{
			config.RefSpec(fmt.Sprintf("+refs/heads/%s:refs/remotes/origin/%s", branch, branch)),
		},
		Force: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}

	// Checkout the branch
	branchRef := plumbing.NewBranchReferenceName(branch)
	err = w.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Force:  true,
	})
	if err != nil {
		remoteRef := plumbing.NewRemoteReferenceName("origin", branch)
		ref, err2 := repo.Reference(remoteRef, true)
		if err2 != nil {
			return nil, fmt.Errorf("checkout failed: %w", err)
		}
		err = w.Checkout(&git.CheckoutOptions{
			Hash:   ref.Hash(),
			Branch: branchRef,
			Create: true,
			Force:  true,
		})
		if err != nil {
			return nil, fmt.Errorf("checkout failed: %w", err)
		}
	}

	// Pull latest
	err = w.Pull(&git.PullOptions{
		Auth:          auth,
		RemoteName:    "origin",
		ReferenceName: branchRef,
		Force:         true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, fmt.Errorf("pull failed: %w", err)
	}

	return repo, nil
}
