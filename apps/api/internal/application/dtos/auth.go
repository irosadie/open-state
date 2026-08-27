package dtos

// Auth request DTOs

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Auth response DTOs

type UserDTO struct {
	ID     string  `json:"id"`
	Email  string  `json:"email"`
	Name   string  `json:"name"`
	Role   string  `json:"role"`
	Status string  `json:"status"`
	Photo  *string `json:"photo"`
}

type LoginDTO struct {
	AccessToken string  `json:"accessToken"`
	User        UserDTO `json:"user"`
}
