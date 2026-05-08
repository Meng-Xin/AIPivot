package types

// --- Requests ---

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	NickName string `json:"nick_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// --- Show Types (脱敏展示) ---

type ShowUser struct {
	ID        int64  `json:"id"`
	UUID      string `json:"uuid"`
	TenantID  int64  `json:"tenantId"`
	Email     string `json:"email"`
	NickName  string `json:"nickName"`
	Role      string `json:"role"`
	Status    string `json:"status"`
	LastLogin int64  `json:"lastLogin"`
	CreatedAt int64  `json:"createdAt"`
}

// --- Response Data ---

type LoginData struct {
	Token    string   `json:"token"`
	ExpireAt int64    `json:"expireAt"`
	User     ShowUser `json:"user"`
}

type LoginResponse struct {
	Code      int32     `json:"code"`
	Msg       string    `json:"msg"`
	Timestamp int64     `json:"timestamp"`
	Data      LoginData `json:"data"`
}
