package newrelic

type NerdGraphRequest struct {
	Query string `json:"query"`
}

type NerdGraphResponse struct {
	Data struct {
		Actor struct {
			Account struct {
				NRQL struct {
					Results []map[string]interface{} `json:"results"`
				} `json:"nrql"`
			} `json:"account"`
		} `json:"actor"`
	} `json:"data"`
}
