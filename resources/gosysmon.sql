-- Host table
CREATE TABLE Hosts
(
    Id        SERIAL,
    Name      VARCHAR(64) PRIMARY KEY,
    FirstSeen Timestamp NOT NULL,
    Active    BOOL DEFAULT TRUE
);

-- Process table
CREATE TABLE Processes
(
    Id           SERIAL,
    HostName     VARCHAR(64) REFERENCES Hosts (Name),
    ProcessGuid  VARCHAR(64) PRIMARY KEY,
    CreatedAt    Timestamp NOT NULL,
    TerminatedAt Timestamp,
    State        int       NOT NULL,

    ProcessId    INT       NOT NULL,
    Image        TEXT      NOT NULL,
    Marshal      TEXT      NOT NULL,

-- Marshal is the json representation of these fields below
--     OriginalFileName  TEXT,
--     CommandLine       TEXT,
--     CurrentDirectory  TEXT,
--     IntegrityLevel    TEXT,
--     Hashes            TEXT,
--
--     FileVersion       TEXT,
--     Description       TEXT,
--     Product           TEXT,
--     Company           TEXT,

    PProcessGuid VARCHAR(64)
);

-- KafkaOffset table
CREATE TABLE KafkaOffsets
(
    Id          SERIAL,
    KafkaOffset BIGINT NOT NULL,
    ModTime     TIMESTAMP DEFAULT current_timestamp
);