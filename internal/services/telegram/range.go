package telegram

import (
	"fmt"
	"strconv"
	"strings"
)

type Range struct {
	Start int64
	End   int64
}

func ParseRange(contentLength int64, rangeHeader string) ([]Range, error) {
	const bytesPrefix = "bytes="
	if !strings.HasPrefix(rangeHeader, bytesPrefix) {
		return nil, fmt.Errorf("invalid range format: %s", rangeHeader)
	}

	ranges := strings.Split(strings.TrimPrefix(rangeHeader, bytesPrefix), ",")
	parsedRanges := make([]Range, 0, len(ranges))

	for _, part := range ranges {
		parts := strings.Split(strings.TrimSpace(part), "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid range segment: %s", part)
		}

		var start, end int64
		var err error

		if parts[0] == "" {
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid range end: %s", parts[1])
			}
			start = contentLength - end
			end = contentLength - 1
		} else {
			start, err = strconv.ParseInt(parts[0], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid range start: %s", parts[0])
			}

			if parts[1] != "" {
				end, err = strconv.ParseInt(parts[1], 10, 64)
				if err != nil {
					return nil, fmt.Errorf("invalid range end: %s", parts[1])
				}
			} else {
				end = contentLength - 1
			}
		}

		if start < 0 || end >= contentLength || start > end {
			return nil, fmt.Errorf("invalid range: %d-%d", start, end)
		}

		parsedRanges = append(parsedRanges, Range{Start: start, End: end})
	}

	return parsedRanges, nil
}
