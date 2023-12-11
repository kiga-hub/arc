package mysql

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql" //_ 导入所需要的驱动
	"github.com/jinzhu/gorm"

	error2 "github.com/kiga-hub/common/error"
)

// CreateDB 创建数据库对象
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
	db.SingularTable(true) // 全局设置表名不可以为复数形式
	return db, nil
}

// CheckDB 检查数据库链接
func CheckDB(db *gorm.DB, config Config) error {
	if db == nil {
		fmt.Println("DB instance is nil!")
		return error2.ErrDbConnection
	}

	return db.DB().Ping()
}

// DropDatabase 删除数据库
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

// CreateDatabase 创建数据库
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
