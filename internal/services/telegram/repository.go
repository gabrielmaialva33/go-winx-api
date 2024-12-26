package telegram

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/telegram/downloader"
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

type RangeWriter struct {
	Output io.Writer
	Start  int
	End    int
	pos    int
}

func (rw *RangeWriter) Write(p []byte) (int, error) {
	if rw.pos+len(p) < rw.Start {
		rw.pos += len(p)
		return len(p), nil
	}

	start := 0
	if rw.pos < rw.Start {
		start = rw.Start - rw.pos
	}

	end := len(p)
	if rw.pos+len(p) > rw.End {
		end = rw.End - rw.pos
	}

	n, err := rw.Output.Write(p[start:end])
	rw.pos += len(p)
	return n, err
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

	return &models.PaginatedPosts{
		Data:       posts,
		Pagination: pagination,
	}, nil
}

func (r *Repository) GetPost(ctx context.Context, messageID int) (*models.Post, error) {
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

func (r *Repository) StreamVideo(ctx context.Context, messageID int, output io.Writer, start, end int) error {
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

	// Request the message by ID
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

	// Check if the message contains a video
	var document *tg.Document
	switch res := result.(type) {
	case *tg.MessagesChannelMessages:
		for _, msg := range res.Messages {
			if message, ok := msg.(*tg.Message); ok {
				if media, ok := message.Media.(*tg.MessageMediaDocument); ok {
					document, _ = media.Document.AsNotEmpty()
					break
				}
			}
		}
	default:
		return errors.New("unexpected response type from Telegram API")
	}

	if document == nil {
		r.logger.Error("video not found in the message", zap.Int("messageID", messageID))
		return errors.New("video not found in the message")
	}

	documentLocation := &tg.InputDocumentFileLocation{
		ID:            document.ID,
		AccessHash:    document.AccessHash,
		FileReference: document.FileReference,
	}

	dl := downloader.NewDownloader()
	dl = dl.WithPartSize(1024 * 1024) // Set chunk size to 1MB

	builder, err := dl.Download(r.client.API(), documentLocation).Stream(ctx, &RangeWriter{Output: output, Start: start, End: end})
	fmt.Println("Start", start, "End", end)
	if err != nil {
		r.logger.Error("Failed to download the video", zap.Error(err))
		return fmt.Errorf("failed to download the video: %w", err)
	}

	fmt.Println("Builder", builder)

	return nil
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
					post.VideoURL = GetVideoURL(media.ID, post.DocumentSize)
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

	client.PeerStorage.AddPeer(channel.GetID(), channel.AccessHash, storage.TypeChannel, "")
	return channel.AsInput(), nil
}

func GetImageURL(messageID int) string {
	return fmt.Sprintf(config.ValueOf.Host+"/api/v1/posts/images/%d", messageID)
}

func GetVideoURL(messageID int, size int64) string {
	return fmt.Sprintf(config.ValueOf.Host+"/api/v1/posts/videos/%d/%d", messageID, size)
}

type telegramReader struct {
	ctx           context.Context
	log           *zap.Logger
	client        *gotgproto.Client
	location      *tg.InputDocumentFileLocation
	start         int64
	end           int64
	next          func() ([]byte, error)
	buffer        []byte
	bytesread     int64
	chunkSize     int64
	i             int64
	contentLength int64
}

func (*telegramReader) Close() error {
	return nil
}

func NewTelegramReader(
	ctx context.Context,
	client *gotgproto.Client,
	location *tg.InputDocumentFileLocation,
	start int64,
	end int64,
	contentLength int64,
) (io.ReadCloser, error) {

	r := &telegramReader{
		ctx:           ctx,
		log:           utils.Logger.Named("telegram_reader"),
		location:      location,
		client:        client,
		start:         start,
		end:           end,
		chunkSize:     int64(1024 * 1024),
		contentLength: contentLength,
	}
	r.log.Sugar().Debug("Start")
	r.next = r.partStream()
	return r, nil
}

func (r *telegramReader) Read(p []byte) (n int, err error) {

	if r.bytesread == r.contentLength {
		r.log.Sugar().Debug("EOF (bytesread == contentLength)")
		return 0, io.EOF
	}

	if r.i >= int64(len(r.buffer)) {
		r.buffer, err = r.next()
		r.log.Debug("Next Buffer", zap.Int64("len", int64(len(r.buffer))))
		if err != nil {
			return 0, err
		}
		if len(r.buffer) == 0 {
			r.next = r.partStream()
			r.buffer, err = r.next()
			if err != nil {
				return 0, err
			}

		}
		r.i = 0
	}
	n = copy(p, r.buffer[r.i:])
	r.i += int64(n)
	r.bytesread += int64(n)
	return n, nil
}

func (r *telegramReader) chunk(offset int64, limit int64) ([]byte, error) {

	req := &tg.UploadGetFileRequest{
		Offset:   offset,
		Limit:    int(limit),
		Location: r.location,
	}

	res, err := r.client.API().UploadGetFile(r.ctx, req)

	if err != nil {
		return nil, err
	}

	switch result := res.(type) {
	case *tg.UploadFile:
		return result.Bytes, nil
	default:
		return nil, fmt.Errorf("unexpected type %T", r)
	}
}

func (r *telegramReader) partStream() func() ([]byte, error) {

	start := r.start
	end := r.end
	offset := start - (start % r.chunkSize)

	firstPartCut := start - offset
	lastPartCut := (end % r.chunkSize) + 1
	partCount := int((end - offset + r.chunkSize) / r.chunkSize)
	currentPart := 1

	readData := func() ([]byte, error) {
		if currentPart > partCount {
			return make([]byte, 0), nil
		}
		res, err := r.chunk(offset, r.chunkSize)
		if err != nil {
			return nil, err
		}
		if len(res) == 0 {
			return res, nil
		} else if partCount == 1 {
			res = res[firstPartCut:lastPartCut]
		} else if currentPart == 1 {
			res = res[firstPartCut:]
		} else if currentPart == partCount {
			res = res[:lastPartCut]
		}

		currentPart++
		offset += r.chunkSize
		r.log.Sugar().Debugf("Part %d/%d", currentPart, partCount)
		return res, nil
	}
	return readData
}
