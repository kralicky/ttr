//go:build linux

package game

import "github.com/kralicky/ttr/pkg/api"

const Executable = "TTREngine"

func ShouldDownload(spec *api.PatchSpec) bool {
	for _, os := range spec.Only {
		if os == "linux" || os == "linux2" {
			return true
		}
	}
	return false
}
