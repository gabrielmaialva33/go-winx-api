package models

import "time"

type PaginationData struct {
	Total         int       `json:"total"`
	PerPage       int       `json:"per_page"`
	OffsetId      int       `json:"offset_id"`
	FirstOffsetId int       `json:"first_offset_id,omitempty"`
	LastOffsetId  int       `json:"last_offset_id,omitempty"`
	OffsetDate    time.Time `json:"offset_date,omitempty"`
	AddOffset     int       `json:"add_offset,omitempty"`
	MaxID         int       `json:"max_id,omitempty"`
	MinID         int       `json:"min_id,omitempty"`
	Search        string    `json:"search,omitempty"`
}

func (m *PaginationData) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"total":           m.Total,
		"per_page":        m.PerPage,
		"offset_id":       m.OffsetId,
		"first_offset_id": m.FirstOffsetId,
		"last_offset_id":  m.LastOffsetId,
		"offset_date":     m.OffsetDate,
		"add_offset":      m.AddOffset,
		"max_id":          m.MaxID,
		"min_id":          m.MinID,
		"search":          m.Search,
	}
}
