package repository

import "errors"

// ErrNoDB 数据库未配置
var ErrNoDB = errors.New("database not configured")
