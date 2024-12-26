package telegram

import (
	"context"
	"errors"
	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/storage"
	"github.com/gotd/td/tg"
	"go-winx-api/config"
	"go-winx-api/internal/models"
	"go.uber.org/zap"
	"sort"
)

type Repository struct {
	client  *gotgproto.Client
	channel *tg.InputChannel
	logger  *zap.Logger
}

func NewRepository(client *gotgproto.Client, logger *zap.Logger) *Repository {
	channel, err := GetInputChannel(context.Background(), client)
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

func (r *Repository) getHistory(ctx context.Context, limit int, offsetID int) ([]*tg.Message, error) {
	peerClass := r.client.PeerStorage.GetInputPeerById(config.ValueOf.ChannelId)
	history, err := r.client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:     peerClass,
		Limit:    limit,
		OffsetID: offsetID,
	})
	if err != nil {
		r.logger.Error("failed to get history", zap.Error(err))
		return nil, err
	}

	var messages []*tg.Message

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

func groupMessagesByGroupedID(messages []*tg.Message) map[int64][]*tg.Message {
	groupedMessages := make(map[int64][]*tg.Message)

	for _, msg := range messages {
		if msg.GroupedID != 0 {
			groupID := msg.GroupedID
			groupedMessages[groupID] = append(groupedMessages[groupID], msg)
		}
	}

	return groupedMessages
}

func createPostFromMessages(messages []*tg.Message) *models.Post {
	var info *tg.Message
	var media *tg.Message

	for _, msg := range messages {
		if msg.Message != "" && info == nil {
			info = msg
		}
		if msg.Media != nil && media == nil {
			media = msg
		}
	}

	if info != nil {
		post := &models.Post{
			MessageID:       info.ID,
			GroupedID:       info.GroupedID,
			Date:            info.Date,
			Author:          info.PostAuthor,
			OriginalContent: info.Message,
			Reactions:       extractReactions(info.Reactions),
			ParsedContent:   nil,
		}
		if media != nil {
			if photo, ok := media.Media.(*tg.MessageMediaPhoto); ok {
				post.ImageURL = extractPhotoURL(photo)
			}
		}
		return post
	}

	return nil
}

func extractReactions(reactions tg.MessageReactions) []models.Reaction {
	var extractedReactions []models.Reaction
	if len(reactions.Results) == 0 {
		return extractedReactions
	}

	for _, reaction := range reactions.Results {
		if emoji, ok := reaction.Reaction.(*tg.ReactionEmoji); ok {
			extractedReactions = append(extractedReactions, models.Reaction{
				Reaction: emoji.Emoticon,
				Count:    reaction.Count,
			})
		}
	}

	return extractedReactions
}

func extractPhotoURL(photo *tg.MessageMediaPhoto) string {
	return ""
}

func (r *Repository) PaginatePosts(ctx context.Context, pagination models.PaginationData) (*models.PaginatedPosts, error) {
	limit := pagination.PerPage
	offsetID := pagination.OffsetId

	messages, err := r.getHistory(ctx, limit, offsetID)
	if err != nil {
		return nil, err
	}

	groupedMessages := groupMessagesByGroupedID(messages)

	var posts []models.Post
	for _, group := range groupedMessages {
		post := createPostFromMessages(group)
		if post != nil {
			posts = append(posts, *post)
		}
	}

	sort.Slice(posts, func(i, j int) bool {
		return posts[i].MessageID > posts[j].MessageID
	})

	total := len(posts)
	if total > 0 {
		pagination.FirstOffsetId = posts[0].MessageID
		pagination.LastOffsetId = posts[total-1].MessageID
		pagination.Total = total
	}

	return &models.PaginatedPosts{
		Data:       posts,
		Pagination: pagination,
	}, nil
}

func GetInputChannel(ctx context.Context, client *gotgproto.Client) (*tg.InputChannel, error) {
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
