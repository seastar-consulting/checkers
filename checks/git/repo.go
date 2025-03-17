package git

import (
	"fmt"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/seastar-consulting/checkers/checks"
	"github.com/seastar-consulting/checkers/types"
)

func init() {
	checks.Register("git.is_up_to_date", "Check if the current branch contains the latest changes from the default remote branch", CheckRepoUpToDate)
}

// findDefaultBranch attempts to find the default branch reference. If defaultBranch is provided,
// it will look for that branch only. Otherwise, it will try main and master in that order.
func findDefaultBranch(repo *git.Repository, defaultBranch string) (*plumbing.Reference, error) {
	var branches []string
	if defaultBranch != "" {
		branches = []string{defaultBranch}
	} else {
		// Try main first, then master
		branches = []string{"main", "master"}
	}

	for _, branch := range branches {
		refName := plumbing.NewRemoteReferenceName("origin", branch)
		ref, err := repo.Reference(refName, true)
		if err == nil {
			return ref, nil
		}
	}

	if defaultBranch != "" {
		return nil, fmt.Errorf("could not find specified default branch '%s' in remote", defaultBranch)
	}
	return nil, fmt.Errorf("could not find default branch (main or master) in remote")
}

// isAncestor checks if the potential ancestor commit is an ancestor of the target commit
func isAncestor(repo *git.Repository, ancestorHash, targetHash plumbing.Hash) (bool, error) {
	// Get commit history
	logOpts := &git.LogOptions{
		From: targetHash,
	}
	commits, err := repo.Log(logOpts)
	if err != nil {
		return false, fmt.Errorf("failed to get commit history: %v", err)
	}
	defer commits.Close()

	// Look for the ancestor commit in the history
	found := false
	err = commits.ForEach(func(c *object.Commit) error {
		if c.Hash == ancestorHash {
			found = true
			return fmt.Errorf("found") // Use error to break early
		}
		return nil
	})

	if err != nil && err.Error() != "found" {
		return false, fmt.Errorf("error traversing history: %v", err)
	}
	return found, nil
}

// CheckRepoUpToDate verifies if the current branch contains the latest changes from the default remote branch
func CheckRepoUpToDate(item types.CheckItem) (types.CheckResult, error) {
	path, ok := item.Parameters["path"]
	if !ok || path == "" {
		path = "." // Default to current directory
	}

	// Check if we should fail when not up to date
	shouldFail := false
	if failStr, ok := item.Parameters["fail_out_of_date"]; ok {
		var err error
		shouldFail, err = strconv.ParseBool(failStr)
		if err != nil {
			return types.CheckResult{
				Name:   item.Name,
				Type:   item.Type,
				Status: types.Error,
				Error:  fmt.Sprintf("Invalid value for 'fail_out_of_date' parameter: %v", err),
			}, nil
		}
	}

	// Get default branch if specified
	defaultBranch := item.Parameters["default_branch"]

	// Open repository
	repo, err := git.PlainOpen(path)
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("Failed to open git repository at '%s': %v", path, err),
		}, nil
	}

	// Get remote
	remote, err := repo.Remote("origin")
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("Failed to get remote 'origin': %v", err),
		}, nil
	}

	// Get current HEAD reference
	head, err := repo.Head()
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("Failed to get HEAD reference: %v", err),
		}, nil
	}

	// Try to fetch latest changes
	err = remote.Fetch(&git.FetchOptions{
		Force: true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		// Check if it's an authentication error
		if err == transport.ErrAuthenticationRequired {
			return types.CheckResult{
				Name:   item.Name,
				Type:   item.Type,
				Status: types.Error,
				Error:  "Authentication required. Please ensure your Git credentials are properly configured.",
			}, nil
		}
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("Failed to fetch from remote: %v", err),
		}, nil
	}

	// Find the default branch reference
	defaultRef, err := findDefaultBranch(repo, defaultBranch)
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  err.Error(),
		}, nil
	}

	// Check if the default branch commit is an ancestor of the current branch
	isAncestorResult, err := isAncestor(repo, defaultRef.Hash(), head.Hash())
	if err != nil {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Error,
			Error:  fmt.Sprintf("Failed to check if branches are up to date: %v", err),
		}, nil
	}

	if isAncestorResult {
		return types.CheckResult{
			Name:   item.Name,
			Type:   item.Type,
			Status: types.Success,
			Output: fmt.Sprintf("Current branch '%s' contains all changes from default branch '%s'",
				head.Name().Short(), defaultRef.Name().Short()),
		}, nil
	}

	status := types.Warning
	if shouldFail {
		status = types.Failure
	}

	return types.CheckResult{
		Name:   item.Name,
		Type:   item.Type,
		Status: status,
		Output: fmt.Sprintf("Current branch '%s' is missing changes from default branch '%s'. Please merge or rebase.",
			head.Name().Short(), defaultRef.Name().Short()),
	}, nil
}
