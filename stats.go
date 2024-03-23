package blog

type PageStats struct {
	URL    string `json:"url"`
	Visits int    `json:"visits"`
}

type RefererStats struct {
	Share   string `json:"share"`
	Referer string `json:"referrer"`
	Visits  int    `json:"visits"`
}

type CountryStats struct {
	Share   string `json:"share"`
	Country string `json:"country"`
	Visits  int    `json:"visits"`
}

type BrowserStats struct {
	Share   string `json:"share"`
	Browser string `json:"browser"`
	Visits  int    `json:"visits"`
}

type OSStats struct {
	Share  string `json:"share"`
	OS     string `json:"os"`
	Visits int    `json:"visits"`
}

type StatService interface {
	Insert(e *Event) error
	ImportIP2LocationDB(token string) error
	GetCountryFromIP(ip string) (string, error)
	Top10VisitedPages() ([]PageStats, error)
	Top10Referers(domain string) ([]RefererStats, error)
	Top10Countries() ([]CountryStats, error)
	Top10Browsers() ([]BrowserStats, error)
	Top10OperatingSystems() ([]OSStats, error)
}
