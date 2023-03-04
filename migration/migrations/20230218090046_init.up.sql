CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE entity
(
    id     uuid          NOT NULL DEFAULT uuid_generate_v4(),
    secret bytea         NOT NULL,
    attrs  varchar array NOT NULL,
    note   varchar,

    PRIMARY KEY (id)
);
