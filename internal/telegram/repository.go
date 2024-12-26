package telegram

import (
	"context"
	"errors"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/tg"
	"go-winx-api/config"
	"go.uber.org/zap"
)

type Repository struct {
	Client  *gotgproto.Client
	Channel *tg.InputChannel
	Logger  *zap.Logger
}

func NewRepository(client *gotgproto.Client, logger *zap.Logger) *Repository {
	channel, err := GetChannelPeer(context.Background(), client)
	if err != nil {
		logger.Error("failed to get channel peer", zap.Error(err))
		return nil
	}

	return &Repository{
		Client:  client,
		Channel: channel,
		Logger:  logger,
	}
}

func (r *Repository) GetHistory(ctx context.Context) {
	peerClass := r.Client.PeerStorage.GetInputPeerById(config.ValueOf.ChannelID)
	history, err := r.Client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:  peerClass,
		Limit: 10,
	})
	if err != nil {
		r.Logger.Error("failed to get history", zap.Error(err))
		return
	}

	switch messages := history.(type) {
	case *tg.MessagesMessages:
		for _, msg := range messages.GetMessages() {
			if message, ok := msg.(*tg.Message); ok {
				r.Logger.Info("Message received", zap.Int("id", message.ID), zap.String("text", message.Message))
			}
		}
	case *tg.MessagesMessagesSlice:
		for _, msg := range messages.GetMessages() {
			if message, ok := msg.(*tg.Message); ok {
				r.Logger.Info("Message received", zap.Int("id", message.ID), zap.String("text", message.Message))
			}
		}
	case *tg.MessagesChannelMessages:
		for _, msg := range messages.GetMessages() {
			if message, ok := msg.(*tg.Message); ok {
				r.Logger.Info("Message received", zap.Int("id", message.ID), zap.String("text", message.Message))
			}
		}
	default:
		r.Logger.Warn("Unexpected response type for MessagesGetHistory", zap.Any("type", messages))
	}

}

func GetChannelPeer(ctx context.Context, client *gotgproto.Client) (*tg.InputChannel, error) {
	peerClass := client.PeerStorage.GetInputPeerById(config.ValueOf.ChannelID)

	switch peer := peerClass.(type) {
	case *tg.InputPeerEmpty:
		break
	case *tg.InputPeerChannel:
		return &tg.InputChannel{
			ChannelID:  peer.ChannelID,
			AccessHash: peer.AccessHash,
		}, nil
	default:
		return nil, errors.New("unexpected type of input peer")
	}

	inputChannel := &tg.InputChannel{
		ChannelID: config.ValueOf.ChannelID,
	}
	channels, err := client.API().ChannelsGetChannels(ctx, []tg.InputChannelClass{inputChannel})
	if err != nil {
		return nil, err
	}
	if len(channels.GetChats()) == 0 {
		return nil, errors.New("no channels found")
	}
	channel, ok := channels.GetChats()[0].(*tg.Channel)
	if !ok {
		return nil, errors.New("type assertion to *tg.Channel failed")
	}

	client.PeerStorage.AddPeer(channel.GetID(), channel.AccessHash, storage.TypeChannel, "")

	return channel.AsInput(), nil
}
