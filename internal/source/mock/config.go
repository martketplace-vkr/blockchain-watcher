package mock

type Config struct {
	LatestBlock int64      `validate:"required"`
	Transfers   []Transfer `validate:"-"`
}

type Transfer struct {
	TxHash      string `validate:"required"`
	LogIndex    int64  `validate:"required"`
	BlockNumber int64  `validate:"required"`
	FromAddress string `validate:"required"`
	ToAddress   string `validate:"required"`
	Amount      string `validate:"required"`
	Asset       string `validate:"required"`
	Network     string `validate:"required"`
}
