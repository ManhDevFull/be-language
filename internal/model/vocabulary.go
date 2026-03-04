package model

// Vocabulary keeps one learning unit with EN/RU fields.
type Vocabulary struct {
	ID              int    `json:"id"`
	EnglishWord     string `json:"englishWord"`
	EnglishPhonetic string `json:"englishPhonetic"`
	RussianWord     string `json:"russianWord"`
	RussianPhonetic string `json:"russianPhonetic"`
	PartOfSpeech    string `json:"partOfSpeech"`
	Meaning         string `json:"meaning"`
}

// VocabularyListQuery contains search and pagination arguments.
type VocabularyListQuery struct {
	Query  string
	Limit  int
	Offset int
}
