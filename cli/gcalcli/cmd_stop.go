package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"google.golang.org/api/calendar/v3"
)

func init() {
	registerCommand(stopCommand())
}

// stopCommand returns the stop subcommand.
func stopCommand() *cli.Command {
	return &cli.Command{
		Name:  "stop",
		Usage: "Stop a push notification channel.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "channel-id",
				Usage:    "Channel ID for the notification channel to stop.",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "resource-id",
				Usage:    "Resource ID for the notification channel to stop.",
				Required: true,
			},
		},
		Action: run(stopAction),
	}
}

func stopAction(c *cli.Context, srv *calendar.Service) (any, error) {
	ctx := c.Context

	channelID := c.String("channel-id")
	resourceID := c.String("resource-id")

	err := StopChannel(ctx, srv, channelID, resourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to stop channel: %w", err)
	}

	return "success", nil
}
