package game_test

import (
	"bytes"
	"image/png"
	"os"
	"testing"

	"github.com/kralicky/ttr/pkg/game"
	log "github.com/sirupsen/logrus"
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

	err := game.ShowMintInfo(info)
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

	for status := range statusTracker.C {
		log.Infof("new zone: %s", status.Request)
		go func() {
			if status.Request.Where == "MintInterior" {
				log.Debugf("entered mint, waiting for logs...")
				info, err := game.ScanForMintInfo(status.ZoneLogs)
				if err != nil {
					log.Warnf("error scanning for mint info: %v", err)
					return
				}
				log.Infof("mint info: %s", info)
				game.ShowMintInfo(info)
			}
		}()
	}
}
