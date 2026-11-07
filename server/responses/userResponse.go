package responses

import "email-summary-tool/database/models"

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func NewUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}
}
