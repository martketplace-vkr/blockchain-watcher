create schema if not exists blockchain_watcher;

create table if not exists blockchain_watcher.watch_address (
    id bigserial primary key,
    user_id bigint not null,
    address text not null,
    network text not null,
    asset text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (network, address)
);

create index if not exists idx_blockchain_watcher_watch_address_user_id
    on blockchain_watcher.watch_address(user_id);

create table if not exists blockchain_watcher.deposit (
    id bigserial primary key,
    user_id bigint not null,
    address text not null,
    network text not null,
    asset text not null,
    tx_hash text not null,
    log_index bigint not null,
    amount numeric(24, 8) not null check (amount > 0),
    block_number bigint not null,
    confirmations bigint not null default 0,
    status text not null,
    confirmed_at timestamptz,
    outbox_sent_at timestamptz,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    unique (network, tx_hash, log_index)
);

create index if not exists idx_blockchain_watcher_deposit_user_id
    on blockchain_watcher.deposit(user_id);

create index if not exists idx_blockchain_watcher_deposit_status
    on blockchain_watcher.deposit(status);

create table if not exists blockchain_watcher.cursor (
    network text primary key,
    last_block bigint not null,
    last_tx_at timestamptz not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
