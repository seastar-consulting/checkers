package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/seastar-consulting/checkers/types"
	"github.com/stretchr/testify/assert"
)

func setupTestRepo(t *testing.T) (string, *git.Repository) {
	// Create a temporary directory for test repository
	tmpDir, err := os.MkdirTemp("", "git-check-test")
	if err != nil {
		t.Fatal(err)
	}

	// Initialize a test repository
	repo, err := git.PlainInit(tmpDir, false)
	if err != nil {
		t.Fatal(err)
	}

	return tmpDir, repo
}

func createTestCommit(t *testing.T, repo *git.Repository, filename, content string) plumbing.Hash {
	w, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	// Create a test file
	filePath := filepath.Join(w.Filesystem.Root(), filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Add and commit the file
	_, err = w.Add(filename)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := w.Commit("Add "+filename, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Test User",
			Email: "test@example.com",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	return hash
}

func createTestBranch(t *testing.T, repo *git.Repository, name string, hash plumbing.Hash) {
	headRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(name), hash)
	err := repo.Storer.SetReference(headRef)
	if err != nil {
		t.Fatal(err)
	}
}

func setupRemote(t *testing.T, repo *git.Repository, mainCommit plumbing.Hash) {
	// Set up remote reference to simulate origin/main
	remoteRef := plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName("origin", "main"),
		mainCommit,
	)
	err := repo.Storer.SetReference(remoteRef)
	if err != nil {
		t.Fatal(err)
	}

	// Get worktree
	w, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}

	// Configure the remote
	remoteConfig := &config.RemoteConfig{
		Name: "origin",
		URLs: []string{"file://" + w.Filesystem.Root()}, // Use local path as remote
	}
	_, err = repo.CreateRemote(remoteConfig)
	if err != nil {
		t.Fatal(err)
	}

	// Create remote config to avoid fetch errors
	err = repo.CreateBranch(&config.Branch{
		Name:   "main",
		Remote: "origin",
		Merge:  plumbing.NewBranchReferenceName("main"),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCheckRepoUpToDate(t *testing.T) {
	tmpDir, repo := setupTestRepo(t)
	defer os.RemoveAll(tmpDir)

	// Create a stale commit first
	staleCommit := createTestCommit(t, repo, "stale.txt", "stale content")
	createTestBranch(t, repo, "stale", staleCommit)

	// Create main branch with new changes
	mainCommit := createTestCommit(t, repo, "main.txt", "main content")
	createTestBranch(t, repo, "main", mainCommit)

	// Set up remote
	setupRemote(t, repo, mainCommit)

	// Create a feature branch that's up to date with main
	w, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName("main"),
	})
	if err != nil {
		t.Fatal(err)
	}

	featureCommit := createTestCommit(t, repo, "feature.txt", "feature content")
	createTestBranch(t, repo, "feature", featureCommit)

	tests := []struct {
		name           string
		setupFn        func() // Additional setup for the test
		item          types.CheckItem
		expectedStatus types.CheckStatus
		expectedError  bool
		checkOutput   func(t *testing.T, output string)
	}{
		{
			name: "Feature branch contains main branch changes",
			setupFn: func() {
				err := w.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName("feature"),
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			item: types.CheckItem{
				Name: "git.is_up_to_date",
				Type: "git",
				Parameters: map[string]string{
					"path": tmpDir,
				},
			},
			expectedStatus: types.Success,
			expectedError:  false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "contains all changes from default branch")
			},
		},
		{
			name: "Stale branch missing main branch changes (warning)",
			setupFn: func() {
				err := w.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName("stale"),
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			item: types.CheckItem{
				Name: "git.is_up_to_date",
				Type: "git",
				Parameters: map[string]string{
					"path": tmpDir,
				},
			},
			expectedStatus: types.Warning,
			expectedError:  false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "missing changes from default branch")
			},
		},
		{
			name: "Stale branch missing main branch changes (failure)",
			setupFn: func() {
				err := w.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName("stale"),
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			item: types.CheckItem{
				Name: "git.is_up_to_date",
				Type: "git",
				Parameters: map[string]string{
					"path":            tmpDir,
					"fail_out_of_date": "true",
				},
			},
			expectedStatus: types.Failure,
			expectedError:  false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "missing changes from default branch")
			},
		},
		{
			name: "Invalid fail_out_of_date parameter",
			setupFn: func() {
				err := w.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName("stale"),
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			item: types.CheckItem{
				Name: "git.is_up_to_date",
				Type: "git",
				Parameters: map[string]string{
					"path":            tmpDir,
					"fail_out_of_date": "invalid",
				},
			},
			expectedStatus: types.Error,
			expectedError:  false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "Invalid value for 'fail_out_of_date' parameter")
			},
		},
		{
			name: "Invalid repository path",
			item: types.CheckItem{
				Name: "git.is_up_to_date",
				Type: "git",
				Parameters: map[string]string{
					"path": "/nonexistent/path",
				},
			},
			expectedStatus: types.Error,
			expectedError:  false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "Failed to open git repository")
			},
		},
		{
			name: "Feature branch contains main branch changes (explicit default branch)",
			setupFn: func() {
				err := w.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName("feature"),
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			item: types.CheckItem{
				Name: "git.is_up_to_date",
				Type: "git",
				Parameters: map[string]string{
					"path":           tmpDir,
					"default_branch": "main",
				},
			},
			expectedStatus: types.Success,
			expectedError:  false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "contains all changes from default branch")
			},
		},
		{
			name: "Invalid default branch",
			setupFn: func() {
				err := w.Checkout(&git.CheckoutOptions{
					Branch: plumbing.NewBranchReferenceName("feature"),
				})
				if err != nil {
					t.Fatal(err)
				}
			},
			item: types.CheckItem{
				Name: "git.is_up_to_date",
				Type: "git",
				Parameters: map[string]string{
					"path":           tmpDir,
					"default_branch": "nonexistent",
				},
			},
			expectedStatus: types.Error,
			expectedError:  false,
			checkOutput: func(t *testing.T, output string) {
				assert.Contains(t, output, "could not find specified default branch 'nonexistent' in remote")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFn != nil {
				tt.setupFn()
			}

			result, err := CheckRepoUpToDate(tt.item)
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedStatus, result.Status)
			if tt.checkOutput != nil {
				if result.Error != "" {
					tt.checkOutput(t, result.Error)
				} else {
					tt.checkOutput(t, result.Output)
				}
			}
		})
	}
}
