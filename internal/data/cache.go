package data

import (
	"class/internal/biz"
	"encoding/json"
	"github.com/go-redis/redis"
	"time"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(cli *redis.Client) Cache {
	return &RedisCache{
		client: cli,
	}
}

func (r *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.client.Set(key, val, expiration).Err()
}
func (r *RedisCache) AddEleToZset(stuId string, claId string, day, st, end int64) error {
	score := float64(day)*100 + float64(st+end)/2.0
	err := r.client.ZAdd(GenerateZsetName(stuId), redis.Z{
		Member: claId,
		Score:  score,
	}).Err()
	return err
}
func (r *RedisCache) GetClassIDFromZset(stuId string) ([]string, error) {
	return r.client.ZRange(GenerateZsetName(stuId), 0, -1).Result()
}
func (r *RedisCache) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	return r.client.Scan(cursor, match, count).Result()
}
func (r *RedisCache) GetClassInfo(key string) (*biz.ClassInfo, error) {
	var classInfo = &biz.ClassInfo{}
	val, err := r.client.Get(key).Result()
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(val), &classInfo)
	if err != nil {
		return nil, err
	}
	return classInfo, nil
}
func (r *RedisCache) ScanKeys(pattern string) ([]string, error) {
	var cursor uint64
	var keys []string

	for {
		scannedKeys, newCursor, err := r.client.Scan(cursor, pattern, 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, scannedKeys...)
		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}
func (r *RedisCache) DeleteKey(key string) error {
	err := r.client.Del(key).Err()
	return err
}
func GenerateZsetName(stu_id string) string {
	return "stu_id:" + stu_id
}
