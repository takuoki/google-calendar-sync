package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"google.golang.org/api/calendar/v3"
)

func init() {
	registerCommand(eventCommand())
}

// eventCommand returns the event subcommand.
func eventCommand() *cli.Command {
	return &cli.Command{
		Name:  "event",
		Usage: "Get a specific event by calendar ID and event ID.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "calendar-id",
				Usage:    "Google Calendar ID.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "event-id",
				Usage:    "Event ID.",
				Required: true,
			},
		},
		Action: run(eventAction),
	}
}

func eventAction(c *cli.Context, srv *calendar.Service) (any, error) {
	ctx := c.Context

	calendarID := c.String("calendar-id")
	eventID := c.String("event-id")

	event, err := GetEvent(ctx, srv, calendarID, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return event, nil
}
