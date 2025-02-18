-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS notify_settings(
  id bigserial PRIMARY KEY,
  student_id bigint NOT NULL,
  morning bool NOT NULL DEFAULT true,
  evening bool NOT NULL DEFAULT true,
  week bool NOT NULL DEFAULT true,
  FOREIGN KEY(student_id) REFERENCES students(id) 
  ON DELETE CASCADE
  ON UPDATE CASCADE
);

-- Trigger for creating notify_settings on new student
CREATE OR REPLACE FUNCTION create_notify_settings_for_student() RETURNS trigger AS $create_notify_settings$
BEGIN
  INSERT INTO notify_settings(student_id) VALUES (NEW.id);
  RETURN NEW;
END;
$create_notify_settings$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS create_notify_settings_on_new_student ON students;

CREATE TRIGGER create_notify_settings_on_new_student
AFTER INSERT ON students
FOR EACH ROW
EXECUTE FUNCTION create_notify_settings_for_student();

-- Create settings for existing students
INSERT INTO notify_settings (student_id)
SELECT id FROM students;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS create_notify_settings_on_new_student ON students;
DROP FUNCTION IF EXISTS create_notify_settings_for_student;
DROP TABLE IF EXISTS notify_settings;
-- +goose StatementEnd
