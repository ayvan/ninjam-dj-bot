package tracks

type Author struct {
	Model
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}
