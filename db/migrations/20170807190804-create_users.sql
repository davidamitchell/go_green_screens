
-- +migrate Up
create table users (
  id serial not null primary key,
  name text not null,
  created timestamp not null default current_timestamp,
  uid UUID not null UNIQUE
);

-- +migrate Down
DROP TABLE users;
