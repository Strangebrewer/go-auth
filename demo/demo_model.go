package demo

type DemoRegisterResponse struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
