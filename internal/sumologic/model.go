package sumologic

type SearchJobRequest struct {
	Query    string `json:"query"`
	From     string `json:"from"`
	To       string `json:"to"`
	TimeZone string `json:"timeZone"`
}
type SearchJobResponse struct {
	ID string `json:"id"`
}

type SearchJobStatus struct {
	State        string `json:"state"`
	MessageCount int    `json:"messageCount"`
}

type SearchJobMessage struct {
	Map map[string]string `json:"map"`
}

type SearchJobMessageResponse struct {
	Fields   []SearchJobField   `json:"fields"`
	Messages []SearchJobMessage `json:"messages"`
}

type SearchJobField struct {
	Name string `json:"name"`
}
