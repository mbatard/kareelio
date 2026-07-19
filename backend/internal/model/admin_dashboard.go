package model

type AdminDashboard struct {
	Users          AdminDashboardUsers          `json:"users"`
	Applications   AdminDashboardApplications   `json:"applications"`
	Funnels        AdminDashboardFunnels        `json:"funnels"`
	ByStatus       map[string]int               `json:"by_status"`
	BySource       map[string]int               `json:"by_source"`
	ByRemote       map[string]int               `json:"by_remote"`
	ByPriority     map[string]int               `json:"by_priority"`
}

type AdminDashboardUsers struct {
	Total    int `json:"total"`
	Active   int `json:"active"`
	Disabled int `json:"disabled"`
}

type AdminDashboardApplications struct {
	Total               int     `json:"total"`
	CreatedLast7Days    int     `json:"created_last_7_days"`
	CreatedLast30Days   int     `json:"created_last_30_days"`
	AveragePerActiveUser float64 `json:"average_per_active_user"`
}

type AdminDashboardFunnels struct {
	ResponseRate float64 `json:"response_rate"`
	InterviewRate float64 `json:"interview_rate"`
	TestRate     float64 `json:"test_rate"`
	OfferRate    float64 `json:"offer_rate"`
}
