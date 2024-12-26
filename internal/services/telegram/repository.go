package telegram

import (
	"context"
	"errors"

	"go-winx-api/config"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type Repository struct {
	client  *gotgproto.Client
	channel *tg.InputChannel
	logger  *zap.Logger
}

func NewRepository(client *gotgproto.Client, logger *zap.Logger) *Repository {
	channel, err := GetChannelPeer(context.Background(), client)
	if err != nil {
		logger.Error("failed to get channel peer", zap.Error(err))
		return nil
	}

	return &Repository{
		client:  client,
		channel: channel,
		logger:  logger,
	}
}

func (r *Repository) GetHistory(ctx context.Context) ([]*tg.Message, error) {
	peerClass := r.client.PeerStorage.GetInputPeerById(config.ValueOf.ChannelId)
	history, err := r.client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:  peerClass,
		Limit: 10,
	})
	if err != nil {
		r.logger.Error("failed to get history", zap.Error(err))
		return nil, err
	}

	var messages []*tg.Message

	//switch messages := history.(type) {
	//case *tg.MessagesMessages:
	//	for _, msg := range messages.GetMessages() {
	//		if message, ok := msg.(*tg.Message); ok {
	//			r.logger.Info("Message received", zap.Int("id", message.ID), zap.String("text", message.Message))
	//			r.logger.Info("Message group id", zap.Int64("group_id", message.GroupedID))
	//		}
	//	}
	//case *tg.MessagesMessagesSlice:
	//	for _, msg := range messages.GetMessages() {
	//		if message, ok := msg.(*tg.Message); ok {
	//			r.logger.Info("Message received", zap.Int("id", message.ID), zap.String("text", message.Message))
	//			r.logger.Info("Message group id", zap.Int64("group_id", message.GroupedID))
	//		}
	//	}
	//case *tg.MessagesChannelMessages:
	//	for _, msg := range messages.GetMessages() {
	//		if message, ok := msg.(*tg.Message); ok {
	//			r.logger.Info("Message received", zap.Int("id", message.ID), zap.String("text", message.Message))
	//			r.logger.Info("Message group id", zap.Int64("group_id", message.GroupedID))
	//		}
	//	}
	//default:
	//	r.logger.Warn("Unexpected response type for MessagesGetHistory", zap.Any("type", messages))
	//}

	switch result := history.(type) {
	case *tg.MessagesChannelMessages:
		for _, msg := range result.Messages {
			if m, ok := msg.(*tg.Message); ok {
				messages = append(messages, m)
			}
		}
	}

	return messages, nil

}

func GetChannelPeer(ctx context.Context, client *gotgproto.Client) (*tg.InputChannel, error) {
	peerClass := client.PeerStorage.GetInputPeerById(config.ValueOf.ChannelId)

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
		ChannelID: config.ValueOf.ChannelId,
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
