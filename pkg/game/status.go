package game

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

type StatusTracker struct {
	logReader io.Reader
	C         chan *ActiveZone
}

type ActiveZone struct {
	Request EnterRequestStatus

	ZoneLogs chan string // buffered channel of lines
}

func NewStatusTracker(logStream io.Reader) *StatusTracker {
	return &StatusTracker{
		logReader: logStream,
		C:         make(chan *ActiveZone, 1),
	}
}

const enterPrefix = "enter(requestStatus="

const (
	CoinMintId    = 12500
	DollarMintId  = 12600
	BullionMintId = 12700
)

type EnterRequestStatus struct {
	Loader  string `json:"loader"`
	Where   string `json:"where"`
	How     string `json:"how"`
	ZoneId  *int64 `json:"zoneId,omitempty"`
	HoodId  *int64 `json:"hoodId,omitempty"`
	ShardId *int64 `json:"shardId,omitempty"`
	AvId    *int64 `json:"avId,omitempty"`
	MintId  *int64 `json:"mintId,omitempty"`
}

func (e EnterRequestStatus) String() string {
	bytes, _ := json.Marshal(e)
	return string(bytes)
}

func toRequestStatusJson(str string) string {
	return strings.ReplaceAll(strings.ReplaceAll(str, `'`, `"`), `: None`, `: null`)
}

func (r *StatusTracker) Run() {
	defer close(r.C)
	scan := bufio.NewScanner(r.logReader)
	var curZoneLogs chan string
	for scan.Scan() {
		line := scan.Text()
		idx := strings.Index(line, enterPrefix)
		if idx != -1 {
			// new zone
			if curZoneLogs != nil {
				close(curZoneLogs)
			}
			statusJson := toRequestStatusJson(line[idx+len(enterPrefix) : len(line)-1])
			var status EnterRequestStatus
			if err := json.Unmarshal([]byte(statusJson), &status); err != nil {
				continue
			}
			curZoneLogs = make(chan string, 2048)
			az := &ActiveZone{
				Request:  status,
				ZoneLogs: curZoneLogs,
			}
			r.C <- az
		} else {
			if curZoneLogs != nil {
				select {
				case curZoneLogs <- line:
				default:
					// buffer is full - it's not being read, so ignore
				}
			}
		}
	}
	if curZoneLogs != nil {
		close(curZoneLogs)
	}
}
