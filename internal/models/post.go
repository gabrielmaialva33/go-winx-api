package models

type Reaction struct {
	Reaction string `json:"reaction,omitempty"`
	Count    int    `json:"count"`
}

type Post struct {
	ImageURL          string     `json:"image_url,omitempty"`
	VideoURL          string     `json:"video_url,omitempty"`
	GroupedID         int64      `json:"grouped_id,omitempty"`
	MessageID         int        `json:"message_id"`
	Date              int        `json:"date"`
	Author            string     `json:"author,omitempty"`
	Reactions         []Reaction `json:"reactions,omitempty"`
	OriginalContent   string     `json:"original_content"`
	ParsedContent     MovieData  `json:"parsed_content"`
	DocumentID        int64      `json:"document_id,omitempty"`
	DocumentSize      int64      `json:"document_size,omitempty"`
	DocumentMessageID int        `json:"document_message_id,omitempty"`
}

func (m *Post) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"image_url":           m.ImageURL,
		"video_url":           m.VideoURL,
		"grouped_id":          m.GroupedID,
		"message_id":          m.MessageID,
		"date":                m.Date,
		"author":              m.Author,
		"reactions":           m.Reactions,
		"original_content":    m.OriginalContent,
		"parsed_content":      m.ParsedContent.ToMap(),
		"document_id":         m.DocumentID,
		"document_size":       m.DocumentSize,
		"document_message_id": m.DocumentMessageID,
	}
}

type PaginatedPosts struct {
	Data       []Post         `json:"data"`
	Pagination PaginationData `json:"pagination"`
}

func (m *PaginatedPosts) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"data":       m.Data,
		"pagination": m.Pagination.ToMap(),
	}
}
