package main

import (
	"database/sql"
	"fmt"
	flags "github.com/jessevdk/go-flags"
	_ "github.com/ziutek/mymysql/godrv"
	"log"
	"os"
	"strings"
)

const (
	VERSION = "0.9.0"
)

var appLog = log.New(os.Stdout, "", log.LstdFlags)

var cliFlags struct {
	Help bool   `short:"?" long:"help" description:"Shows this help."`
	Host string `short:"h" long:"host" description:"MySQL host to connect."`
	Port uint16 `short:"P" long:"port" description:"MySQL port to connect."`
	User string `short:"u" long:"user" description:"MySQL user to use."`
	Pass string `short:"p" long:"password" description:"MySQL password to use."`
}

type DatabaseColumn struct {
	TableName  string `db:"table_name"`
	Name       string `db:"column_name"`
	DataType   string `db:"data_type"`
	IsNullable bool   `db:"is_nullable"`
}

func main() {
	cliFlags.Host = "127.0.0.1"
	cliFlags.Port = 3306

	fmt.Printf("mysql-tz-converter v%s.\n", VERSION)
	fmt.Println("Copyright (c) 2013 Philippe Gerber <philippe@bigwhoop.ch>")
	fmt.Println()

	argsParser := flags.NewParser(&cliFlags, flags.PrintErrors|flags.PassDoubleDash)
	argsParser.Usage = "[OPTIONS] DATABASE_NAME FROM_TZ TO_TZ"

	cliArgs, err := argsParser.Parse()
	if err != nil {
		argsParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	if cliFlags.Help {
		argsParser.WriteHelp(os.Stderr)
		os.Exit(0)
	}

	if len(cliArgs) != 3 {
		argsParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	database := cliArgs[0]
	fromTZ := cliArgs[1]
	toTZ := cliArgs[2]

	dsn := fmt.Sprintf(
		"tcp:%s:%d*%s/%s/%s",
		cliFlags.Host, cliFlags.Port, "information_schema", cliFlags.User, cliFlags.Pass,
	)

	db, err := sql.Open("mymysql", dsn)
	if err != nil {
		appLog.Fatalln(err)
		os.Exit(1)
	}

	sql := `
	SELECT
	  table_name, column_name, data_type, IF(is_nullable = "YES", true, false)
	FROM
	  columns
	WHERE
	  data_type IN('datetime', 'date', 'timestamp')
	  AND table_schema = ?
	ORDER BY
	  table_name ASC,
	  column_name ASC
	`

	rows, err := db.Query(sql, database)
	if err != nil {
		appLog.Fatalln(err)
		os.Exit(1)
	}

	columns := make([]*DatabaseColumn, 0)
	columnsByTable := make(map[string][]*DatabaseColumn)

	for rows.Next() {
		column := new(DatabaseColumn)

		scanErr := rows.Scan(&column.TableName, &column.Name, &column.DataType, &column.IsNullable)
		if scanErr != nil {
			panic(scanErr)
		}

		columns = append(columns, column)
		columnsByTable[column.TableName] = append(columnsByTable[column.TableName], column)
	}

	appLog.Printf("Found %d columns in %d tables.\n", len(columns), len(columnsByTable))
	appLog.Printf("Will convert from %s to %s.\n", fromTZ, toTZ)

	for tableName, columns := range columnsByTable {
		columnNames := make([]string, 0)
		columnUpdates := make([]string, 0)

		for _, column := range columns {
			columnNames = append(columnNames, column.Name)
			columnUpdates = append(
				columnUpdates,
				fmt.Sprintf("%s = CONVERT_TZ(%s, '%s', '%s')", column.Name, column.Name, fromTZ, toTZ),
			)
		}

		appLog.Println("---")
		appLog.Printf("Table:  %s\n", tableName)
		appLog.Printf("Fields: %s\n", strings.Join(columnNames, ", "))

		sql = fmt.Sprintf("UPDATE %s.%s SET %s", database, tableName, strings.Join(columnUpdates, ", "))

		updateResult, updateErr := db.Exec(sql)
		if updateErr == nil {
			numRowsUpdated, _ := updateResult.RowsAffected()
			appLog.Printf("DONE. (%d rows)\n", numRowsUpdated)
		} else {
			appLog.Printf("FAILED. (%s)\n", updateErr)
		}
	}

	appLog.Println("---")
}
