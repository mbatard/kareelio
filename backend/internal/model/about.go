package model

type AboutResponse struct {
	Version     string `json:"version"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

var Version = "dev"

func GetAbout() AboutResponse {
	return AboutResponse{
		Version:     Version,
		Name:        "Kareelio",
		Description: "Job application tracker",
	}
}
