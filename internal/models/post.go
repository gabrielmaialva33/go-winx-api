package models

type Reaction struct {
	Reaction string `json:"reaction,omitempty"`
	Count    int    `json:"count"`
}

type Post struct {
	ImageURL        string                 `json:"image_url,omitempty"`
	VideoURL        string                 `json:"video_url,omitempty"`
	GroupedID       int64                  `json:"grouped_id,omitempty"`
	MessageID       int                    `json:"message_id"`
	Date            string                 `json:"date"`
	Author          string                 `json:"author,omitempty"`
	Reactions       []Reaction             `json:"reactions,omitempty"`
	OriginalContent string                 `json:"original_content"`
	ParsedContent   map[string]interface{} `json:"parsed_content,omitempty"`
	DocumentID      int64                  `json:"document_id,omitempty"`
	DocumentSize    int64                  `json:"document_size,omitempty"`
	MessageDocID    int64                  `json:"message_doc_id,omitempty"`
}

type PaginatedPosts struct {
	Data       []Post         `json:"data"`
	Pagination PaginationData `json:"pagination"`
}
