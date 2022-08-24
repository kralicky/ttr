package game

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kralicky/ttr/pkg/api"
)

func LaunchProcess(ctx context.Context, creds *api.LoginSuccessPayload) error {
	dir, err := DataDir()
	if err != nil {
		return err
	}
	binary := filepath.Join(dir, Executable)
	// ensure the file is executable
	if err := os.Chmod(binary, 0755); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, binary)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"TTR_GAMESERVER="+creds.Gameserver,
		"TTR_PLAYCOOKIE="+creds.Cookie,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Wait()
}
