package demo

import "context"

type ServiceIFace interface {
	ChatStream(ctx context.Context, question string) (chan ChatStream, error)
	WeatherStream(ctx context.Context, question string) (chan ChatStream, error)
}
