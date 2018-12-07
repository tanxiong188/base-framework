// Copyright 2018 cloudy 272685110@qq.com.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package models

import (
	"errors"
	"fmt"
	"github.com/itcloudy/base-framework/pkg/conf"
	"github.com/itcloudy/base-framework/pkg/consts"
	"github.com/itcloudy/base-framework/pkg/logs"
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var (
	// DBConn is orm connection
	DBConn *gorm.DB
	SqlxDB *sqlx.DB

	// ErrRecordNotFound is Not Found Record wrapper
	ErrRecordNotFound = errors.New("Record Not Found ")

	// ErrDBConn database connection error
	ErrDBConn = errors.New("Database connection error ")
)

func GetDBConnectionString(dbType string, host string, port int, user string, pass string, dbName string, charset string) (str string) {

	if dbType == "postgres" {
		str = fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable password=%s", host, port, user, dbName, pass)
	} else if dbType == "mysql" {

		str = fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=%s&parseTime=True&loc=Local", user, pass, host, port, dbName, charset)
	}
	return
}

// GetDBConnection is initializes sqlx connection
func GetDBConnection(dbType string, host string, port int, user string, pass string, dbName string, charset string, action string) error {
	var err error
	connStr := GetDBConnectionString(dbType, host, port, user, pass, dbName, charset)
	if action != "" {
		SqlxDB, err = sqlx.Open(dbType, connStr)
	} else {
		DBConn, err = gorm.Open(dbType, connStr)
	}

	if err != nil {
		logs.Logger.Error("can't open connection to DB", zap.String("type", consts.DBError), zap.Error(err))
		DBConn = nil

		return err
	}
	return nil
}

// DropTables is dropping all of the tables
func DropTables() (err error) {
	dbType := conf.Config.DB.DbType
	db := SqlxDB
	if dbType == "postgres" {

		_, err = db.Exec(`
		DO $$ DECLARE
	    	r RECORD;
		BEGIN
	    	FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = current_schema()) LOOP
			EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
	    	END LOOP;
		END $$;
		`)
		return err
	} else if dbType == "mysql" {
		// delete foreign key
		db.Exec(`DROP  PROCEDURE IF  EXISTS procedure_drop_foreign_key;`)
		_, er := db.Exec(fmt.Sprintf(`
CREATE PROCEDURE procedure_drop_foreign_key()
BEGIN
  DECLARE DB_NAME varchar(50) DEFAULT "%s"; 
  DECLARE done INT DEFAULT 0;
  DECLARE tableName varchar(50);   
  DECLARE constraintName varchar(50);   
  DECLARE cmd varchar(450);         
  DECLARE sur CURSOR               
  FOR 
  
  SELECT   TABLE_NAME , CONSTRAINT_NAME 
  FROM information_schema.key_column_usage 
  WHERE CONSTRAINT_SCHEMA = DB_NAME 
  AND referenced_table_name IS NOT NULL;
  DECLARE CONTINUE HANDLER FOR SQLSTATE '02000' SET done = 1;
 
  OPEN sur;
  REPEAT
    FETCH sur INTO tableName,constraintName;
    IF NOT done THEN 
		set cmd=concat('ALTER TABLE ', tableName, ' DROP FOREIGN KEY ', constraintName);
        SET @E=cmd; 
        PREPARE stmt FROM @E; 
          EXECUTE stmt;  
        DEALLOCATE PREPARE stmt;  
    END IF;
  UNTIL done END REPEAT;
  CLOSE sur;
END;`, conf.Config.DB.Name))
		if er != nil {
			return er
		}
		_, err = db.Exec(`call procedure_drop_foreign_key();`)
		if err != nil {
			return err
		}
		// drop all tables
		db.Exec(`DROP  PROCEDURE IF  EXISTS procedure_drop_table;`)
		db.Exec(fmt.Sprintf(`
CREATE PROCEDURE procedure_drop_table()
BEGIN
  DECLARE DB_NAME varchar(50) DEFAULT "%s"; 
  DECLARE done INT DEFAULT 0;
  DECLARE tableName varchar(50);   
  DECLARE cmd varchar(50);         
  DECLARE sur CURSOR               
  FOR 
  SELECT table_name FROM information_schema.TABLES WHERE table_schema=DB_NAME; 
  DECLARE CONTINUE HANDLER FOR SQLSTATE '02000' SET done = 1;
 
  OPEN sur;
  REPEAT
    FETCH sur INTO tableName;
    IF NOT done THEN 
       set cmd=concat('DROP TABLE ',DB_NAME,'.',tableName);    
        SET @E=cmd; 
        PREPARE stmt FROM @E; 
          EXECUTE stmt;  
         DEALLOCATE PREPARE stmt;  
    END IF;
  UNTIL done END REPEAT;
  CLOSE sur;
END;

`, conf.Config.DB.Name))
		_, err = db.Exec(`call procedure_drop_table();`)
		return err
	} else {
		return errors.New("db type not support")
	}
}