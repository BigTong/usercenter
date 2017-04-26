package user

type User struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Gender      string `json:"gender,omitempty"`
	Age         int    `json:"age,omitempty"`
	CreatedTime int64  `json:"createdTime,omitempty"`
	Type        string `json:"type,omitempty"`
}
