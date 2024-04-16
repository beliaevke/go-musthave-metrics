-- +goose Up
CREATE TABLE Gauges (
    mname TEXT PRIMARY KEY,
    mvalue DOUBLE PRECISION NOT NULL
);

CREATE TABLE Counters (
    mname TEXT PRIMARY KEY,
    mvalue INTEGER NOT NULL
);

-- +goose Down
DROP TABLE Gauges;
DROP TABLE Counters;