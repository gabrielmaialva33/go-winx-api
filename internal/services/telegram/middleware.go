package telegram

import (
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	"github.com/gotd/td/telegram"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"time"
)

func GetFloodMiddleware(log *zap.Logger) []telegram.Middleware {
	log = log.Named("flood-middleware")

	waiter := floodwait.NewSimpleWaiter().WithMaxRetries(10)
	rateLimiter := ratelimit.New(rate.Every(time.Millisecond*100), 5)

	log.Info("flood middleware initialized")

	return []telegram.Middleware{
		waiter,
		rateLimiter,
	}
}
