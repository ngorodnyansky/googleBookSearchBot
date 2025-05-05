package googleBook

type GoogleBookResponce struct {
	Kind       string  `json:"kind"`
	TotalItems int     `json:"totalItems"`
	Items      []Items `json:"items"`
}

type Items struct {
	Kind      string    `json:"kind"`
	SelfLink  string    `json:"selfLink"`
	VolmeInfo VolmeInfo `json:"volumeInfo"`
}

type VolmeInfo struct {
	Titile        string     `json:"title"`
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
