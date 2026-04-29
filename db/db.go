package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

type DeviceHistory struct {
	ID          int64     `json:"id"`
	DeviceID    string    `json:"device_id"`
	Group       string    `json:"group"`
	Description string    `json:"description"`
	IPAddr      string    `json:"ip_addr"`
	Proto       uint8     `json:"proto"`
	OnlineTime  time.Time `json:"online_time"`
	OfflineTime *time.Time `json:"offline_time,omitempty"`
	Duration    *int64    `json:"duration,omitempty"` // seconds
}

func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create db directory failed: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("open database failed: %w", err)
	}

	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)

	if err := createTables(); err != nil {
		return fmt.Errorf("create tables failed: %w", err)
	}

	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS device_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		device_id TEXT NOT NULL,
		group_name TEXT DEFAULT '',
		description TEXT DEFAULT '',
		ip_addr TEXT DEFAULT '',
		proto INTEGER DEFAULT 0,
		online_time DATETIME NOT NULL,
		offline_time DATETIME,
		duration INTEGER
	);
	
	CREATE INDEX IF NOT EXISTS idx_device_id ON device_history(device_id);
	CREATE INDEX IF NOT EXISTS idx_online_time ON device_history(online_time);
	CREATE INDEX IF NOT EXISTS idx_group_name ON device_history(group_name);
	`

	_, err := DB.Exec(schema)
	return err
}

func RecordDeviceOnline(devID, group, desc, ipAddr string, proto uint8) (int64, error) {
	result, err := DB.Exec(`
		INSERT INTO device_history (device_id, group_name, description, ip_addr, proto, online_time)
		VALUES (?, ?, ?, ?, ?, ?)
	`, devID, group, desc, ipAddr, proto, time.Now())

	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func RecordDeviceOffline(historyID int64) error {
	now := time.Now()
	
	result, err := DB.Exec(`
		UPDATE device_history 
		SET offline_time = ?, 
		    duration = CAST((julianday(?) - julianday(online_time)) * 86400 AS INTEGER)
		WHERE id = ? AND offline_time IS NULL
	`, now, now, historyID)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("no record found for id %d", historyID)
	}

	return nil
}

func QueryDeviceHistory(deviceID string, limit int) ([]DeviceHistory, error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := DB.Query(`
		SELECT id, device_id, group_name, description, ip_addr, proto, 
		       online_time, offline_time, duration
		FROM device_history
		WHERE device_id = ?
		ORDER BY online_time DESC
		LIMIT ?
	`, deviceID, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []DeviceHistory
	for rows.Next() {
		var h DeviceHistory
		var offlineTime sql.NullTime
		var duration sql.NullInt64

		err := rows.Scan(&h.ID, &h.DeviceID, &h.Group, &h.Description, 
			&h.IPAddr, &h.Proto, &h.OnlineTime, &offlineTime, &duration)
		if err != nil {
			return nil, err
		}

		if offlineTime.Valid {
			h.OfflineTime = &offlineTime.Time
		}
		if duration.Valid {
			h.Duration = &duration.Int64
		}

		histories = append(histories, h)
	}

	return histories, rows.Err()
}

func QueryAllDeviceHistory(group string, limit int) ([]DeviceHistory, error) {
	if limit <= 0 {
		limit = 100
	}

	var rows *sql.Rows
	var err error

	if group == "" {
		rows, err = DB.Query(`
			SELECT id, device_id, group_name, description, ip_addr, proto, 
			       online_time, offline_time, duration
			FROM device_history
			WHERE offline_time IS NULL
			ORDER BY online_time DESC
			LIMIT ?
		`, limit)
	} else {
		rows, err = DB.Query(`
			SELECT id, device_id, group_name, description, ip_addr, proto, 
			       online_time, offline_time, duration
			FROM device_history
			WHERE group_name = ? AND offline_time IS NULL
			ORDER BY online_time DESC
			LIMIT ?
		`, group, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []DeviceHistory
	for rows.Next() {
		var h DeviceHistory
		var offlineTime sql.NullTime
		var duration sql.NullInt64

		err := rows.Scan(&h.ID, &h.DeviceID, &h.Group, &h.Description, 
			&h.IPAddr, &h.Proto, &h.OnlineTime, &offlineTime, &duration)
		if err != nil {
			return nil, err
		}

		if offlineTime.Valid {
			h.OfflineTime = &offlineTime.Time
		}
		if duration.Valid {
			h.Duration = &duration.Int64
		}

		histories = append(histories, h)
	}

	return histories, rows.Err()
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
