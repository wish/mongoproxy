package command

type Collation struct {
	Locale          string `bson:"locale,omitempty"`          // The locale
	CaseLevel       bool   `bson:"caseLevel,omitempty"`       // The case level
	CaseFirst       string `bson:"caseFirst,omitempty"`       // The case ordering
	Strength        int    `bson:"strength,omitempty"`        // The number of comparison levels to use
	NumericOrdering bool   `bson:"numericOrdering,omitempty"` // Whether to order numbers based on numerical order and not collation order
	Alternate       string `bson:"alternate,omitempty"`       // Whether spaces and punctuation are considered base characters
	MaxVariable     string `bson:"maxVariable,omitempty"`     // Which characters are affected by alternate: "shifted"
	Normalization   bool   `bson:"normalization,omitempty"`   // Causes text to be normalized into Unicode NFD
	Backwards       bool   `bson:"backwards,omitempty"`       // Causes secondary differences to be considered in reverse order, as it is done in the French language
}
