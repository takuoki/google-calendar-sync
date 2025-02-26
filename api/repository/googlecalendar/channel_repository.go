package googlecalendar

import (
	"context"
	"fmt"

	"github.com/takuoki/google-calendar-sync/api/domain/valueobject"
	calendar "google.golang.org/api/calendar/v3"
)

func (r *googleCalendarRepository) StopWatch(
	ctx context.Context, channelID valueobject.ChannelID) error {

	err := r.service.Channels.Stop(&calendar.Channel{
		Id: string(channelID),
	}).Context(ctx).Do()

	if err != nil {
		return fmt.Errorf("fail to stop watch: %w", err)
	}

	return nil
}
