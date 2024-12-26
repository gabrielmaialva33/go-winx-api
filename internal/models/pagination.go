package models

import "time"

type PaginationData struct {
	Total         int       `json:"total"`
	PerPage       int       `json:"per_page"`
	OffsetID      int       `json:"offset_id"`
	FirstOffsetID int       `json:"first_offset_id,omitempty"`
	LastOffsetID  int       `json:"last_offset_id,omitempty"`
	OffsetDate    time.Time `json:"offset_date,omitempty"`
	AddOffset     int       `json:"add_offset,omitempty"`
	MaxID         int       `json:"max_id,omitempty"`
	MinID         int       `json:"min_id,omitempty"`
	Search        string    `json:"search,omitempty"`
}
