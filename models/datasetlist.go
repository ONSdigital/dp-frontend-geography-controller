package models

// DatasetListResults contains an array of datasets which can be paginated ----------------- editions of a datasets
type DatasetListResults struct {
	Items      []DatasetList `json:"items"`
	Count      int           `json:"count"`
	Offset     int           `json:"offset"`
	Limit      int           `json:"limit"`
	TotalCount int           `json:"total_count"`
}

// DatasetList containing links to all possible codes
type DatasetList struct {
	Editions []Edition       `json:"editions"`
	Label    string          `json:"dimension_label"`
	Links    DatasetListLink `json:"links"`
}

// Edition contains links for all edition of a dataset
type Edition struct {
	Links EditionsLink `json:"links"`
}

// EditionsLink contains links for a edition resource
type EditionsLink struct {
	Self   *Link `json:"self"`
	Latest *Link `json:"latest_version"`
}

// DatasetListLink contains links for a dataset resource
type DatasetListLink struct {
	Self *Link `json:"self"`
}

// DatasetMetadata contains metadata for a dataset edition
type DatasetMetadata struct {
	Description string `json:"description"`
	Title       string `json:"title"`
}
