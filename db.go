package main

import (
	"database/sql"
	"github.com/gomodule/redigo/redis"
)

// redis
var RedisConn redis.Conn

// postgresql
type DBConn struct {
	db           *sql.DB
	preparedSmts map[string]*sql.Stmt
}

var PgConn *DBConn

func InitRedis(redisConUrl string) error {
	var err error
	RedisConn, err = redis.DialURL(redisConUrl)
	if err != nil {
		return err
	}
	return nil
}

func InitPg(driverName, connUrl string) error {
	db, err := sql.Open(driverName, connUrl)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	PgConn = &DBConn{
		db:           db,
		preparedSmts: make(map[string]*sql.Stmt),
	}
	return nil
}

func (conn *DBConn) Close() {
	for _, smt := range conn.preparedSmts {
		_ = smt.Close()
	}
	_ = conn.db.Close()
}

// DeleteAll deletes all entries in related tables
func (conn *DBConn) DeleteAll() error {
	if _, err := conn.db.Exec("DELETE FROM KafkaOffsets"); err != nil {
		return err
	}
	if _, err := conn.db.Exec("DELETE FROM IOCs"); err != nil {
		return err
	}
	if _, err := conn.db.Exec("DELETE FROM Features"); err != nil {
		return err
	}
	if _, err := conn.db.Exec("DELETE FROM Processes"); err != nil {
		return err
	}
	if _, err := conn.db.Exec("DELETE FROM Hosts"); err != nil {
		return err
	}
	return nil
}

func (conn *DBConn) GetOrPreparedSmt(query string) (*sql.Stmt, error) {
	stmt, ok := conn.preparedSmts[query]
	if !ok {
		stmt, err := conn.db.Prepare(query)
		if err != nil {
			return nil, err
		}
		conn.preparedSmts[query] = stmt
		return stmt, nil
	}
	return stmt, nil
}

// SaveHost inserts new host into IOCs table
func (conn *DBConn) SaveHost(providerGuid string, host *Host) error {
	query := "INSERT INTO Hosts(ProviderGuid, Name, FirstSeen, Active) VALUES($1, $2, $3, $4)"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	if _, err = stmt.Exec(providerGuid, host.Name, host.FirstSeen, host.Active); err != nil {
		return err
	}
	return nil
}

// SaveProc inserts new process into Processes table
func (conn *DBConn) SaveProc(providerGuid string, proc *Process) error {
	query := `INSERT INTO Processes(ProviderGuid, ProcessGuid, CreatedAt, TerminatedAt, State, ProcessId, Image, Marshal, PProcessGuid)
			  	VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	json, _ := json.Marshal(proc)
	ppGuid := ""
	if proc.Parent != nil {
		ppGuid = proc.Parent.ProcessGuid
	}
	if _, err = stmt.Exec(providerGuid, proc.ProcessGuid, proc.CreatedAt, proc.TerminatedAt, proc.State, proc.ProcessId, proc.Image,
		json, ppGuid); err != nil {
		return err
	}
	return nil
}

// UpdateProcTerm updates the process to terminated state
func (conn *DBConn) UpdateProcTerm(providerGuid string, proc *Process) error {
	query := "UPDATE Processes SET State=$1, TerminatedAt=$2 WHERE ProviderGuid=$3 and ProcessGuid=$4"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(proc.State, proc.TerminatedAt, providerGuid, proc.ProcessGuid); err != nil {
		return err
	}
	return nil
}

// SaveFeature inserts new feature into Features table
func (conn *DBConn) SaveFeature(fea *MitreATTCKResult) error {
	query := "INSERT INTO Features(ProviderGuid, ProcessGuid, IsAlert, Context, Message, TechniqueId) VALUES($1, $2, $3, $4, $5, $6)"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	contextJson, _ := json.Marshal(fea.Context)
	if _, err := stmt.Exec(fea.ProviderGUID, fea.ProcessGuid, fea.IsAlert, contextJson, fea.Message, fea.Technique.Id); err != nil {
		return err
	}
	return nil
}

// SaveIOC inserts new ioc into IOCs table
func (conn *DBConn) SaveIOC(ioc *IOCResult) error {
	query := "INSERT INTO IOCs(ProviderGuid, ProcessGuid, IOCType, Indicator, Message, ExternalUrl) VALUES($1, $2, $3, $4, $5, $6)"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(ioc.ProviderGUID, ioc.ProcessGuid, ioc.IOCType, ioc.Indicator, ioc.Message, ioc.ExternalUrl); err != nil {
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

// GetAllHosts returns all hosts
func (conn *DBConn) GetAllHosts() ([]*Host, error) {
	hosts := make([]*Host, 0)
	query := "SELECT ProviderGuid,Name,FirstSeen,Active FROM Hosts ORDER BY Id"
	rows, err := conn.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		host := &Host{
			Procs: make(map[string]*Process, 10000),
		}
		if err := rows.Scan(&host.ProviderGuid, &host.Name, &host.FirstSeen, &host.Active); err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hosts, nil
}

// GetProcessesByHost returns all processes by host
func (conn *DBConn) GetProcessesByHost(providerGuid string) ([]*Process, error) {
	procs := make([]*Process, 0)
	query := `SELECT ProcessGuid, CreatedAt, TerminatedAt, State, ProcessId, Image, Marshal, PProcessGuid
				FROM Processes WHERE ProviderGuid=$1 ORDER BY Id`
	rows, err := conn.db.Query(query, providerGuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var marshal string
		proc := &Process{
			Features: make([]*MitreATTCKResult, 0, 32),
		}
		if err := rows.Scan(&proc.ProcessGuid, &proc.CreatedAt, &proc.TerminatedAt, &proc.State, &proc.ProcessId,
			&proc.Image, &marshal, &proc.ParentPGuid); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(marshal), proc); err != nil {
			return nil, err
		}
		procs = append(procs, proc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return procs, nil
}

// GetFeaturesByProcess returns all features by process
func (conn *DBConn) GetFeaturesByProcess(providerGuid, processGuid string) ([]*MitreATTCKResult, error) {
	features := make([]*MitreATTCKResult, 0)
	query := "SELECT IsAlert, Context, Message, TechniqueId FROM Features WHERE ProviderGuid=$1 and ProcessGuid=$2 ORDER BY Id"
	rows, err := conn.db.Query(query, providerGuid, processGuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var context, techID string
		fea := &MitreATTCKResult{
			ResultId: ResultId{
				ProviderGUID: providerGuid,
				ProcessGuid:  processGuid,
			},
		}
		if err := rows.Scan(&fea.IsAlert, &context, &fea.Message, &techID); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(context), &fea.Context); err != nil {
			return nil, err
		}
		fea.Technique = Techniques[techID]
		features = append(features, fea)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return features, nil
}

// GetAllIOCs returns all IOCs
func (conn *DBConn) GetAllIOCs() ([]*IOCResult, error) {
	iocs := make([]*IOCResult, 0)
	query := "SELECT ProviderGuid, ProcessGuid, IOCType, Indicator, Message, ExternalUrl FROM IOCs ORDER BY Id"
	rows, err := conn.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		ioc := new(IOCResult)
		if err := rows.Scan(&ioc.ProviderGUID, &ioc.ProcessGuid, &ioc.IOCType, &ioc.Indicator, &ioc.Message, &ioc.ExternalUrl); err != nil {
			return nil, err
		}
		iocs = append(iocs, ioc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return iocs, nil
}
