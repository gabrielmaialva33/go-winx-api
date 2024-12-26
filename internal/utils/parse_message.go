package utils

import (
	"go-winx-api/internal/models"
	"regexp"
	"strings"
	"unicode"
)

// FieldDefinition struct to define a field and its processing
type FieldDefinition struct {
	Field       string
	Labels      []string
	Regex       []*regexp.Regexp
	Process     func(match []string, data *models.MovieData, buffer *[]string)
	IsMultiline bool
}

// IsEmoji checks if a character is an emoji
func IsEmoji(character rune) bool {
	if unicode.IsSymbol(character) {
		return true
	}
	if character >= 0x1F1E6 && character <= 0x1F1FF {
		return true
	}
	return false
}

// ProcessTitle Process functions for each field
func ProcessTitle(match []string, data *models.MovieData, buffer *[]string) {
	fullTitle := strings.TrimSpace(match[1])
	year := match[2]

	if strings.Contains(fullTitle, "#") {
		titleParts := strings.Split(fullTitle, "#")
		data.Title = strings.TrimSpace(titleParts[0])
		if len(titleParts) > 1 && year == "" {
			extractedYear := strings.TrimSpace(titleParts[1])
			data.ReleaseDate = nonDigitRegex.ReplaceAllString(extractedYear, "")
		}
	} else {
		data.Title = fullTitle
	}
	if year != "" {
		data.ReleaseDate = nonDigitRegex.ReplaceAllString(year, "")
	}
}

func ProcessCountryOfOrigin(match []string, data *models.MovieData, buffer *[]string) {
	countries := splitAndTrim(match[1], "|")
	data.CountryOfOrigin = []string{}
	data.FlagsOfOrigin = []string{}
	for _, country := range countries {
		var flags string
		var countryName string
		for _, char := range country {
			if IsEmoji(char) {
				flags += string(char)
			} else {
				countryName += string(char)
			}
		}
		countryName = strings.TrimSpace(strings.TrimPrefix(countryName, "#"))
		if flags != "" {
			data.FlagsOfOrigin = append(data.FlagsOfOrigin, flags)
		}
		if countryName != "" {
			data.CountryOfOrigin = append(data.CountryOfOrigin, countryName)
		}
	}
}

func ProcessDirectors(match []string, data *models.MovieData, buffer *[]string) {
	names := splitAndTrim(match[1], "#")
	if strings.Contains(match[0], "Direção/Roteiro") {
		data.Directors = append(data.Directors, names...)
		data.Writers = append(data.Writers, names...)
	} else {
		data.Directors = append(data.Directors, names...)
	}
}

func ProcessWriters(match []string, data *models.MovieData, buffer *[]string) {
	writers := splitAndTrim(match[1], "#")
	data.Writers = append(data.Writers, writers...)
}

func ProcessCast(match []string, data *models.MovieData, buffer *[]string) {
	data.Cast = splitAndTrim(match[1], "#")
}

func ProcessLanguages(match []string, data *models.MovieData, buffer *[]string) {
	languages := splitAndTrim(match[1], "|")
	data.Languages = []string{}
	data.FlagsOfLanguage = []string{}
	for _, language := range languages {
		var flags string
		var languageName string
		for _, char := range language {
			if IsEmoji(char) {
				flags += string(char)
			} else {
				languageName += string(char)
			}
		}
		languageName = strings.ReplaceAll(strings.TrimSpace(strings.TrimPrefix(languageName, "#")), "#", "")
		if flags != "" {
			data.FlagsOfLanguage = append(data.FlagsOfLanguage, flags)
		}
		if languageName != "" {
			data.Languages = append(data.Languages, languageName)
		}
	}
}

func ProcessSubtitles(match []string, data *models.MovieData, buffer *[]string) {
	subtitles := splitAndTrim(match[1], "|")
	data.Subtitles = []string{}
	data.FlagsOfSubtitles = []string{}
	for _, subtitle := range subtitles {
		var flags string
		var subtitleLanguage string
		for _, char := range subtitle {
			if IsEmoji(char) {
				flags += string(char)
			} else {
				subtitleLanguage += string(char)
			}
		}
		subtitleLanguage = strings.ReplaceAll(strings.TrimSpace(strings.TrimPrefix(subtitleLanguage, "#")), "#", "")
		if flags != "" {
			data.FlagsOfSubtitles = append(data.FlagsOfSubtitles, flags)
		}
		if subtitleLanguage != "" {
			data.Subtitles = append(data.Subtitles, subtitleLanguage)
		}
	}
}

func ProcessGenres(match []string, data *models.MovieData, buffer *[]string) {
	data.Genres = splitAndTrim(match[1], "#")
}

func ProcessMultiline(match []string, data *models.MovieData, buffer *[]string) {
	if match[1] != "" && strings.TrimSpace(match[1]) != "" {
		*buffer = append(*buffer, strings.TrimSpace(match[1]))
	}
}

// Helper functions
func splitAndTrim(input, sep string) []string {
	parts := strings.Split(input, sep)
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

var nonDigitRegex = regexp.MustCompile(`\D`)

// Field definitions
var fieldDefinitions = []FieldDefinition{
	{
		Field:       "title",
		Labels:      []string{"📺", "Título:"},
		Regex:       []*regexp.Regexp{regexp.MustCompile(`^.*?(?:📺|Título:)\s*(.*?)(?:\s*[-—:]?\s*#(\d{4}y?)?.*?)?$`)},
		Process:     ProcessTitle,
		IsMultiline: false,
	},
	{
		Field:       "country_of_origin",
		Labels:      []string{"País de Origem:", "📍 País de Origem:", "Pais de Origem:"},
		Regex:       []*regexp.Regexp{regexp.MustCompile(`^.*?Pa[íi]s de Origem:\s*(.*)$`)},
		Process:     ProcessCountryOfOrigin,
		IsMultiline: false,
	},
	{
		Field:       "directors",
		Labels:      []string{"Direção:", "Diretor:", "👑 Direção:", "👑 Direção/Roteiro:"},
		Regex:       []*regexp.Regexp{regexp.MustCompile(`^.*?(?:Direção|Diretor|Direção/Roteiro):\s*(.*)$`)},
		Process:     ProcessDirectors,
		IsMultiline: false,
	},
	{
		Field:       "writers",
		Labels:      []string{"Roteiro:", "Roteirista:", "Roteiristas:", "✏️ Roteirista:", "✏️ Roteiristas:"},
		Regex:       []*regexp.Regexp{regexp.MustCompile(`^.*?(?:Roteiro|Roteirista|Roteiristas):\s*(.*)$`)},
		Process:     ProcessWriters,
		IsMultiline: false,
	},
	{
		Field:       "cast",
		Labels:      []string{"Elenco:", "✨ Elenco:"},
		Regex:       []*regexp.Regexp{regexp.MustCompile(`^.*?Elenco:\s*(.*)$`)},
		Process:     ProcessCast,
		IsMultiline: false,
	},
	{
		Field:       "languages",
		Labels:      []string{"Idioma:", "Idiomas:", "📣 Idiomas:", "💬 Idiomas:"},
		Regex:       []*regexp.Regexp{regexp.MustCompile(`^.*?(?:Idiomas?|Idioma):\s*(.*)$`)},
		Process:     ProcessLanguages,
		IsMultiline: false,
	},
	{
		Field:       "subtitles",
		Labels:      []string{"Legenda:", "Legendado:", "💬 Legendado:"},
		Regex:       []*regexp.Regexp{regexp.MustCompile(`^.*?(?:Legenda|Legendado):\s*(.*)$`)},
		Process:     ProcessSubtitles,
		IsMultiline: false,
	},
	{
		Field:       "genres",
		Labels:      []string{"Gênero:", "Gêneros:", "🎭 Gêneros:"},
		Regex:       []*regexp.Regexp{regexp.MustCompile(`^.*?(?:Gêneros?|Gênero):\s*(.*)$`)},
		Process:     ProcessGenres,
		IsMultiline: false,
	},
	{
		Field:       "synopsis",
		Labels:      []string{"Sinopse", "🗣 Sinopse:", "🗣 Sinopse"},
		Regex:       []*regexp.Regexp{regexp.MustCompile(`^.*?(?:Sinopse|🗣 Sinopse)[:：]?\s*(.*)$`)},
		Process:     ProcessMultiline,
		IsMultiline: true,
	},
	{
		Field:       "curiosities",
		Labels:      []string{"Curiosidades:", "💡 Curiosidades:"},
		Regex:       []*regexp.Regexp{regexp.MustCompile(`^.*?Curiosidades[:：]?\s*(.*)$`)},
		Process:     ProcessMultiline,
		IsMultiline: true,
	},
}

// ParseMessageContent parses the content of a message and returns a models.MovieData struct
func ParseMessageContent(content string) models.MovieData {
	lines := splitAndTrim(content, "\n")

	dataInfo := models.MovieData{}
	var multilineBuffer []string
	currentField := ""

	endOfFieldMarkers := []string{
		"▶",
		"▶️",
		"Para outros conteúdos",
		"💡 Curiosidades:",
		"🥇 Prêmios:",
		"🥈 Prêmios:",
		"Prêmios:",
		"Clique Para Entrar",
		"🚨 Para outros conteúdos",
		"📣 Idiomas:",
		"💬 Legendado:",
		"📣",
		"💬",
		"#",
		"✨ Elenco:",
		"📢",
	}

	lineStartsWithLabel := func(line string, labels []string) bool {
		for _, label := range labels {
			if strings.HasPrefix(line, label) {
				return true
			}
		}
		return false
	}

	for _, line := range lines {
		if line == "" {
			continue
		}

		if currentField != "" {
			isNewField := false
			for _, fieldDef := range fieldDefinitions {
				if lineStartsWithLabel(line, fieldDef.Labels) {
					isNewField = true
					break
				}
			}
			isEndOfField := lineStartsWithLabel(line, endOfFieldMarkers)
			if isNewField || isEndOfField {
				switch currentField {
				case "synopsis":
					dataInfo.Synopsis = strings.Join(multilineBuffer, " ")
				case "curiosities":
					dataInfo.Curiosities = strings.Join(multilineBuffer, " ")
				}
				currentField = ""
				multilineBuffer = []string{}
				if isNewField {
					continue
				}
			} else {
				multilineBuffer = append(multilineBuffer, line)
				continue
			}
		}

		if strings.HasPrefix(line, "#") {
			tags := splitAndTrim(line, "#")
			dataInfo.Tags = append(dataInfo.Tags, tags...)
			continue
		}

		for _, fieldDef := range fieldDefinitions {
			for _, regex := range fieldDef.Regex {
				match := regex.FindStringSubmatch(line)
				if match != nil {
					fieldDef.Process(match, &dataInfo, &multilineBuffer)
					if fieldDef.IsMultiline {
						currentField = fieldDef.Field
						multilineBuffer = []string{}
						if match[1] != "" && strings.TrimSpace(match[1]) != "" {
							multilineBuffer = append(multilineBuffer, strings.TrimSpace(match[1]))
						}
					}
					break
				}
			}
			if currentField != "" {
				break
			}
		}
	}

	if currentField != "" {
		switch currentField {
		case "synopsis":
			dataInfo.Synopsis = strings.Join(multilineBuffer, " ")
		case "curiosities":
			dataInfo.Curiosities = strings.Join(multilineBuffer, " ")
		}
	}

	return dataInfo
}
