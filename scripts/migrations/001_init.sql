create table if not exists links (
    id bigserial primary key,
    short_code text not null unique,
    long_url text not null,
    redirect_count bigint not null default 0,
    created_at timestamptz not null default now(),
    last_accessed_at timestamptz
);

create index if not exists links_created_at_idx on links(created_at desc);

