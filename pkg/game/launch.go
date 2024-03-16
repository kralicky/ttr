package game

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kralicky/ttr/pkg/api"
	log "github.com/sirupsen/logrus"
)

func LaunchProcess(ctx context.Context, creds *api.LoginSuccessPayload) error {
	dir, err := DataDir()
	if err != nil {
		return err
	}
	logsDir := filepath.Join(dir, "logs")
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return err
	}

	timestamp := time.Now().Unix()
	logFile := filepath.Join(logsDir, fmt.Sprintf("ttr-%d.log", time.Now().Unix()))
	// if the name exists, add 1 to the timestamp
	for {
		_, err := os.Stat(logFile)
		if os.IsNotExist(err) {
			break
		}
		timestamp++
		logFile = filepath.Join(logsDir, fmt.Sprintf("ttr-%d.log", timestamp))
	}
	// open the log file for writing
	f, err := os.Create(logFile)
	if err != nil {
		return err
	}
	defer f.Close()
	log.Infof("writing logs to %s", logFile)

	binary := filepath.Join(dir, Executable)
	// ensure the file is executable
	if err := os.Chmod(binary, 0o755); err != nil {
		return err
	}
	ctx, ca := context.WithCancel(ctx)
	// cancel on sigint
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT)
	go func() {
		// if the user hits ctrl-c twice, cancel the context
		count := 0
		for {
			select {
			case <-sigint:
				if count == 0 {
					log.Warn("\nReceived Ctrl+C once; press again to exit")
					count++
				} else {
					log.Warn("\nReceived Ctrl+C twice; exiting...")
					ca()
				}
			case <-ctx.Done():
			}
		}
	}()
	cmd := exec.CommandContext(ctx, binary)
	cmd.Dir = dir
	cmd.Env = append(cmd.Environ(),
		"TTR_GAMESERVER="+creds.Gameserver,
		"TTR_PLAYCOOKIE="+creds.Cookie,
	)
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	return cmd.Run()
}
