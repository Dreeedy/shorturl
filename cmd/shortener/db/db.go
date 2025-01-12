package db

import "os"

func GetInitScript() (string, error) {
	sqlFile, err := os.ReadFile("db/migrations/init.sql")
	if err != nil {
		return "", err
	}

	return string(sqlFile), nil
}
