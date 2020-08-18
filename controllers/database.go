package controllers

import (
	"bytes"
	cyndiv1beta1 "cyndi-operator/api/v1beta1"
	"fmt"
	"github.com/jackc/pgx"
	"strconv"
	"text/template"
)

func checkIfTableExists(tableName string, db *pgx.Conn) (bool, error) {
	if tableName == "" {
		return false, nil
	}

	query := fmt.Sprintf(
		"SELECT exists (SELECT FROM information_schema.tables WHERE table_schema = 'inventory' AND table_name = '%s')",
		tableName)
	rows, err := db.Query(query)

	var exists bool
	rows.Next()
	err = rows.Scan(&exists)
	if err != nil {
		return false, err
	}
	rows.Close()

	return exists, err
}

func deleteTable(tableName string, db *pgx.Conn) error {
	tableExists, err := checkIfTableExists(tableName, db)
	if err != nil {
		return err
	} else if tableExists != true {
		return nil
	}

	query := fmt.Sprintf(
		"DROP table inventory.%s CASCADE", tableName)
	_, err = db.Query(query)
	return err
}

func createTable(tableName string, db *pgx.Conn, dbSchema string) error {
	m := make(map[string]string)
	m["TableName"] = tableName
	tmpl, err := template.New("dbSchema").Parse(dbSchema)
	if err != nil {
		return err
	}

	var dbSchemaBuffer bytes.Buffer
	err = tmpl.Execute(&dbSchemaBuffer, m)
	if err != nil {
		return err
	}
	dbSchemaParsed := dbSchemaBuffer.String()
	_, err = db.Exec(dbSchemaParsed)
	return err
}

func updateView(instance *cyndiv1beta1.CyndiPipeline, db *pgx.Conn) error {
	_, err := db.Exec(fmt.Sprintf(`CREATE OR REPLACE view inventory.hosts as select * from inventory.%s`, instance.Status.TableName))
	return err
}

func connectToInventoryDB(instance *cyndiv1beta1.CyndiPipeline) (*pgx.Conn, error) {
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=%s port=%s",
		instance.Spec.InventoryDBHostname,
		instance.Spec.InventoryDBUser,
		instance.Spec.InventoryDBPassword,
		instance.Spec.InventoryDBName,
		instance.Spec.InventoryDBSSLMode,
		strconv.FormatInt(instance.Spec.InventoryDBPort, 10))
	config, err := pgx.ParseDSN(connStr)
	db, err := pgx.Connect(config)
	return db, err
}

func connectToAppDB(instance *cyndiv1beta1.CyndiPipeline) (*pgx.Conn, error) {
	connStr := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=%s port=%s",
		instance.Spec.AppDBHostname,
		instance.Spec.AppDBUser,
		instance.Spec.AppDBPassword,
		instance.Spec.AppDBName,
		instance.Spec.AppDBSSLMode,
		strconv.FormatInt(instance.Spec.AppDBPort, 10))
	config, err := pgx.ParseDSN(connStr)
	db, err := pgx.Connect(config)
	return db, err
}
