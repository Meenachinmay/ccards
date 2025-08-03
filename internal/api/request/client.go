package request

type RegisterCompany struct {
	Name    string `json:"name" binding:"required,min=2,max=255"`
	Email   string `json:"email" binding:"required,email"`
	Address string `json:"address" binding:"max=500"`
	Phone   string `json:"phone" binding:"max=50"`
}

type LoginCompany struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
