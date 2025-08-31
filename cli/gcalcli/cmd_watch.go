package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"google.golang.org/api/calendar/v3"
)

func init() {
	registerCommand(watchCommand())
}

// watchCommand returns the watch subcommand.
func watchCommand() *cli.Command {
	return &cli.Command{
		Name:  "watch",
		Usage: "Start watching for changes to events in a calendar.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "calendar-id",
				Usage:    "Google Calendar ID.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "channel-id",
				Usage:    "Channel ID for the notification channel.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "address",
				Usage:    "Notification endpoint URL.",
				Required: true,
			},
		},
		Action: run(watchAction),
	}
}

func watchAction(c *cli.Context, srv *calendar.Service) (any, error) {
	ctx := c.Context

	calendarID := c.String("calendar-id")
	channelID := c.String("channel-id")
	address := c.String("address")

	channel, err := WatchEvents(ctx, srv, calendarID, channelID, address)
	if err != nil {
		return nil, fmt.Errorf("failed to start watching events: %w", err)
	}

	return channel, nil
}
