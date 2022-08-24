//go:build darwin

package game

import "github.com/kralicky/ttr/pkg/api"

const Executable = "Toontown Rewritten"

func ShouldDownload(spec *api.PatchSpec) bool {
	for _, os := range spec.Only {
		if os == "darwin" {
			return true
		}
	}
	return false
}
