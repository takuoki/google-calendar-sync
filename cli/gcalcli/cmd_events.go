package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/urfave/cli/v2"
	"google.golang.org/api/calendar/v3"
)

func init() {
	registerCommand(eventsCommand())
}

// eventsCommand returns the events subcommand.
func eventsCommand() *cli.Command {
	return &cli.Command{
		Name:  "events",
		Usage: "Get a list of events from a calendar by calendar ID.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "calendar-id",
				Usage:    "Google Calendar ID.",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "page-token",
				Usage: "Token specifying which result page to return.",
			},
			&cli.StringFlag{
				Name:  "sync-token",
				Usage: "Token obtained from the nextSyncToken field returned on the last page of results from the previous list request.",
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
		Action: run(eventsAction),
	}
}

func eventsAction(c *cli.Context, srv *calendar.Service) (any, error) {
	ctx := c.Context

	calendarID := c.String("calendar-id")
	pageToken := c.String("page-token")
	syncToken := c.String("sync-token")
	timeMax := c.String("time-max")
	timeMin := c.String("time-min")

	// Validation: SyncToken cannot be used with TimeMax or TimeMin
	if syncToken != "" && (timeMax != "" || timeMin != "") {
		return nil, errors.New("sync-token cannot be used with time-max or time-min")
	}

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

	events, err := ListEvents(ctx, srv, calendarID, pageToken, syncToken, timeMax, timeMin)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return events, nil
}
