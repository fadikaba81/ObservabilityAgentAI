package sumologic

type SearchJobRequest struct {
	Query    string `json: "query"`
	From     string `json: "from"`
	To       string `json: "to"`
	TimeZone string `json: "timeZone"`
}
type SearchJobResponse struct {
	ID string `json: "id"`
}

type SearchJobStatus struct {
	State        string `json: "state"`
	MessageCount int    `json: "messageCount"`
}
