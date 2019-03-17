package main

func main() {
	state := NewState()

	repo, err := state.GetRepository()
	if err != nil {
		state.Panic(err, "Error fetching repo!")
	}

	defaultBranch := repo.GetDefaultBranch()
	state.Log("default branch: %s", defaultBranch)

	branch, err := state.GetBranch(defaultBranch)
	if err != nil {
		state.Panic(err, "Error fetching branch %s!", defaultBranch)
	}

	commit := branch.GetCommit()
	if commit == nil {
		state.Fatal("Error fetching %s head commit!", defaultBranch)
	}

	sha := commit.GetSHA()
	if sha == "" {
		state.Fatal("Error fetching %s head commit!", defaultBranch)
	}
	state.Log("%s HEAD commit %s", defaultBranch, sha)
}
