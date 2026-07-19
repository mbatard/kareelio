package model

type AboutResponse struct {
	Version     string `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GoVersion   string `json:"go_version"`
}

func GetAbout() AboutResponse {
	return AboutResponse{
		Version:     "0.1.0",
		Name:        "Kareelio",
		Description: "Job application tracker",
		GoVersion:   "1.22",
	}
}
