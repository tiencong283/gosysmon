-- Host table
CREATE TABLE Hosts
(
    Id           SERIAL,
    ProviderGuid VARCHAR(64) PRIMARY KEY,
    Name         VARCHAR(64) NOT NULL,
    FirstSeen    Timestamp   NOT NULL,
    Active       BOOL DEFAULT TRUE
);

-- PaginationNav table
CREATE TABLE Processes
(
    Id           SERIAL,
    ProviderGuid VARCHAR(64) REFERENCES Hosts (ProviderGuid),
    ProcessGuid  VARCHAR(64) PRIMARY KEY,
    CreatedAt    Timestamp,
    TerminatedAt Timestamp,
    State        int  NOT NULL,

    ProcessId    INT  NOT NULL,
    Image        TEXT NOT NULL,
    Marshal      TEXT NOT NULL,

-- Marshal is the json representation of these fields below
--     Abandoned         BOOL NOT NULL,
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

-- Feature table
CREATE TABLE Features
(
    Id          SERIAL PRIMARY KEY,
    Timestamp   Timestamp,
    ProviderGuid VARCHAR(64) REFERENCES Hosts (ProviderGuid),
    ProcessGuid VARCHAR(64) REFERENCES Processes (ProcessGuid),
    IsAlert     BOOL DEFAULT TRUE,
    Context     TEXT NOT NULL,
    Message     TEXT NOT NULL,
    TechniqueId TEXT NOT NULL
);

-- IOC table
CREATE TABLE IOCs
(
    Id          SERIAL PRIMARY KEY,
    Timestamp   Timestamp,
    ProviderGuid VARCHAR(64) REFERENCES Hosts (ProviderGuid),
    ProcessGuid VARCHAR(64) REFERENCES Processes (ProcessGuid),
    IOCType     INT  NOT NULL,
    Indicator   TEXT NOT NULL,
    Message     TEXT NOT NULL,
    ExternalUrl TEXT NOT NULL
);

-- KafkaOffset table
CREATE TABLE KafkaOffsets
(
    Id          SERIAL,
    KafkaOffset BIGINT NOT NULL,
    ModTime     TIMESTAMP DEFAULT current_timestamp
);