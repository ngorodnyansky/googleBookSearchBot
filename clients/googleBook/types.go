package googleBook

type GoogleBookResponse struct {
	Kind       string `json:"kind"`
	TotalItems int    `json:"totalItems"`
	Items      []Item `json:"items"`
}

type Item struct {
	Kind       string     `json:"kind"`
	SelfLink   string     `json:"selfLink"`
	VolumeInfo VolumeInfo `json:"volumeInfo"`
}

type VolumeInfo struct {
	Title         string     `json:"title"`
	Subtitle      string     `json:"subtitle"`
	Authors       []string   `json:"authors"`
	Description   string     `json:"description"`
	AverageRating float32    `json:"averageRating"`
	ImageLinks    ImageLinks `json:"imageLinks"`
}

type ImageLinks struct {
	Thumbnail string `json:"thumbnail"`
	Small     string `json:"small"`
	Medium    string `json:"medium"`
}
