package utils

import (
	"go-winx-api/internal/models"
	"regexp"
	"strings"
)

// FieldDefinition defines the structure for each field we want to extract
type FieldDefinition struct {
	Field       string
	Labels      []string
	Regexes     []*regexp.Regexp
	Process     func(match []string, data *models.MovieData, buffer *[]string)
	IsMultiline bool
}

// FieldDefinitions list all the fields we want to parse from the content
var fieldDefinitions = []FieldDefinition{
	{
		Field:  "title",
		Labels: []string{"📺", "Título:"},
		Regexes: []*regexp.Regexp{
			regexp.MustCompile(`^.*?(?:📺|Título:)\s*(.*?)(?:\s*#(\d{4})y)?(?:💄)?$`),
		},
		Process: func(match []string, data *models.MovieData, buffer *[]string) {
			data.Title = strings.TrimSpace(match[1])
			if len(match) > 2 && match[2] != "" {
				data.ReleaseDate = strings.TrimSpace(match[2])
			} else {
				data.ReleaseDate = ""
			}
		},
		IsMultiline: false,
	},
	{
		Field:  "country_of_origin",
		Labels: []string{"País de Origem:", "📍 País de Origem:", "Pais de Origem:"},
		Regexes: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^.*?Pa[íi]s de Origem:\s*(.*)$`),
		},
		Process: func(match []string, data *models.MovieData, buffer *[]string) {
			processCountryOfOrigin(match[1], data)
		},
		IsMultiline: false,
	},
	{
		Field:  "directors",
		Labels: []string{"Direção:", "Diretor:", "👑 Direção:", "👑 Direção/Roteiro:"},
		Regexes: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^.*?(?:Direção|Diretor|Direção\/Roteiro):\s*(.*)$`),
		},
		Process: func(match []string, data *models.MovieData, buffer *[]string) {
			processDirectors(match[1], data, buffer)
		},
		IsMultiline: false,
	},
	{
		Field:  "writers",
		Labels: []string{"Roteiro:", "Roteirista:", "Roteiristas:", "✏️ Roteirista:", "✏️ Roteiristas:"},
		Regexes: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^.*?(?:Roteiro|Roteirista|Roteiristas):\s*(.*)$`),
		},
		Process: func(match []string, data *models.MovieData, buffer *[]string) {
			processWriters(match[1], data)
		},
		IsMultiline: false,
	},
	{
		Field:  "cast",
		Labels: []string{"Elenco:", "✨ Elenco:"},
		Regexes: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^.*?Elenco:\s*(.*)$`),
		},
		Process: func(match []string, data *models.MovieData, buffer *[]string) {
			data.Cast = parseHashList(match[1]) // “#actor1 #actor2”
		},
		IsMultiline: false,
	},
	{
		Field:  "languages",
		Labels: []string{"Idioma:", "Idiomas:", "📣 Idiomas:", "💬 Idiomas:"},
		Regexes: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^.*?(?:Idiomas?|Idioma):\s*(.*)$`),
		},
		Process: func(match []string, data *models.MovieData, buffer *[]string) {
			processLanguages(match[1], data)
		},
		IsMultiline: false,
	},
	{
		Field:  "subtitles",
		Labels: []string{"Legenda:", "Legendado:", "💬 Legendado:"},
		Regexes: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^.*?(?:Legenda|Legendado):\s*(.*)$`),
		},
		Process: func(match []string, data *models.MovieData, buffer *[]string) {
			processSubtitles(match[1], data)
		},
		IsMultiline: false,
	},
	{
		Field:  "genres",
		Labels: []string{"Gênero:", "Gêneros:", "🎭 Gêneros:"},
		Regexes: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^.*?(?:Gêneros?|Gênero):\s*(.*)$`),
		},
		Process: func(match []string, data *models.MovieData, buffer *[]string) {
			data.Genres = parseHashList(match[1])
		},
		IsMultiline: false,
	},
	{
		Field:  "synopsis",
		Labels: []string{"Sinopse", "🗣 Sinopse:"},
		Regexes: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^.*?(?:Sinopse|🗣 Sinopse)[:：]?\s*(.*)$`),
		},
		Process: func(match []string, data *models.MovieData, buffer *[]string) {
			if strings.TrimSpace(match[1]) != "" {
				*buffer = append(*buffer, strings.TrimSpace(match[1]))
			}
		},
		IsMultiline: true,
	},
	{
		Field:  "curiosities",
		Labels: []string{"Curiosidades:", "💡 Curiosidades:"},
		Regexes: []*regexp.Regexp{
			regexp.MustCompile(`(?i)^.*?(?:Curiosidades)[:：]?\s*(.*)$`),
		},
		Process: func(match []string, data *models.MovieData, buffer *[]string) {
			if strings.TrimSpace(match[1]) != "" {
				*buffer = append(*buffer, strings.TrimSpace(match[1]))
			}
		},
		IsMultiline: true,
	},
}

// processCountryOfOrigin processes the country of origin field
func processCountryOfOrigin(value string, data *models.MovieData) {
	data.CountryOfOrigin = []string{}
	data.FlagsOfOrigin = []string{}

	parts := strings.Split(value, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		flags := extractFlags(part)
		if flags != "" {
			data.FlagsOfOrigin = append(data.FlagsOfOrigin, flags)
		}

		countryName := removeFlags(part)
		countryName = strings.TrimPrefix(countryName, "#")
		countryName = strings.TrimSpace(countryName)
		if countryName != "" {
			data.CountryOfOrigin = append(data.CountryOfOrigin, countryName)
		}
	}
}

// processDirectors processes the directors field
func processDirectors(value string, data *models.MovieData, buffer *[]string) {
	directors := parseHashList(value)
	data.Directors = append(data.Directors, directors...)
}

// processWriters processes the writers field
func processWriters(value string, data *models.MovieData) {
	if data.Writers == nil {
		data.Writers = []string{}
	}
	data.Writers = append(data.Writers, parseHashList(value)...)
}

// processLanguages processes the languages field
func processLanguages(value string, data *models.MovieData) {
	data.Languages = []string{}
	data.FlagsOfLanguage = []string{}

	parts := strings.Split(value, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		flags := extractFlags(part)
		if flags != "" {
			data.FlagsOfLanguage = append(data.FlagsOfLanguage, flags)
		}

		lang := removeFlags(part)
		lang = strings.ReplaceAll(lang, "#", "")
		lang = strings.TrimSpace(lang)
		if lang != "" {
			data.Languages = append(data.Languages, lang)
		}
	}
}

// processSubtitles processes the subtitles field
func processSubtitles(value string, data *models.MovieData) {
	data.Subtitles = []string{}
	data.FlagsOfSubtitles = []string{}
	parts := strings.Split(value, "|")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		flags := extractFlags(part)
		if flags != "" {
			data.FlagsOfSubtitles = append(data.FlagsOfSubtitles, flags)
		}
		sub := removeFlags(part)
		sub = strings.ReplaceAll(sub, "#", "")
		sub = strings.TrimSpace(sub)
		if sub != "" {
			data.Subtitles = append(data.Subtitles, sub)
		}
	}
}

// extractFlags extracts emoji flags from a string
func extractFlags(s string) string {
	var flags []rune
	for _, r := range s {
		if isEmoji(r) {
			flags = append(flags, r)
		}
	}
	return string(flags)
}

// removeFlags removes emoji flags from a string
func removeFlags(s string) string {
	var out []rune
	for _, r := range s {
		if !isEmoji(r) {
			out = append(out, r)
		}
	}
	return strings.TrimSpace(string(out))
}

// isEmoji checks if a rune is an emoji
func isEmoji(r rune) bool {
	emojiRanges := []struct{ start, end rune }{
		{0x1F600, 0x1F64F}, // Emoticons
		{0x1F300, 0x1F5FF}, // Symbols & Pictographs
		{0x1F680, 0x1F6FF}, // Transport & Map
		{0x2600, 0x26FF},   // Misc Symbols
	}

	for _, er := range emojiRanges {
		if r >= er.start && r <= er.end {
			return true
		}
	}
	return false
}

// parseHashList parses a string with '#' delimiters
func parseHashList(s string) []string {
	parts := strings.Split(s, "#")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// endOfFieldMarkers marks the end of a field in the content
var endOfFieldMarkers = []string{
	"▶", "▶️", "Para outros conteúdos",
	"💡 Curiosidades:", "🥇 Prêmios:", "🥈 Prêmios:",
	"Prêmios:", "Clique Para Entrar", "🚨 Para outros conteúdos",
	"📣 Idiomas:", "💬 Legendado:", "📣", "💬", "#", "✨ Elenco:", "📢",
}

// isEndOfField checks if a line marks the end of a field
func isEndOfField(line string) bool {
	for _, marker := range endOfFieldMarkers {
		if strings.HasPrefix(line, marker) {
			return true
		}
	}
	return false
}

// isNewField checks if a line marks the beginning of a new field
func isNewField(line string) bool {
	for _, fd := range fieldDefinitions {
		for _, lab := range fd.Labels {
			if strings.Contains(line, lab) {
				return true
			}
		}
	}
	return false
}

// finalizeMultilineField finalizes the processing of a multiline field
func finalizeMultilineField(fd *FieldDefinition, data *models.MovieData, buffer []string) {
	text := strings.Join(buffer, "\n")
	switch fd.Field {
	case "synopsis":
		data.Synopsis = strings.TrimSpace(text)
	case "curiosities":
		data.Curiosities = strings.TrimSpace(text)
	}
}

// ParseMessageContent parses the content of a message into a MovieData struct
func ParseMessageContent(content string) *models.MovieData {
	data := &models.MovieData{
		CountryOfOrigin:  []string{},
		FlagsOfOrigin:    []string{},
		Directors:        []string{},
		Writers:          []string{},
		Cast:             []string{},
		Languages:        []string{},
		FlagsOfLanguage:  []string{},
		Subtitles:        []string{},
		FlagsOfSubtitles: []string{},
		Genres:           []string{},
		Tags:             []string{},
	}
	lines := strings.Split(content, "\n")

	var multilineBuffer []string
	var currentField *FieldDefinition

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		if currentField != nil {
			if isNewField(line) || isEndOfField(line) {
				finalizeMultilineField(currentField, data, multilineBuffer)
				currentField = nil
				multilineBuffer = []string{}
				i--
				continue
			} else {
				multilineBuffer = append(multilineBuffer, line)
				continue
			}
		}

		if strings.HasPrefix(line, "#") {
			tags := parseHashList(line)
			data.Tags = append(data.Tags, tags...)
			continue
		}

		matched := false
		for idx := range fieldDefinitions {
			fd := &fieldDefinitions[idx]
			for _, rx := range fd.Regexes {
				sub := rx.FindStringSubmatch(line)
				if len(sub) > 0 {
					fd.Process(sub, data, &multilineBuffer)
					if fd.IsMultiline {
						currentField = fd
					}
					matched = true
					break
				}
			}
			if matched {
				break
			}
		}
	}

	if currentField != nil && len(multilineBuffer) > 0 {
		finalizeMultilineField(currentField, data, multilineBuffer)
	}

	return data
}
