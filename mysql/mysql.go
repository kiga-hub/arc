package mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql" //_ Import the required drivers
	"github.com/jinzhu/gorm"

	error2 "github.com/kiga-hub/arc/error"
)

// CreateDB create db
func CreateDB(config Config) (*gorm.DB, error) {
	connection := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=true&loc=%s&timeout=10s", config.User, //&readTimeout=30s
		config.Password, config.Host, config.Port, config.DB, config.TimeZone)
	// fmt.Println(connection)
	db, err := gorm.Open("mysql", connection)
	if err != nil {
		fmt.Println("Mysql connection error", err)
		return nil, err
	}
	db.LogMode(config.LogMode)
	db.DB().SetMaxIdleConns(config.MaxIdleConns)
	db.DB().SetMaxOpenConns(config.MaxOpenConns)
	db.SingularTable(true) // globally set that table names cannot be in plural form
	return db, nil
}

// CheckDB check db
//
//goland:noinspection GoUnusedExportedFunction
func CheckDB(db *gorm.DB, config Config) error {
	_ = config
	if db == nil {
		fmt.Println("DB instance is nil!")
		return error2.ErrDbConnection
	}

	return db.DB().Ping()
}

// DropDatabase drop db
//
//goland:noinspection GoUnusedExportedFunction
func DropDatabase(config Config) error {
	dropNacosConfig := fmt.Sprintf("DROP DATABASE IF EXISTS %s;",
		config.DB)
	mysqlConfig := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.User,
		config.Password,
		config.Host,
		config.Port)
	dh, err := gorm.Open("mysql", mysqlConfig)
	defer func() {
		err = dh.Close()
		if err != nil {
			fmt.Println("dh.Close:", err)
		}
	}()

	if err != nil {
		return err
	}
	fmt.Println("open mysql success")

	dh.Exec(dropNacosConfig)
	fmt.Println("drop database success")

	return nil
}

// CreateDatabase create db
//
//goland:noinspection GoUnusedExportedFunction
func CreateDatabase(config Config) error {
	createNacosConfig := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s CHARACTER SET utf8mb4  COLLATE utf8mb4_general_ci;",
		config.DB)
	mysqlConfig := fmt.Sprintf("%s:%s@tcp(%s:%d)/",
		config.User,
		config.Password,
		config.Host,
		config.Port)
	dh, err := gorm.Open("mysql", mysqlConfig)
	defer func() {
		err = dh.Close()
		if err != nil {
			fmt.Println("dh.Close:", err)
		}
	}()

	if err != nil {
		return err
	}
	fmt.Println("open mysql success")

	dh.Exec(createNacosConfig)
	fmt.Println("create database success")

	return nil
}
