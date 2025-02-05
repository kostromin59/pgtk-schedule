-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS students(
  id bigint PRIMARY KEY,
  nickname text,
  stream varchar(255),
  substream varchar(255)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS students;
-- +goose StatementEnd
