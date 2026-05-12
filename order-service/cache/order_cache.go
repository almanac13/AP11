package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"order-service/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

type OrderCache interface {
	GetOrder(ctx context.Context, id string) (*domain.Order, error)
	SetOrder(ctx context.Context, order *domain.Order) error
	DeleteOrder(ctx context.Context, id string) error
}

type RedisOrderCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisOrderCache(client *redis.Client, ttl time.Duration) *RedisOrderCache {
	return &RedisOrderCache{client: client, ttl: ttl}
}

func NewRedisClient(addr string) *redis.Client {
	return redis.NewClient(&redis.Options{Addr: addr})
}

func orderKey(id string) string {
	return "order:" + id
}

func (c *RedisOrderCache) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	value, err := c.client.Get(ctx, orderKey(id)).Result()
	if err != nil {
		return nil, err
	}

	var order domain.Order
	if err := json.Unmarshal([]byte(value), &order); err != nil {
		return nil, err
	}

	log.Printf("[Redis] cache HIT for order_id=%s", id)
	return &order, nil
}

func (c *RedisOrderCache) SetOrder(ctx context.Context, order *domain.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}

	if err := c.client.Set(ctx, orderKey(order.ID), data, c.ttl).Err(); err != nil {
		return err
	}

	log.Printf("[Redis] cached order_id=%s ttl=%s", order.ID, c.ttl)
	return nil
}

func (c *RedisOrderCache) DeleteOrder(ctx context.Context, id string) error {
	if err := c.client.Del(ctx, orderKey(id)).Err(); err != nil {
		return err
	}

	log.Printf("[Redis] invalidated order_id=%s", id)
	return nil
}

func IsCacheMiss(err error) bool {
	return errors.Is(err, redis.Nil)
}
