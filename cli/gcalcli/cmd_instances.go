package main

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"google.golang.org/api/calendar/v3"
)

func init() {
	registerCommand(instancesCommand())
}

// instancesCommand returns the instances subcommand.
func instancesCommand() *cli.Command {
	return &cli.Command{
		Name:  "instances",
		Usage: "Get a list of instances of a recurring event by calendar ID and event ID.",
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
			&cli.StringFlag{
				Name:  "page-token",
				Usage: "Token specifying which result page to return.",
			},
			&cli.StringFlag{
				Name:  "time-max",
				Usage: "Upper bound (exclusive) for an event's start time to filter by (RFC3339 timestamp).",
			},
			&cli.StringFlag{
				Name:  "time-min",
				Usage: "Lower bound (inclusive) for an event's end time to filter by (RFC3339 timestamp).",
			},
		},
		Action: run(instancesAction),
	}
}

func instancesAction(c *cli.Context, srv *calendar.Service) (any, error) {
	ctx := c.Context

	calendarID := c.String("calendar-id")
	eventID := c.String("event-id")
	pageToken := c.String("page-token")
	timeMax := c.String("time-max")
	timeMin := c.String("time-min")

	// Validation: RFC3339 format for time parameters
	if timeMax != "" {
		if _, err := time.Parse(time.RFC3339, timeMax); err != nil {
			return nil, fmt.Errorf("time-max must be in RFC3339 format: %w", err)
		}
	}
	if timeMin != "" {
		if _, err := time.Parse(time.RFC3339, timeMin); err != nil {
			return nil, fmt.Errorf("time-min must be in RFC3339 format: %w", err)
		}
	}

	events, err := ListInstances(ctx, srv, calendarID, eventID, pageToken, timeMax, timeMin)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	return events, nil
}
