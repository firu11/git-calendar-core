//go:build !js

package filesystem

import (
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
)

func GetRepoFS() (billy.Filesystem, error) {
	return osfs.New("./my-git-repo"), nil
}
