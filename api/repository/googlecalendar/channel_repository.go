package googlecalendar

import (
	"context"
	"fmt"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	calendar "google.golang.org/api/calendar/v3"
)

func (r *googleCalendarRepository) StopWatch(ctx context.Context, channel entity.Channel) error {
	return stopWatch(ctx, r.service, channel)
}

func (r *googleCalendarWithOauthRepository) StopWatch(ctx context.Context, channel entity.Channel) error {

	service, err := r.getCalendarService(ctx, channel.CalendarID)
	if err != nil {
		return fmt.Errorf("fail to get calendar service: %w", err)
	}

	return stopWatch(ctx, service, channel)
}

func stopWatch(ctx context.Context, service *calendar.Service, channel entity.Channel) error {
	err := service.Channels.Stop(&calendar.Channel{
		Id:         channel.CalendarID.ToChannelID(),
		ResourceId: string(channel.ResourceID),
	}).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("fail to stop watch: %w", err)
	}

	return nil
}
