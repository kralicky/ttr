package game_test

import (
	"bytes"
	"context"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/kralicky/ttr/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestMintMaps(t *testing.T) {
	info := game.MintInfo{StageId: game.CoinMintId, Floor: 5}
	mapPng, err := info.MapImage()
	if err != nil {
		t.Fatalf("error loading map image: %v", err)
	}

	actualBytes, err := game.MintMaps.ReadFile("maps/coin_06.png")
	if err != nil {
		t.Fatalf("error loading map image: %v", err)
	}
	actual, err := png.Decode(bytes.NewReader(actualBytes))
	if err != nil {
		t.Fatalf("error decoding map image: %v", err)
	}
	assert.Equal(t, mapPng.Bounds(), actual.Bounds())
}

func TestMain(m *testing.M) {
	go func() {
		code := m.Run()
		game.ShutdownGLFW()
		os.Exit(code)
	}()
	game.RunGLFW()
}

func TestWindow(t *testing.T) {
	info := game.MintInfo{StageId: game.CoinMintId, Floor: 5}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := game.ShowMintInfo(ctx, info)
	if err != nil {
		t.Fatalf("error showing mint info: %v", err)
	}
}

func TestMintStatusTracker(t *testing.T) {
	logFile, err := os.Open("") // set a log file here
	if err != nil {
		t.Fatalf("error opening log file: %v", err)
	}
	statusTracker := game.NewStatusTracker(logFile)
	go statusTracker.Run()
	go game.RunMintInfoManager(statusTracker)
}
