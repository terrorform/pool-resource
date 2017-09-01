package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type inRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type inResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata"`
}

type Source struct {
	URI        string `json:"uri"`
	Branch     string `json:"branch"`
	PrivateKey string `json:"private_key"`
	Pool       string `json:"pool"`
}

type Version struct {
	Ref string `json:"ref"`
}

type Metadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func main() {
	var req inRequest
	err := json.NewDecoder(os.Stdin).Decode(&req)
	if err != nil {
		panic(err)
	}

	defer os.Stdin.Close()

	location := os.Args[1]

	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	cloneOptions := &git.CloneOptions{
		URL:      req.Source.URI,
		Progress: os.Stderr,
		Depth:    100,
	}

	if req.Source.PrivateKey != "" {
		// cloneOptions.AuthMethod = ssh.NewPublicKeys()
	}

	repo, err := git.PlainClone(tmpDir, false, cloneOptions)
	if err != nil {
		panic(err)
	}

	work, err := repo.Worktree()
	if err != nil {
		panic(err)
	}

	err = work.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(req.Version.Ref),
	})
	if err != nil {
		panic(err)
	}

	ref, err := repo.Head()
	if err != nil {
		panic(err)
	}

	commit, err := repo.CommitObject(ref.Hash())
	if err != nil {
		panic(err)
	}

	f, err := commit.Files()
	if err != nil {
		panic(err)
	}

	var lockPath string
	f.ForEach(func(f *object.File) error {
		if ".gitkeep" != filepath.Base(f.Name) {
			lockPath = f.Name

			handle, err := os.Create(filepath.Join(location, "metadata"))
			if err != nil {
				panic(err)
			}
			defer handle.Close()

			c, err := f.Contents()
			if err != nil {
				panic(err)
			}

			_, err = handle.Write([]byte(c))
			if err != nil {
				panic(err)
			}

			handle, err = os.Create(filepath.Join(location, "name"))
			if err != nil {
				panic(err)
			}
			defer handle.Close()

			_, err = handle.Write([]byte(filepath.Base(lockPath)))
			if err != nil {
				panic(err)
			}

		}
		return nil
	})

	json.NewEncoder(os.Stdout).Encode(inResponse{
		Version: Version{Ref: ref.Hash().String()},
		Metadata: []Metadata{
			{Name: "lock_name", Value: filepath.Base(lockPath)},
			{Name: "pool_name", Value: req.Source.Pool},
		},
	})
}
