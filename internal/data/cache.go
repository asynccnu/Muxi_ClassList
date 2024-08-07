package data

import (
	"class/internal/biz"
	"encoding/json"
	"fmt"
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
func (r *RedisCache) AddEleToSet(stuId string, xnm, xqm, claId string) error {
	//score := float64(day)*100 + float64(st+end)/2.0
	err := r.client.SAdd(GenerateSetName(stuId, xnm, xqm), claId).Err()
	//TODO:设置过期时间
	return err
}
func (r *RedisCache) GetClassIDFromSet(stuId, xnm, xqm string) ([]string, error) {
	return r.client.SMembers(GenerateSetName(stuId, xnm, xqm)).Result()
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
func GenerateSetName(stuId, xnm, xqm string) string {
	return fmt.Sprintf("StuAndCla:%s:%s:%s", stuId, xnm, xqm)
}
