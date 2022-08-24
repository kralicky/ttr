//go:build windows

package game

import "github.com/kralicky/ttr/pkg/api"

const Executable = "TTREngine.exe"

func ShouldDownload(spec *api.PatchSpec) bool {
	for _, os := range spec.Only {
		if os == "win32" || os == "win64" {
			return true
		}
	}
	return false
}
