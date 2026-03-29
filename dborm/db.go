package dborm

import (
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/rehiy/pango/logman"
)

var Db *gorm.DB
var RawLogger *logman.Logger

type Config struct {
	Type     string `note:"数据库类型"`
	Host     string `note:"主机地址"`
	User     string `note:"用户名"`
	Password string `note:"用户密码"`
	DbName   string `note:"数据库名称"`
	Option   string `note:"数据库选项"`
}

func Connect(args *Config) *gorm.DB {
	if RawLogger == nil {
		RawLogger = logman.Named("gorm")
		logger.Default = NewLogger(RawLogger)
	}

	config := &gorm.Config{
		Logger: logger.Default,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}

	if db, err := gorm.Open(dialector(args), config); err != nil {
		logman.Fatal("connect to databse failed", "error", err)
	} else {
		Db = db
	}

	return Db
}

func Destroy() error {
	if db, err := Db.DB(); db != nil {
		return db.Close()
	} else {
		return err
	}
}

func dialector(args *Config) gorm.Dialector {
	switch args.Type {
	case "sqlite":
		return useSqlite(args)
	case "mysql":
		return useMysql(args)
	case "pgsql", "postgres", "postgresql":
		return usePgsql(args)
	default:
		logman.Fatal("database type error", "type", args.Type)
	}

	return nil
}
