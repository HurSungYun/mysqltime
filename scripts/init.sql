-- Initialize the database schema and data

CREATE TABLE IF NOT EXISTS work_hour (
    user_id BIGINT NOT NULL,
    work_hour_start TIME,
    work_hour_end TIME
);

INSERT INTO work_hour (user_id, work_hour_start, work_hour_end)
VALUES
(1, '09:00:00', '17:00:00'),
(2, '08:30:00', '16:30:00'),
(3, '-05:15:30', '06:45:15'); -- Including negative time example
