-- +goose Up
-- +goose StatementBegin
ALTER TABLE students ADD COLUMN is_payed BOOLEAN DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE students DROP COLUMN is_payed;
-- +goose StatementEnd
