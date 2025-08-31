package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"google.golang.org/api/calendar/v3"
)

var commands []*cli.Command

// registerCommand adds a command to the global commands slice.
func registerCommand(cmd *cli.Command) {
	commands = append(commands, cmd)
}

func main() {
	app := &cli.App{
		Name:  "gcalcli",
		Usage: "A CLI tool for Google Calendar API operations.",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "oauth",
				Usage: "Use OAuth 2.0 authentication instead of service account.",
			},
			&cli.StringFlag{
				Name:  "refresh-token",
				Usage: "OAuth 2.0 refresh token (required if --oauth is set).",
			},
		},
		Commands: commands,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

// run is a helper function that creates a calendar service and passes it to the provided function.
func run(fn func(*cli.Context, *calendar.Service) (any, error)) cli.ActionFunc {
	return func(c *cli.Context) error {
		srv, err := createCalendarService(c)
		if err != nil {
			return fmt.Errorf("failed to create calendar service: %w", err)
		}

		result, err := fn(c, srv)
		if err != nil {
			return err
		}

		return outputJSON(result)
	}
}

// createCalendarService creates a Calendar service based on the authentication flags.
func createCalendarService(c *cli.Context) (*calendar.Service, error) {
	ctx := c.Context
	useOAuth := c.Bool("oauth")
	refreshToken := c.String("refresh-token")

	if useOAuth {
		return CreateCalendarServiceWithOAuth(ctx, refreshToken)
	}

	return CreateCalendarServiceWithServiceAccount(ctx)
}

// outputJSON outputs the given data as JSON.
func outputJSON(data any) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data to JSON: %w", err)
	}
	fmt.Println(string(jsonData))

	return nil
}
