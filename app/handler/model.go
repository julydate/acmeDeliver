package handler

import (
	"github.com/zekroTJA/timedmap"
)

type Handler struct {
	key      string
	timeDiff int64

	cache *timedmap.TimedMap
}
