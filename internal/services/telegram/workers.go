package telegram

import (
	"context"
	"fmt"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/tg"
	"go-winx-api/config"
	"go.uber.org/zap"
	"sync"
	"sync/atomic"
	"time"
)

type Worker struct {
	Id     int
	Client *gotgproto.Client
	Self   *tg.User
	log    *zap.Logger
}

func (w *Worker) String() string {
	return fmt.Sprintf("{Worker (%d|@%s)}", w.Id, w.Self.Username)
}

type UserWorkers struct {
	Users    []*Worker
	starting int
	index    int
	mut      sync.Mutex
	log      *zap.Logger
}

var Workers *UserWorkers = &UserWorkers{
	log:   nil,
	Users: make([]*Worker, 0),
}

func (w *UserWorkers) Init(log *zap.Logger) {
	w.log = log.Named("workers")
}

func (w *UserWorkers) AddDefaultClient(client *gotgproto.Client, self *tg.User) {
	if w.Users == nil {
		w.Users = make([]*Worker, 0)
	}

	w.incStarting()
	w.Users = append(w.Users, &Worker{
		Client: client,
		Id:     w.starting,
		Self:   self,
		log:    w.log,
	})
	w.log.Sugar().Info("default user loaded")
}

func (w *UserWorkers) Add(token string) (err error) {
	w.incStarting()
	var userId int = w.starting
	client, err := startWorker(w.log, token, userId)
	if err != nil {
		return err
	}
	w.log.Sugar().Infof("bot @%s loaded with ID %d", client.Self.Username, userId)
	w.Users = append(w.Users, &Worker{
		Client: client,
		Id:     userId,
		Self:   client.Self,
		log:    w.log,
	})
	return nil
}

func (w *Worker) EnsureValidAccessHash(ctx context.Context, channelID int64) error {
	inputChannel := &tg.InputChannel{
		ChannelID: channelID,
	}

	channels, err := w.Client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
	if err != nil {
		w.log.Error("Failed to refresh access_hash", zap.Error(err))
		return err
	}

	if len(channels.GetChats()) == 0 {
		return fmt.Errorf("no channels found for ID %d", channelID)
	}

	channel, ok := channels.GetChats()[0].(*tg.Channel)
	if !ok {
		return fmt.Errorf("failed to cast channel response to tg.Channel for ID %d", channelID)
	}

	w.Client.PeerStorage.AddPeer(channel.GetID(), channel.AccessHash, storage.TypeChannel, channel.Username)
	w.log.Info("access hash updated successfully", zap.Int64("channel_id", channel.GetID()))

	return nil
}

func GetNextWorker() *Worker {
	Workers.mut.Lock()
	defer Workers.mut.Unlock()
	index := (Workers.index + 1) % len(Workers.Users)
	Workers.index = index
	worker := Workers.Users[index]

	err := worker.EnsureValidAccessHash(context.Background(), config.ValueOf.ChannelId)
	if err != nil {
		Workers.log.Error("Failed to update access_hash for worker", zap.Int("worker_id", worker.Id), zap.Error(err))
	} else {
		Workers.log.Info("Access hash updated successfully", zap.Int("worker_id", worker.Id))
	}

	Workers.log.Sugar().Infof("using worker %d", worker.Id)
	return worker
}

func StartWorkers(log *zap.Logger) (*UserWorkers, error) {
	Workers.Init(log)

	if len(config.ValueOf.StringSessions) == 0 {
		Workers.log.Sugar().Info("no worker bot tokens provided, skipping worker initialization")
		return Workers, nil
	}

	Workers.log.Sugar().Info("starting")

	var wg sync.WaitGroup
	var successfulStarts int32
	totalUsers := len(config.ValueOf.StringSessions)

	for i := 0; i < totalUsers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				err := Workers.Add(config.ValueOf.StringSessions[i])
				done <- err
			}()

			select {
			case err := <-done:
				if err != nil {
					Workers.log.Error("Failed to start worker", zap.Int("index", i), zap.Error(err))
				} else {
					atomic.AddInt32(&successfulStarts, 1)
				}
			case <-ctx.Done():
				Workers.log.Error("Timed out starting worker", zap.Int("index", i))
			}
		}(i)
	}

	wg.Wait() // Wait for all goroutines to finish
	Workers.log.Sugar().Infof("successfully started %d/%d bots", successfulStarts, totalUsers)
	return Workers, nil
}

func (w *UserWorkers) incStarting() {
	w.mut.Lock()
	defer w.mut.Unlock()
	w.starting++
}

func startWorker(l *zap.Logger, botToken string, index int) (*gotgproto.Client, error) {
	log := l.Named("worker").Sugar()
	log.Infof("starting worker with index - %d", index)

	session := sessionMaker.TelethonSession(config.ValueOf.StringSessions[index]).Name(fmt.Sprintf("worker-%d", index))

	client, err := gotgproto.NewClient(
		config.ValueOf.ApiId,
		config.ValueOf.ApiHash,
		gotgproto.ClientTypeBot(botToken),
		&gotgproto.ClientOpts{
			Session:          session,
			DisableCopyright: true,
			Middlewares:      GetFloodMiddleware(log.Desugar()),
		},
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}
