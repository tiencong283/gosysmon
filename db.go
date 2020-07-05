package main

import (
	"database/sql"
)

type DBConn struct {
	db           *sql.DB
	preparedSmts map[string]*sql.Stmt
}

func NewDBConn(driverName, connUrl string) (*DBConn, error) {
	db, err := sql.Open(driverName, connUrl)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &DBConn{
		db:           db,
		preparedSmts: make(map[string]*sql.Stmt),
	}, nil
}

func (conn *DBConn) Close() {
	_ = conn.db.Close()
}

// DeleteAll deletes all entries in related tables
func (conn *DBConn) DeleteAll() error {
	if _, err := conn.db.Exec("DELETE FROM Processes"); err != nil {
		return err
	}
	if _, err := conn.db.Exec("DELETE FROM Hosts"); err != nil {
		return err
	}
	if _, err := conn.db.Exec("DELETE FROM KafkaOffsets"); err != nil {
		return err
	}
	return nil
}

// SaveHost inserts new host into Hosts table
func (conn *DBConn) SaveHost(host *Host) error {
	query := "INSERT INTO Hosts(Name, FirstSeen, Active) VALUES($1, $2, $3)"
	stmt, ok := conn.preparedSmts[query]
	if !ok {
		var err error
		stmt, err = conn.db.Prepare(query)
		if err != nil {
			return err
		}
		conn.preparedSmts[query] = stmt
	}
	_, err := stmt.Exec(host.Name, host.FirstSeen, host.Active)
	if err != nil {
		return err
	}
	return nil
}

// SaveProc inserts new process into Processes table
func (conn *DBConn) SaveProc(host *Host, proc *Process) error {
	query := `INSERT INTO Processes(HostName, ProcessGuid, CreatedAt, TerminatedAt, State, ProcessId, Image, Marshal, PProcessGuid)
			  	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	stmt, ok := conn.preparedSmts[query]
	if !ok {
		var err error
		stmt, err = conn.db.Prepare(query)
		if err != nil {
			return err
		}
		conn.preparedSmts[query] = stmt
	}
	json, _ := json.Marshal(proc)
	ppGuid := ""
	if proc.Parent != nil {
		ppGuid = proc.Parent.ProcessGuid
	}
	_, err := stmt.Exec(host.Name, proc.ProcessGuid, proc.CreatedAt, proc.TerminatedAt, proc.State, proc.ProcessId, proc.Image,
		json, ppGuid)
	if err != nil {
		return err
	}
	return nil
}

// UpdateProcTerm updates the process to terminated state
func (conn *DBConn) UpdateProcTerm(host *Host, proc *Process) error {
	query := "UPDATE Processes SET State=$1, TerminatedAt=$2 WHERE Hostname=$3 and ProcessGuid=$4"
	stmt, ok := conn.preparedSmts[query]
	if !ok {
		var err error
		stmt, err = conn.db.Prepare(query)
		if err != nil {
			return err
		}
		conn.preparedSmts[query] = stmt
	}
	_, err := stmt.Exec(proc.State, proc.TerminatedAt, host.Name, proc.ProcessGuid)
	if err != nil {
		return err
	}
	return nil
}

// GetPreKafkaOffset returns the latest kafka broker offset
func (conn *DBConn) GetPreKafkaOffset() int64 {
	var offset int64
	query := "SELECT KafkaOffset FROM KafkaOffsets ORDER BY ModTime DESC LIMIT 1"
	err := conn.db.QueryRow(query).Scan(&offset)
	if err != nil {
		return 0
	}
	return offset
}

// SaveKafkaOffset insert one entry into KafkaOffsets table
func (conn *DBConn) SaveKafkaOffset(offset int64) error {
	query := "INSERT INTO KafkaOffsets(KafkaOffset) VALUES($1)"
	_, err := conn.db.Exec(query, offset)
	if err != nil {
		return err
	}
	return nil
}

