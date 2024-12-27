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
	reader.log.Sugar().Debug("Starting Telegram reader")
	reader.log.Sugar().Debug("Content length", contentLength)
	reader.log.Sugar().Debug("Start", start)
	reader.log.Sugar().Debug("End", end)
	reader.next = reader.partStream()
	return reader, nil
}

func (*Reader) Close() error {
	return nil
}

func (r *Reader) Read(p []byte) (int, error) {
	if r.bytesRead == r.contentLength {
		r.log.Sugar().Debug("EOF (bytesRead == contentLength)")
		return 0, io.EOF
	}

	if r.bufferIndex >= int64(len(r.buffer)) {
		var err error
		r.buffer, err = r.next()
		r.log.Debug("Updating buffer", zap.Int64("len", int64(len(r.buffer))))
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
		r.bufferIndex = 0
	}

	n := copy(p, r.buffer[r.bufferIndex:])
	r.bufferIndex += int64(n)
	r.bytesRead += int64(n)
	return n, nil
}

// Gets a chunk of bytes from the file
func (r *Reader) chunk(offset, limit int64) ([]byte, error) {
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
		return nil, fmt.Errorf("unexpected type %T", res)
	}
}

// Helper function to apply cuts to chunks
func (r *Reader) applyPartCut(res []byte, isFirstPart, isLastPart bool, firstPartCut, lastPartCut int64) []byte {
	if isFirstPart && isLastPart {
		return res[firstPartCut:lastPartCut]
	} else if isFirstPart {
		return res[firstPartCut:]
	} else if isLastPart {
		return res[:lastPartCut]
	}
	return res
}

// Creates a stream to split the file into parts (chunks)
func (r *Reader) partStream() func() ([]byte, error) {
	offset := r.start - (r.start % r.chunkSize)
	firstPartCut := r.start - offset
	lastPartCut := (r.end % r.chunkSize) + 1
	partCount := int((r.end - offset + r.chunkSize) / r.chunkSize)
	currentPart := 1

	return func() ([]byte, error) {
		if currentPart > partCount {
			return make([]byte, 0), nil
		}

		res, err := r.chunk(offset, r.chunkSize)
		if err != nil {
			return nil, err
		}

		isFirstPart := currentPart == 1
		isLastPart := currentPart == partCount
		res = r.applyPartCut(res, isFirstPart, isLastPart, firstPartCut, lastPartCut)

		currentPart++
		offset += r.chunkSize
		r.log.Sugar().Debugf("Part %d/%d", currentPart, partCount)
		return res, nil
	}
}
