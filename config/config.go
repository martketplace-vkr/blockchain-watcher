package config

import (
	"github.com/martketplace-vkr/blockchain-watcher/internal/app/cmp/inbox"
	"github.com/martketplace-vkr/blockchain-watcher/internal/app/cmp/outbox"
	"github.com/martketplace-vkr/blockchain-watcher/internal/app/cmp/processor"
	"github.com/martketplace-vkr/blockchain-watcher/internal/source/trongrid"
	"github.com/martketplace-vkr/pkg/build/components/pgxsqlxcomponent"
	"github.com/martketplace-vkr/pkg/kafkaconnector"
)

type Config struct {
	Postgres  pgxsqlxcomponent.Config     `validate:"required"`
	Inbox     inbox.Config                `validate:"required"`
	Outbox    outbox.Config               `validate:"required"`
	Processor processor.Config            `validate:"required"`
	Tron      trongrid.Config             `validate:"required"`
	Kafka     kafkaconnector.ClientConfig `validate:"required"`
}
