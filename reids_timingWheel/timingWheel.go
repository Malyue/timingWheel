package reids_timingWheel

import (
	"github.com/go-redis/redis/v8"
	"net/http"
	"sync"
	"time"
)

type RTimeWheel struct {
	sync.Once
	redisClient *redis.Client
	httpClient  *http.Client
	stopc       chan struct{}
	ticker      *time.Ticker
}

type RTaskElement struct {
	Key         string            `json:"key"`
	CallbackURL string            `json:"callback_url"`
	Method      string            `json:"method"`
	Req         interface{}       `json:"req"`
	Header      map[string]string `json:"header"`
}
