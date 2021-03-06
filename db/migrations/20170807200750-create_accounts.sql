
-- +migrate Up
create table accounts (
  id serial not null primary key,
  name text not null,
  users_id integer not null references users (id),
  created timestamp not null default current_timestamp,
  uid UUID not null UNIQUE
);

-- +migrate Down
DROP TABLE accounts;
