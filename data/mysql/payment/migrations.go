package payment

import "embed"

//go:embed *.sql
var MysqlMigrations embed.FS
