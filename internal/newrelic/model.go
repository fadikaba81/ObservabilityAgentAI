package newrelic

type NerdGraphRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type NerdGraphResponse struct {
	Data struct {
		Actor struct {
			Account struct {
				NRQL struct {
					Results []map[string]any `json:"results"`
				} `json:"nrql"`
			} `json:"account"`
		} `json:"actor"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}
