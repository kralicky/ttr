package game_test

import (
	"io"
	"testing"

	"github.com/kralicky/ttr/pkg/game"
	"github.com/stretchr/testify/assert"
)

func TestStatusTracker(t *testing.T) {
	zone1 := `:vlt: enter(requestStatus={'loader': 'SafeZoneLoader', 'where': 'Estate', 'how': 'TeleportIn', 'hoodId': 16000, 'zoneId': 12345, 'shardId': None, 'avId': -1, 'ownerId': 12345})`
	zone2 := `:vlt: enter(requestStatus={'loader': 'CogHQLoader', 'where': 'MintInterior', 'how': 'TeleportIn', 'zoneId': 23456, 'mintId': 12700, 'hoodId': 12000})`

	r, w := io.Pipe()
	tracker := game.NewStatusTracker(r)
	c := tracker.C
	go tracker.Run()

	w.Write([]byte(zone1))
	w.Write([]byte("\n"))

	az := <-c
	assert.Equal(t, az.Request.Loader, "SafeZoneLoader")
	assert.Equal(t, az.Request.Where, "Estate")
	assert.Equal(t, az.Request.How, "TeleportIn")
	assert.Equal(t, *az.Request.HoodId, int64(16000))
	assert.Equal(t, *az.Request.ZoneId, int64(12345))
	assert.Nil(t, az.Request.ShardId)
	assert.Equal(t, *az.Request.AvId, int64(-1))

	w.Write([]byte("sample log 1\n"))
	w.Write([]byte("sample log 2\n"))

	l1, ok := <-az.ZoneLogs
	assert.True(t, ok)
	assert.Equal(t, l1, "sample log 1")
	l2, ok := <-az.ZoneLogs
	assert.True(t, ok)
	assert.Equal(t, l2, "sample log 2")

	w.Write([]byte(zone2))
	w.Write([]byte("\n"))

	_, ok = <-az.ZoneLogs
	if ok {
		t.Errorf("expected channel to be closed")
	}

	az = <-c
	assert.Equal(t, az.Request.Loader, "CogHQLoader")
	assert.Equal(t, az.Request.Where, "MintInterior")
	assert.Equal(t, az.Request.How, "TeleportIn")
	assert.Equal(t, *az.Request.ZoneId, int64(23456))
	assert.Equal(t, *az.Request.MintId, int64(12700))
	assert.Equal(t, *az.Request.HoodId, int64(12000))
	assert.Nil(t, az.Request.ShardId)
	assert.Nil(t, az.Request.AvId)

	w.Write([]byte("sample log 3\n"))
	w.Write([]byte("sample log 4\n"))

	l3, ok := <-az.ZoneLogs
	assert.True(t, ok)
	assert.Equal(t, l3, "sample log 3")
	l4, ok := <-az.ZoneLogs
	assert.True(t, ok)
	assert.Equal(t, l4, "sample log 4")

	w.Close()

	_, ok = <-az.ZoneLogs
	if ok {
		t.Errorf("expected channel to be closed")
	}

	_, ok = <-c
	if ok {
		t.Errorf("expected channel to be closed")
	}
}
