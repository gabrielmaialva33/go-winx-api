package telegram

import (
	"context"
	"fmt"
	"io"

	"go-winx-api/internal/utils"

	"github.com/celestix/gotgproto"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

const defaultChunkSize = int64(1024 * 1024) // 1MB

type Reader struct {
	ctx           context.Context
	log           *zap.Logger
	client        *gotgproto.Client
	location      *tg.InputDocumentFileLocation
	start         int64
	end           int64
	next          func() ([]byte, error)
	buffer        []byte
	bytesRead     int64
	chunkSize     int64
	bufferIndex   int64
	contentLength int64
}

func NewReader(
	ctx context.Context,
	client *gotgproto.Client,
	location *tg.InputDocumentFileLocation,
	start, end, contentLength int64,
) (io.ReadCloser, error) {
	reader := &Reader{
		ctx:           ctx,
		log:           utils.Logger.Named("telegram_reader"),
		location:      location,
		client:        client,
		start:         start,
		end:           end,
		chunkSize:     defaultChunkSize,
		contentLength: contentLength,
	}

	reader.log.Sugar().Debug("starting telegram reader")
	reader.log.Sugar().Debug("content length", contentLength)
	reader.log.Sugar().Debug("start", start)
	reader.log.Sugar().Debug("end", end)

	reader.next = reader.partStream()
	return reader, nil
}

func (*Reader) Close() error {
	return nil
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.bytesRead == r.contentLength {
		r.log.Sugar().Debug("EOF (bytesRead equals contentLength)")
		return 0, io.EOF
	}

	if r.bufferIndex >= int64(len(r.buffer)) {
		r.buffer, err = r.next()
		r.log.Debug("next Buffer", zap.Int64("len", int64(len(r.buffer))))

		if err != nil {
			r.log.Error("error fetching next buffer", zap.Error(err))
			return 0, err
		}

		if len(r.buffer) == 0 {
			r.log.Sugar().Warn("buffer is empty, resetting partStream")
			r.next = r.partStream()
			r.buffer, err = r.next()
			if err != nil {
				r.log.Error("error fetching buffer after resetting partStream", zap.Error(err))
				return 0, err
			}
		}
		r.bufferIndex = 0
	}

	n = copy(p, r.buffer[r.bufferIndex:])
	r.bufferIndex += int64(n)
	r.bytesRead += int64(n)

	r.log.Debug("read buffer", zap.Int("bytes", n), zap.Int64("bytesRead", r.bytesRead))
	return n, nil
}

func (r *Reader) chunk(offset int64, limit int64) ([]byte, error) {
	r.log.Sugar().Debugf("requesting chunk: Offset=%d, Limit=%d", offset, limit)
	req := &tg.UploadGetFileRequest{
		Offset:   offset,
		Limit:    int(limit),
		Location: r.location,
	}

	res, err := r.client.API().UploadGetFile(r.ctx, req)
	if err != nil {
		r.log.Error("failed to fetch chunk", zap.Error(err))
		return nil, err
	}

	switch result := res.(type) {
	case *tg.UploadFile:
		r.log.Sugar().Debugf("chunk received: %d bytes", len(result.Bytes))
		if len(result.Bytes) == 0 {
			r.log.Warn("empty chunk received despite no error")
		}
		return result.Bytes, nil
	default:
		err := fmt.Errorf("unexpected type %T from UploadGetFile", res)
		r.log.Error("failed to fetch chunk", zap.Error(err))
		return nil, err
	}
}

func (r *Reader) partStream() func() ([]byte, error) {
	start := r.start
	end := r.end
	offset := start - (start % r.chunkSize)

	firstPartCut := start - offset
	lastPartCut := (end % r.chunkSize) + 1
	partCount := int((end - offset + r.chunkSize) / r.chunkSize)
	currentPart := 1

	return func() ([]byte, error) {
		if currentPart > partCount {
			r.log.Debug("all parts have been read")
			return make([]byte, 0), nil
		}

		r.log.Sugar().Debugf("fetching part %d/%d, Offset=%d", currentPart, partCount, offset)

		res, err := r.chunk(offset, r.chunkSize)
		if err != nil {
			r.log.Error("failed to fetch chunk", zap.Error(err))
			return nil, err
		}

		if len(res) == 0 {
			return res, nil
		}

		if partCount == 1 {
			res = res[firstPartCut:lastPartCut]
		} else if currentPart == 1 {
			res = res[firstPartCut:]
		} else if currentPart == partCount {
			res = res[:lastPartCut]
		}

		currentPart++
		offset += r.chunkSize
		return res, nil
	}
}
