package catalog

import "embed"

//go:embed *.sql
var MysqlMigrations embed.FS
