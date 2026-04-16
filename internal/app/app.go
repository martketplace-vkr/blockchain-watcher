package app

import (
	"context"

	"github.com/martketplace-vkr/blockchain-watcher/config"
	inboxComponent "github.com/martketplace-vkr/blockchain-watcher/internal/app/cmp/inbox"
	outboxComponent "github.com/martketplace-vkr/blockchain-watcher/internal/app/cmp/outbox"
	processorComponent "github.com/martketplace-vkr/blockchain-watcher/internal/app/cmp/processor"
	repository "github.com/martketplace-vkr/blockchain-watcher/internal/repository/pg"
	watcherservice "github.com/martketplace-vkr/blockchain-watcher/internal/service/watcher"
	trongrid "github.com/martketplace-vkr/blockchain-watcher/internal/source/trongrid"
	"github.com/martketplace-vkr/blockchain-watcher/pkg/eventmapper"
	"github.com/martketplace-vkr/pkg/build"
	"github.com/martketplace-vkr/pkg/build/components/pgxsqlxcomponent"
	"github.com/martketplace-vkr/pkg/kafkaconnector"
	outboxclient "github.com/martketplace-vkr/pkg/outbox"
)

func Run(ctx context.Context, cfg *config.Config) error {
	pg := pgxsqlxcomponent.New(cfg.Postgres)

	kafkaClient := kafkaconnector.NewClient(cfg.Kafka)
	kafkaProducer := kafkaClient.NewSyncProducer()

	outboxCl, err := outboxclient.NewDefaultWithOptions(
		cfg.Outbox.Outbox,
		outboxclient.WithSqlxDB(pg.DB),
		outboxclient.WithKafkaProducer(kafkaProducer),
	)
	if err != nil {
		return err
	}

	outboxCmp := outboxComponent.New(cfg.Outbox, outboxCl)
	repo := repository.New(pg.DB)
	provider := trongrid.New(cfg.Tron, cfg.Processor.Network, cfg.Processor.Asset)
	service := watcherservice.New(repo, provider, outboxCmp, cfg.Processor)
	processorCmp := processorComponent.New(cfg.Processor, service)

	inboxCmp := inboxComponent.New(
		cfg.Inbox,
		pg,
		kafkaClient,
		eventmapper.GetEventMapper(service),
	)

	cmps := build.Components{
		pg,
		outboxCmp,
		inboxCmp,
		processorCmp,
	}

	app, err := build.NewApp(cmps)
	if err != nil {
		return err
	}

	return build.Run(ctx, app)
}
