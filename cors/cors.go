package cors

import (
	"context"
	"github.com/rs/cors"
	"github.com/twitchtv/twirp"
)

func CorsWrapper() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"localhost:8080"},
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Content-Type"},
	})
}

func CorsInterceptor() twirp.Interceptor {
	return func(next twirp.Method) twirp.Method {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			err := twirp.SetHTTPResponseHeader(ctx, "access-control-allow-origin", "localhost:8080")
			if err != nil {
				return nil, twirp.InternalErrorWith(err)
			}
			return next(ctx, req)
		}
	}
}
