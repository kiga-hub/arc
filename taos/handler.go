package taos

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	_ "github.com/taosdata/driver-go/v2/taosSql" // taosSql -
	"github.com/taosdata/driver-go/v2/types"
)

// IHandler -
type IHandler interface {
	Name() string
	Ping() error
	Exec(s string) (int64, error)
	Query(s string, args ...interface{}) ([]map[string]interface{}, error)
	Close() error
}

// Handler taoclient struct
type Handler struct {
	db     *sql.DB
	dbname string
}

// NewTaoClient Open  taosql
func NewTaoClient(config *Config) (IHandler, error) {
	var c Handler
	url := fmt.Sprintf("%s:%s@/tcp(%s:%d)/",
		config.User,
		config.Password,
		config.Host,
		config.Port)
	db, err := sql.Open(config.Device, url)
	if err != nil {
		return nil, err
	}
	c.db = db
	c.dbname = config.Name
	return &c, err
}

// Name -
func (c *Handler) Name() string {
	return c.dbname
}

// DB -
func (c *Handler) DB() *sql.DB {
	return c.db
}

// Ping -
func (c *Handler) Ping() error {
	return c.db.Ping()
}

// Exec -
func (c *Handler) Exec(s string) (int64, error) {
	res, err := c.db.Exec(s)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Query -
func (c *Handler) Query(s string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := c.db.Query(s, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var row []interface{}
	tps, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	for _, tp := range tps {
		switch tp.ScanType() {
		case reflect.TypeOf(types.NullTime{}):
			row = append(row, &time.Time{})
		case reflect.TypeOf(types.NullFloat32{}):
			row = append(row, new(float32))
		case reflect.TypeOf(types.NullFloat64{}):
			row = append(row, new(float64))
		case reflect.TypeOf(types.NullInt32{}):
			row = append(row, new(int32))
		case reflect.TypeOf(types.NullInt64{}):
			row = append(row, new(int64))
		case reflect.TypeOf(types.NullString{}):
			row = append(row, new(string))
		default:
			row = append(row, new(interface{}))
		}
	}
	var res []map[string]interface{}
	for rows.Next() {
		if err := rows.Scan(row...); err != nil {
			return nil, err
		}
		one := make(map[string]interface{})
		for i, tp := range tps {
			var v interface{}
			switch row[i].(type) {
			case *time.Time:
				v = *(row[i].(*time.Time))
			case *float32:
				v = *(row[i].(*float32))
			case *float64:
				v = *(row[i].(*float64))
			case *int32:
				v = *(row[i].(*int32))
			case *int64:
				v = *(row[i].(*int64))
			default:
				v = row[i]
			}
			one[tp.Name()] = v
		}
		res = append(res, one)
	}
	return res, nil
}

// Close close db
func (c *Handler) Close() error {
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}
