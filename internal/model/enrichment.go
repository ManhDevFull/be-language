package model

// EnrichmentResult stores processed data for one input value.
type EnrichmentResult struct {
	Input           string `json:"input"`
	EnglishWord     string `json:"englishWord"`
	EnglishPhonetic string `json:"englishPhonetic"`
	RussianWord     string `json:"russianWord"`
	RussianPhonetic string `json:"russianPhonetic"`
	PartOfSpeech    string `json:"partOfSpeech"`
	Meaning         string `json:"meaning"`
}
