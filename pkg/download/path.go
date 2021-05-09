package download

import (
	"os"
	"path"
	"strings"
)

// PathClean replace ~/ by user's home directory and
// call path.Clean to secure the path
func PathClean(p string) string {
	p = path.Clean(p)
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		p = path.Join(home, p[2:])
	}
	return p
}
