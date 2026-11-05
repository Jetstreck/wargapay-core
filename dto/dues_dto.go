package dto

type DuesRequest struct {
	UserID      uint   `json:"user_id"`
	Period      string `json:"period"`
	WasteFee    int64  `json:"waste_fee"`
	SecurityFee int64  `json:"security_fee"`
	RTFund      int64  `json:"rt_fund"`
}
