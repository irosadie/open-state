package dtos

type AppInfoDTO struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

type HealthDTO struct {
	Status string `json:"status"`
}
