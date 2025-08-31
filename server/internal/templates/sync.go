package templates

import (
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/middlewarr/server/internal/tools"
)

// SyncTemplates ensures the local templates repo matches the latest commit
// of the configured remote URL + branch. If missing or corrupted, it reclones.
func SyncTemplates() error {
	l := tools.GetLogger()
	s := tools.GetSettings()

	repositoryURL := s.String("templates.repository")
	repositoryBranch := s.String("templates.branch")

	repoPath := tools.GetTemplatesPath()

	// Try to open repo
	r, err := git.PlainOpen(repoPath)
	if err == git.ErrRepositoryNotExists {
		// Clone fresh if missing
		l.Warn().Msg("Templates repo not found, cloning fresh...")

		_, err = git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:               repositoryURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			SingleBranch:      true,
			ReferenceName:     plumbing.NewBranchReferenceName(repositoryBranch),
		})

		return err
	} else if err != nil {
		// Repo corrupted, remove + reclone
		l.Warn().Msg("Templates repo corrupted, recloning...")

		_ = os.RemoveAll(repoPath)
		_, err = git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:               repositoryURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			SingleBranch:      true,
			ReferenceName:     plumbing.NewBranchReferenceName(repositoryBranch),
		})

		return err
	}

	// If repo exists, fetch updates
	err = r.Fetch(&git.FetchOptions{RemoteName: "origin"})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		l.Warn().Msgf("Fetch failed, recloning repo: %v", err)

		_ = os.RemoveAll(repoPath)
		_, err = git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:               repositoryURL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			SingleBranch:      true,
			ReferenceName:     plumbing.NewBranchReferenceName(repositoryBranch),
		})

		return err
	}

	// Compare local vs remote
	ref, err := r.Head()
	if err != nil {
		return err
	}

	localHash := ref.Hash()

	remoteRef := plumbing.NewRemoteReferenceName("origin", repositoryBranch)
	remoteHash, err := r.ResolveRevision(plumbing.Revision(remoteRef.String()))
	if err != nil {
		return err
	}

	// If not up-to-date, hard reset to remote
	if localHash != *remoteHash {
		l.Info().Msg("Repo behind remote, resetting and pulling...")

		w, err := r.Worktree()
		if err != nil {
			return err
		}

		// Hard reset to remote
		err = w.Reset(&git.ResetOptions{
			Mode:   git.HardReset,
			Commit: *remoteHash,
		})
		if err != nil {
			return err
		}
	}

	l.Info().Msg("Templates repo is up-to-date")
	return nil
}
