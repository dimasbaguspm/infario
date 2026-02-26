package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Emit publishes an event to a Redis queue (list).
// The event is marshaled to JSON and pushed to the queue.
func Emit(ctx context.Context, client *redis.Client, queue string, event interface{}) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = client.RPush(ctx, queue, string(data)).Err()
	if err != nil {
		return fmt.Errorf("failed to emit event to %s: %w", queue, err)
	}

	return nil
}
