package cloudsql

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"cloud.google.com/go/cloudsqlconn"
	mysql "github.com/go-sql-driver/mysql"
)

func ConnectWithConnector(instanceConnectionName, port, user, password, dbname, usePrivate string) (*sql.DB, error) {

	if err := validateArgs(instanceConnectionName, port, user, password, dbname, usePrivate); err != nil {
		return nil, err
	}

	if port == "" {
		port = "3306"
	}

	d, err := cloudsqlconn.NewDialer(context.Background())
	if err != nil {
		return nil, fmt.Errorf("cloudsqlconn.NewDialer: %w", err)
	}
	var opts []cloudsqlconn.DialOption
	if usePrivate != "" {
		opts = append(opts, cloudsqlconn.WithPrivateIP())
	}
	mysql.RegisterDialContext("cloudsqlconn",
		func(ctx context.Context, addr string) (net.Conn, error) {
			return d.Dial(ctx, instanceConnectionName, opts...)
		})

	dbURI := fmt.Sprintf("%s:%s@cloudsqlconn(localhost:%s)/%s?parseTime=true",
		user, password, port, dbname)

	dbPool, err := sql.Open("mysql", dbURI)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	return dbPool, nil
}

func validateArgs(instanceConnectionName, port, user, password, dbname, usePrivate string) error {
	if instanceConnectionName == "" {
		return fmt.Errorf("instance connection name is required")
	}
	if user == "" {
		return fmt.Errorf("database user is required")
	}
	if password == "" {
		return fmt.Errorf("database password is required")
	}
	if dbname == "" {
		return fmt.Errorf("database name is required")
	}
	return nil
}
