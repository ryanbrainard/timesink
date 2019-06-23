create table cloud_events
(
    time timestamp with time zone not null,
    id text,
    type text,
    source text,
    subject text,
    raw jsonb not null
);

SELECT create_hypertable('cloud_events', 'time');
