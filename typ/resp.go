package typ

type ListResp struct {
	Count int64       `json:"count"`
	List  interface{} `json:"list"`
}

type LoginResp struct {
	Token    string `json:"token"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}
