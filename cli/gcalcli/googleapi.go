package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// CreateCalendarServiceWithOAuth creates a Calendar service using OAuth2 authentication.
func CreateCalendarServiceWithOAuth(ctx context.Context, refreshToken string) (*calendar.Service, error) {
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")
	redirectURL := os.Getenv("OAUTH_REDIRECT_URL")

	if clientID == "" || clientSecret == "" || redirectURL == "" || refreshToken == "" {
		return nil, errors.New("OAUTH_CLIENT_ID, OAUTH_CLIENT_SECRET, OAUTH_REDIRECT_URL env vars and refresh token are required for OAuth2 authentication")
	}

	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{calendar.CalendarScope},
		Endpoint:     google.Endpoint,
	}

	token := &oauth2.Token{RefreshToken: refreshToken}
	httpClient := config.Client(ctx, token)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service (OAuth2): %w", err)
	}

	return srv, nil
}

// CreateCalendarServiceWithServiceAccount creates a Calendar service using service account authentication.
func CreateCalendarServiceWithServiceAccount(ctx context.Context) (*calendar.Service, error) {
	srv, err := calendar.NewService(ctx, option.WithCredentialsFile("./credentials.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service (service account): %w", err)
	}

	return srv, nil
}

// GetEvent gets a calendar event using the provided Calendar service.
func GetEvent(ctx context.Context, srv *calendar.Service, calendarID, eventID string) (*calendar.Event, error) {
	event, err := srv.Events.Get(calendarID, eventID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	return event, nil
}

// ListEvents gets a list of events from a calendar using the provided Calendar service.
func ListEvents(ctx context.Context, srv *calendar.Service, calendarID, pageToken, syncToken, timeMax, timeMin string) (*calendar.Events, error) {
	call := srv.Events.List(calendarID).Context(ctx).ShowDeleted(true)

	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	if syncToken != "" {
		call = call.SyncToken(syncToken)
	}
	if timeMax != "" {
		call = call.TimeMax(timeMax)
	}
	if timeMin != "" {
		call = call.TimeMin(timeMin)
	}

	events, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	return events, nil
}

// ListInstances gets a list of instances of a recurring event using the provided Calendar service.
func ListInstances(ctx context.Context, srv *calendar.Service, calendarID, eventID, pageToken, timeMax, timeMin string) (*calendar.Events, error) {
	call := srv.Events.Instances(calendarID, eventID).Context(ctx).ShowDeleted(true)

	if pageToken != "" {
		call = call.PageToken(pageToken)
	}
	if timeMax != "" {
		call = call.TimeMax(timeMax)
	}
	if timeMin != "" {
		call = call.TimeMin(timeMin)
	}

	events, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	return events, nil
}

// WatchEvents starts watching for changes to events in a calendar using the provided Calendar service.
func WatchEvents(ctx context.Context, srv *calendar.Service, calendarID, channelID, address string) (*calendar.Channel, error) {
	channel := &calendar.Channel{
		Id:      channelID,
		Type:    "web_hook",
		Address: address,
	}

	watchResponse, err := srv.Events.Watch(calendarID, channel).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to start watching events: %w", err)
	}

	return watchResponse, nil
}

// StopChannel stops a push notification channel using the provided Calendar service.
func StopChannel(ctx context.Context, srv *calendar.Service, channelID, resourceID string) error {
	channel := &calendar.Channel{
		Id:         channelID,
		ResourceId: resourceID,
	}

	err := srv.Channels.Stop(channel).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to stop channel: %w", err)
	}

	return nil
}
