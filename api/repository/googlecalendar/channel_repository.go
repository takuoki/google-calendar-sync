package googlecalendar

import (
	"context"
	"fmt"

	"github.com/takuoki/google-calendar-sync/api/domain/entity"
	calendar "google.golang.org/api/calendar/v3"
)

func (r *googleCalendarRepository) StopWatch(ctx context.Context, channel entity.Channel) error {

	err := r.service.Channels.Stop(&calendar.Channel{
		Id:         channel.CalendarID.ToChannelID(),
		ResourceId: string(channel.ResourceID),
	}).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("fail to stop watch: %w", err)
	}

	return nil
}
