package telegram

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/telegram/downloader"
	"go-winx-api/internal/cache"
	"io"
	"sort"

	"go-winx-api/config"
	"go-winx-api/internal/models"
	"go-winx-api/internal/utils"

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

func (r *Repository) GetClient() *gotgproto.Client {
	return r.client
}

func (r *Repository) GetHistory(ctx context.Context, limit int, offsetID int) ([]*tg.Message, error) {
	peerClass := r.client.PeerStorage.GetInputPeerById(config.ValueOf.ChannelId)
	history, err := r.client.API().MessagesGetHistory(ctx, &tg.MessagesGetHistoryRequest{
		Peer:     peerClass,
		Limit:    limit,
		OffsetID: offsetID,
		MaxID:    0,
		MinID:    0,
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

func (r *Repository) GroupedPosts(ctx context.Context, pagination models.PaginationData) (map[int64][]*tg.Message, error) {
	limit := pagination.PerPage
	offsetID := pagination.OffsetId
	totalGroupsNeeded := limit

	groupedMessages := make(map[int64][]*tg.Message)

	maxLoops := 30

	for len(groupedMessages) < totalGroupsNeeded && maxLoops > 0 {
		currentLimit := max(20, limit*2)

		allMessages, err := r.GetHistory(ctx, currentLimit, offsetID)
		if err != nil {
			return nil, err
		}

		if len(allMessages) == 0 {
			maxLoops--
			offsetID--
			continue
		}

		for _, msg := range allMessages {
			if msg.GroupedID != 0 {
				groupedMessages[msg.GroupedID] = append(groupedMessages[msg.GroupedID], msg)
			}
		}

		minID := allMessages[0].ID
		for _, msg := range allMessages {
			if msg.ID < minID {
				minID = msg.ID
			}
		}
		offsetID = minID - 1
		maxLoops--
	}

	if len(groupedMessages) > totalGroupsNeeded {
		groupedMessages = limitGroupsByMostRecent(groupedMessages, totalGroupsNeeded)
	}

	return groupedMessages, nil
}

func (r *Repository) PaginatePosts(ctx context.Context, pagination models.PaginationData) (*models.PaginatedPosts, error) {
	groupedMessages, err := r.GroupedPosts(ctx, pagination)
	if err != nil {
		return nil, err
	}

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
	} else {
		pagination.FirstOffsetId = 0
		pagination.LastOffsetId = 0
		pagination.Total = 0
	}

	// cache posts for 12 hours
	for _, post := range posts {
		key := fmt.Sprintf("post:%d:%d", post.MessageID, r.client.Self.ID)
		err = cache.GetCache().SetPost(key, &post, 3600*12)
		if err != nil {
			r.logger.Error("failed to cache post", zap.Error(err))
		}
	}

	return &models.PaginatedPosts{
		Data:       posts,
		Pagination: pagination,
	}, nil
}

func (r *Repository) GetPost(ctx context.Context, messageID int) (*models.Post, error) {
	key := fmt.Sprintf("post:%d:%d", messageID, r.client.Self.ID)
	var cachedPost models.Post
	err := cache.GetCache().GetPost(key, &cachedPost)
	if err == nil {
		r.logger.Sugar().Info("using cached post", messageID, r.client.Self.ID)
		return &cachedPost, nil
	}

	peerClass := r.client.PeerStorage.GetInputPeerById(config.ValueOf.ChannelId)
	if peerClass == nil {
		r.logger.Error("channel not configured in PeerStorage")
		return nil, errors.New("channel not configured")
	}

	inputChannel, ok := peerClass.(*tg.InputPeerChannel)
	if !ok {
		r.logger.Error("invalid channel type in PeerStorage")
		return nil, errors.New("invalid channel type")
	}

	req := &tg.ChannelsGetMessagesRequest{
		Channel: &tg.InputChannel{
			ChannelID:  inputChannel.ChannelID,
			AccessHash: inputChannel.AccessHash,
		},
		ID: []tg.InputMessageClass{
			&tg.InputMessageID{ID: messageID},
			&tg.InputMessageID{ID: messageID + 1},
		},
	}

	result, err := r.client.API().ChannelsGetMessages(ctx, req)
	if err != nil {
		r.logger.Error("failed to fetch message from channel", zap.Error(err))
		return nil, err
	}

	var messages []*tg.Message
	switch res := result.(type) {
	case *tg.MessagesChannelMessages:
		for _, msg := range res.Messages {
			if m, ok := msg.(*tg.Message); ok {
				messages = append(messages, m)
			}
		}
	default:
		return nil, errors.New("unexpected response type from Telegram API")
	}

	if len(messages) == 0 {
		return nil, errors.New("message not found")
	}

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].ID < messages[j].ID
	})

	post := createPostFromMessages(messages)
	if post == nil {
		return nil, errors.New("failed to create post from messages")
	}

	err = cache.GetCache().SetPost(key, post, 3600*12) // 12 hours
	if err != nil {
		r.logger.Error("failed to cache post", zap.Error(err))
	}

	return post, nil
}

func (r *Repository) GetPostImage(ctx context.Context, messageID int, output io.Writer) error {
	peerClass := r.client.PeerStorage.GetInputPeerById(config.ValueOf.ChannelId)
	if peerClass == nil {
		r.logger.Error("channel not configured in PeerStorage")
		return errors.New("channel not configured")
	}

	inputChannel, ok := peerClass.(*tg.InputPeerChannel)
	if !ok {
		r.logger.Error("invalid channel type in PeerStorage")
		return errors.New("invalid channel type")
	}

	req := &tg.ChannelsGetMessagesRequest{
		Channel: &tg.InputChannel{
			ChannelID:  inputChannel.ChannelID,
			AccessHash: inputChannel.AccessHash,
		},
		ID: []tg.InputMessageClass{
			&tg.InputMessageID{ID: messageID},
		},
	}

	result, err := r.client.API().ChannelsGetMessages(ctx, req)
	if err != nil {
		r.logger.Error("failed to fetch the message", zap.Error(err))
		return fmt.Errorf("failed to fetch the message: %w", err)
	}

	var photo *tg.Photo
	switch msg := result.(type) {
	case *tg.MessagesChannelMessages:
		for _, message := range msg.Messages {
			if telegramMsg, ok := message.(*tg.Message); ok && telegramMsg.Media != nil {
				if media, ok := telegramMsg.Media.(*tg.MessageMediaPhoto); ok && media.Photo != nil {
					if p, ok := media.Photo.(*tg.Photo); ok {
						photo = p
						break
					}
				}
			}
		}
	}

	if photo == nil {
		r.logger.Error("no photo found in the message")
		return errors.New("no photo found in the message")
	}

	thumbSize := ""
	if len(photo.Sizes) > 0 {
		thumbSize = photo.Sizes[len(photo.Sizes)-1].GetType()
	}

	inputLocation := &tg.InputPhotoFileLocation{
		ID:            photo.ID,
		AccessHash:    photo.AccessHash,
		FileReference: photo.FileReference,
		ThumbSize:     thumbSize,
	}

	dl := downloader.NewDownloader()
	_, err = dl.Download(r.client.API(), inputLocation).Stream(ctx, output)
	if err != nil {
		r.logger.Error("failed to stream the image", zap.Error(err))
		return fmt.Errorf("failed to stream the image: %w", err)
	}

	return nil
}

func (r *Repository) GetPostVideo(ctx context.Context, file *models.File, start, end int64) (io.Reader, error) {
	inputLocation := &tg.InputDocumentFileLocation{
		ID:            file.Location.ID,
		AccessHash:    file.Location.AccessHash,
		FileReference: file.Location.FileReference,
	}

	contentLength := end - start + 1
	reader, err := NewReader(ctx, r.client, inputLocation, start, end, contentLength)
	if err != nil {
		r.logger.Error("failed to create telegram reader", zap.Error(err))
		return nil, err
	}

	return reader, nil
}

func (r *Repository) GetFile(ctx context.Context, messageID int) (*models.File, error) {
	key := fmt.Sprintf("file:%d:%d", messageID, r.client.Self.ID)
	var cachedFile models.File
	err := cache.GetCache().GetFile(key, &cachedFile)
	if err == nil {
		r.logger.Sugar().Info("using cached media message properties", messageID, r.client.Self.ID)
		return &cachedFile, nil
	}

	peerClass := r.client.PeerStorage.GetInputPeerById(config.ValueOf.ChannelId)
	if peerClass == nil {
		r.logger.Error("channel not configured in PeerStorage")
		return nil, errors.New("channel not configured")
	}

	inputChannel, ok := peerClass.(*tg.InputPeerChannel)
	if !ok {
		r.logger.Error("invalid channel type in PeerStorage")
		return nil, errors.New("invalid channel type")
	}

	req := &tg.ChannelsGetMessagesRequest{
		Channel: &tg.InputChannel{
			ChannelID:  inputChannel.ChannelID,
			AccessHash: inputChannel.AccessHash,
		},
		ID: []tg.InputMessageClass{
			&tg.InputMessageID{ID: messageID},
		},
	}

	result, err := r.client.API().ChannelsGetMessages(ctx, req)
	if err != nil {
		r.logger.Error("failed to fetch the message", zap.Error(err))
		return nil, fmt.Errorf("failed to fetch the message: %w", err)

	}

	messages := result.(*tg.MessagesChannelMessages)
	message := messages.Messages[0].(*tg.Message)
	media := message.Media.(*tg.MessageMediaDocument)
	document, _ := media.Document.AsNotEmpty()

	var fileName string
	for _, attribute := range document.Attributes {
		if name, ok := attribute.(*tg.DocumentAttributeFilename); ok {
			fileName = name.FileName
			break
		}
	}

	file := &models.File{
		Location: &tg.InputDocumentFileLocation{ID: document.ID, AccessHash: document.AccessHash, FileReference: document.FileReference},
		FileSize: document.Size,
		FileName: fileName,
		MimeType: document.MimeType,
		ID:       document.ID,
	}

	err = cache.GetCache().SetFile(key, file, 3600*12) // 12 hours
	if err != nil {
		r.logger.Error("failed to cache file", zap.Error(err))
	}

	return file, nil
}

func (r *Repository) GetFileHash(ctx context.Context, messageID int) (string, error) {
	file, err := r.GetFile(ctx, messageID)
	if err != nil {
		return "", err
	}

	hash := &models.HashFileStruct{
		FileName: file.FileName,
		FileSize: file.FileSize,
		MimeType: file.MimeType,
		FileID:   file.ID,
	}

	return hash.Pack(), nil
}

func limitGroupsByMostRecent(groups map[int64][]*tg.Message, needed int) map[int64][]*tg.Message {
	type groupInfo struct {
		groupID  int64
		msgs     []*tg.Message
		maxMsgID int
	}

	var groupsSlice []groupInfo
	for gID, msgs := range groups {
		maxID := msgs[0].ID
		for _, m := range msgs {
			if m.ID > maxID {
				maxID = m.ID
			}
		}
		groupsSlice = append(groupsSlice, groupInfo{
			groupID:  gID,
			msgs:     msgs,
			maxMsgID: maxID,
		})
	}

	sort.Slice(groupsSlice, func(i, j int) bool {
		return groupsSlice[i].maxMsgID > groupsSlice[j].maxMsgID
	})

	if len(groupsSlice) > needed {
		groupsSlice = groupsSlice[:needed]
	}

	limited := make(map[int64][]*tg.Message, needed)
	for _, gi := range groupsSlice {
		limited[gi.groupID] = gi.msgs
	}
	return limited
}

func createPostFromMessages(messages []*tg.Message) *models.Post {
	var info *tg.Message
	var media *tg.Message

	for _, msg := range messages {
		if msg.Message != "" && info == nil {
			info = msg
		}
		if _, isDoc := msg.Media.(*tg.MessageMediaDocument); isDoc && media == nil {
			media = msg
		}
	}

	if info != nil {
		parsedContent := utils.ParseMessageContent(info.Message)

		post := &models.Post{
			ImageURL:        GetImageURL(info.ID),
			MessageID:       info.ID,
			GroupedID:       info.GroupedID,
			Date:            info.Date,
			Author:          info.PostAuthor,
			OriginalContent: info.Message,
			Reactions:       extractReactions(info.Reactions),
			ParsedContent:   parsedContent,
		}

		if media != nil {
			if document, ok := media.Media.(*tg.MessageMediaDocument); ok {
				if document.Document != nil {
					if doc, ok := document.Document.AsNotEmpty(); ok {
						post.DocumentID = doc.ID
						post.DocumentSize = doc.Size
					}
					post.DocumentMessageID = media.ID
					post.VideoURL = GetVideoURL(media.ID)
				}
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

	client.PeerStorage.AddPeer(channel.GetID(), channel.AccessHash, storage.TypeChannel, storage.DefaultUsername)
	return channel.AsInput(), nil
}

func GetImageURL(messageID int) string {
	return fmt.Sprintf(config.ValueOf.Host+"/api/v1/posts/images/%d", messageID)
}

func GetVideoURL(messageID int) string {
	return fmt.Sprintf(config.ValueOf.Host+"/api/v1/posts/videos/%d", messageID)
}
