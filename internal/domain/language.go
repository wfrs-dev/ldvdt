package domain

// Language is the practice language.
type Language string

const (
	LanguageSpanish Language = "es"
	LanguageEnglish Language = "en"
)

// Valid reports whether the language is a known practice language.
func (l Language) Valid() bool {
	return l == LanguageSpanish || l == LanguageEnglish
}
