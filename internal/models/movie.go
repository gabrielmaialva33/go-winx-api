package models

type MovieData struct {
	Title            string   `json:"title"`
	ReleaseDate      string   `json:"release_date"`
	CountryOfOrigin  []string `json:"country_of_origin"`
	FlagsOfOrigin    []string `json:"flags_of_origin"`
	Directors        []string `json:"directors"`
	Writers          []string `json:"writers"`
	Cast             []string `json:"cast"`
	Languages        []string `json:"languages"`
	FlagsOfLanguage  []string `json:"flags_of_language"`
	Subtitles        []string `json:"subtitles"`
	FlagsOfSubtitles []string `json:"flags_of_subs"`
	Genres           []string `json:"genres"`
	Tags             []string `json:"tags"`
	Synopsis         string   `json:"synopsis"`
	Curiosities      string   `json:"curiosities"`
}

func (m *MovieData) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"title":             m.Title,
		"release_date":      m.ReleaseDate,
		"country_of_origin": m.CountryOfOrigin,
		"flags_of_origin":   m.FlagsOfOrigin,
		"directors":         m.Directors,
		"writers":           m.Writers,
		"cast":              m.Cast,
		"languages":         m.Languages,
		"flags_of_language": m.FlagsOfLanguage,
		"subtitles":         m.Subtitles,
		"flags_of_subs":     m.FlagsOfSubtitles,
		"genres":            m.Genres,
		"tags":              m.Tags,
		"synopsis":          m.Synopsis,
		"curiosities":       m.Curiosities,
	}
}
