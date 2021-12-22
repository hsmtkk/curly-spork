package reportrepo

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RedisImpl struct {
	sugar  *zap.SugaredLogger
	client *redis.Client
}

func New(sugar *zap.SugaredLogger, client *redis.Client) *RedisImpl {
	return &RedisImpl{sugar, client}
}

func (r *RedisImpl) PutReport(ctx context.Context, sha256, report string) error {
	r.sugar.Infow("PutReport", "sha256", sha256)
	if err := r.client.Set(ctx, sha256, report, 0).Err(); err != nil {
		return fmt.Errorf("redis set failed; %w", err)
	}
	return nil
}

func (r *RedisImpl) GetReport(ctx context.Context, sha256 string) (string, error) {
	r.sugar.Infow("GetReport", "sha256", sha256)
	report, err := r.client.Get(ctx, sha256).Result()
	if err != nil {
		return "", fmt.Errorf("redis get failed; %w", err)
	}
	return report, nil
}
