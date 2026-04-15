package processor

import (
	"time"

	"github.com/martketplace-vkr/pkg/build/components"
)

type Config struct {
	components.ComponentConfig
	PollInterval          time.Duration `validate:"required"`
	Network               string        `validate:"required"`
	Asset                 string        `validate:"required"`
	TokenDecimals         int32         `validate:"required"`
	ConfirmationsRequired int64         `validate:"required"`
	StartBlock            int64
	BlockBatchSize        int64 `validate:"required"`
}
