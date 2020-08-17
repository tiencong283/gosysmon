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

func (conn *DBConn) Close() error {
	for _, smt := range conn.preparedSmts {
		smt.Close()
	}
	return conn.db.Close()
}

// DeleteAll deletes all entries in related tables
func (conn *DBConn) DeleteAll() error {
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
func (conn *DBConn) SaveHost(hostId string, host *Host) error {
	query := "INSERT INTO Hosts(HostId, Name, FirstSeen, Active) VALUES($1, $2, $3, $4)"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	if _, err = stmt.Exec(hostId, host.Name, host.FirstSeen, host.Active); err != nil {
		return err
	}
	return nil
}

// UpdateHostState updates the host status
func (conn *DBConn) UpdateHostState(hostId string, active bool) error {
	query := "UPDATE Hosts SET Active=$1 WHERE HostId=$2"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(active, hostId); err != nil {
		return err
	}
	return nil
}

// SaveProc inserts new process into Processes table
func (conn *DBConn) SaveProc(hostId string, proc *Process) error {
	query := `INSERT INTO Processes(HostId, ProcessGuid, CreatedAt, TerminatedAt, State, ProcessId, Image, Marshal, PProcessGuid)
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
	if _, err = stmt.Exec(hostId, proc.ProcessGuid, proc.CreatedAt, proc.TerminatedAt, proc.State, proc.ProcessId, proc.Image,
		json, ppGuid); err != nil {
		return err
	}
	return nil
}

// UpdateProc updates the process in Processes table
func (conn *DBConn) UpdateProc(hostId string, proc *Process) error {
	query := `UPDATE Processes SET 
				CreatedAt=$1,
				TerminatedAt=$2,
				State=$3,
				Marshal=$4,
				PProcessGuid=$5
				WHERE HostId=$6 and ProcessGuid=$7`
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	json, _ := json.Marshal(proc)
	ppGuid := ""
	if proc.Parent != nil {
		ppGuid = proc.Parent.ProcessGuid
	}
	if _, err = stmt.Exec(proc.CreatedAt, proc.TerminatedAt, proc.State, json, ppGuid, hostId, proc.ProcessGuid); err != nil {
		return err
	}
	return nil
}

// UpdateProcTerm updates the process to terminated state
func (conn *DBConn) UpdateProcTerm(hostId string, proc *Process) error {
	query := "UPDATE Processes SET State=$1, TerminatedAt=$2 WHERE HostId=$3 and ProcessGuid=$4"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(proc.State, proc.TerminatedAt, hostId, proc.ProcessGuid); err != nil {
		return err
	}
	return nil
}

// DeleteProc delete a process
func (conn *DBConn) DeleteProc(hostId, pGuid string) error {
	query := "DELETE FROM Processes WHERE HostId=$1 and ProcessGuid=$2"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(hostId, pGuid); err != nil {
		return err
	}
	return nil
}

// SaveFeature inserts new feature into Features table
func (conn *DBConn) SaveFeature(fea *MitreATTCKResult) error {
	query := "INSERT INTO Features(HostId, ProcessGuid, Timestamp, IsAlert, Context, Message, TechniqueId) VALUES($1, $2, $3, $4, $5, $6, $7)"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	contextJson, _ := json.Marshal(fea.Context)
	if _, err := stmt.Exec(fea.HostId, fea.ProcessGuid, fea.Timestamp, fea.IsAlert, contextJson, fea.Message, fea.Technique.Id); err != nil {
		return err
	}
	return nil
}

// SaveIOC inserts new ioc into IOCs table
func (conn *DBConn) SaveIOC(ioc *IOCResult) error {
	query := "INSERT INTO IOCs(HostId, ProcessGuid, Timestamp, IOCType, Indicator, Message, ExternalUrl) VALUES($1, $2, $3, $4, $5, $6, $7)"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return err
	}
	if _, err := stmt.Exec(ioc.HostId, ioc.ProcessGuid, ioc.Timestamp, ioc.IOCType, ioc.Indicator, ioc.Message, ioc.ExternalUrl); err != nil {
		return err
	}
	return nil
}

// GetAllHosts returns all hosts
func (conn *DBConn) GetAllHosts() ([]*Host, error) {
	hosts := make([]*Host, 0)
	query := "SELECT HostId, Name, FirstSeen, Active FROM Hosts ORDER BY Id"
	rows, err := conn.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		host := &Host{
			Procs: make(map[string]*Process, 10000),
		}
		if err := rows.Scan(&host.HostId, &host.Name, &host.FirstSeen, &host.Active); err != nil {
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
func (conn *DBConn) GetProcessesByHost(hostId string) ([]*Process, error) {
	procs := make([]*Process, 0)
	query := `SELECT ProcessGuid, CreatedAt, TerminatedAt, State, ProcessId, Image, Marshal, PProcessGuid
				FROM Processes WHERE HostId=$1 ORDER BY Id`
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(hostId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var marshal string
		proc := NewProcess()
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
func (conn *DBConn) GetFeaturesByProcess(hostId, processGuid string) ([]*MitreATTCKResult, error) {
	features := make([]*MitreATTCKResult, 0)
	query := "SELECT Timestamp, IsAlert, Context, Message, TechniqueId FROM Features WHERE HostId=$1 and ProcessGuid=$2 ORDER BY Id"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(hostId, processGuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var context, techID string
		fea := &MitreATTCKResult{
			ResultId: ResultId{
				HostId:      hostId,
				ProcessGuid: processGuid,
			},
		}
		if err := rows.Scan(&fea.Timestamp, &fea.IsAlert, &context, &fea.Message, &techID); err != nil {
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
	query := "SELECT HostId, ProcessGuid, Timestamp, IOCType, Indicator, Message, ExternalUrl FROM IOCs ORDER BY Id"
	rows, err := conn.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		ioc := new(IOCResult)
		if err := rows.Scan(&ioc.HostId, &ioc.ProcessGuid, &ioc.Timestamp, &ioc.IOCType, &ioc.Indicator, &ioc.Message, &ioc.ExternalUrl); err != nil {
			return nil, err
		}
		iocs = append(iocs, ioc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return iocs, nil
}

func (conn *DBConn) GetAlertsOrderByTimestampDesc() ([]*MitreATTCKResult, error) {
	alerts := make([]*MitreATTCKResult, 0)
	query := "SELECT HostId, ProcessGuid, Timestamp, IsAlert, Context, Message, " +
		"TechniqueId FROM Features WHERE IsAlert ORDER BY Timestamp DESC"
	rows, err := conn.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var context, techID string
		fea := new(MitreATTCKResult)
		if err := rows.Scan(&fea.HostId, &fea.ProcessGuid, &fea.Timestamp, &fea.IsAlert, &context, &fea.Message,
			&techID); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(context), &fea.Context); err != nil {
			return nil, err
		}
		fea.Technique = Techniques[techID]
		alerts = append(alerts, fea)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return alerts, nil
}

func (conn *DBConn) GetFeaturesByProc(hostId, processGuid string) ([]*MitreATTCKResult, error) {
	alerts := make([]*MitreATTCKResult, 0)
	query := "SELECT Timestamp, IsAlert, Context, Message, " +
		"TechniqueId FROM Features WHERE IsAlert and HostId=$1 and ProcessGuid=$2 ORDER BY Timestamp"
	stmt, err := conn.GetOrPreparedSmt(query)
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(hostId, processGuid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var context, techID string
		fea := new(MitreATTCKResult)
		if err := rows.Scan(&fea.Timestamp, &fea.IsAlert, &context, &fea.Message, &techID); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(context), &fea.Context); err != nil {
			return nil, err
		}
		fea.HostId = hostId
		fea.ProcessGuid = processGuid
		fea.Technique = Techniques[techID]
		alerts = append(alerts, fea)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return alerts, nil
}

func (conn *DBConn) GetTechniqueStats() (*TechniqueStats, error) {
	result := &TechniqueStats{
		Counts: make([]TechniqueCount, 0),
	}
	query := "SELECT TechniqueId, count(TechniqueId) FROM Features GROUP BY TechniqueId"
	rows, err := conn.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var techID string
		var techCount TechniqueCount
		if err := rows.Scan(&techID, &techCount.Count); err != nil {
			return nil, err
		}
		techCount.Technique = Techniques[techID]
		result.Counts = append(result.Counts, techCount)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
