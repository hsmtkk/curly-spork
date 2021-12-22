package filerepo

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

func (r *RedisImpl) PutFile(ctx context.Context, sha256 string, data []byte) error {
	r.sugar.Infow("PutFile", "sha256", sha256)
	if err := r.client.Set(ctx, sha256, data, 0).Err(); err != nil {
		return fmt.Errorf("redis set failed; %w", err)
	}
	return nil
}

func (r *RedisImpl) GetFile(ctx context.Context, sha256 string) ([]byte, error) {
	r.sugar.Infow("GetFile", "sha256", sha256)
	data, err := r.client.Get(ctx, sha256).Bytes()
	if err != nil {
		return nil, fmt.Errorf("redis get failed; %w", err)
	}
	return data, nil
}
