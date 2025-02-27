package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/asynccnu/Muxi_ClassList/internal/classLog"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/model"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/repo"
	"github.com/go-redis/redis"
	"time"
)

type Cache struct {
	cli *redis.Client
}

func NewCache(cli *redis.Client) *Cache {
	return &Cache{
		cli: cli,
	}
}

type RecycleElem struct {
	ClassID       string
	ExpireTime    time.Time
	IsManuallyAdd bool
}

func (c *Cache) AddClassIDToRecycleBin(ctx context.Context, stuID, year, semester, classID string, isManuallyAdd bool) error {
	key := c.generateRecycledBin(stuID, year, semester)

	//一般不会重复添加
	//暂时不考虑

	// 构造 RecycleElem 并序列化为 JSON
	elem := RecycleElem{
		ClassID:       classID,
		ExpireTime:    time.Now().Add(time.Hour * 24 * 15),
		IsManuallyAdd: isManuallyAdd,
	}
	data, err := json.Marshal(elem)
	if err != nil {
		return fmt.Errorf("failed to marshal RecycleElem: %w", err)
	}

	// Lua 脚本（原子操作）
	const luaScript = `
        redis.call('SADD', KEYS[1], ARGV[1])
        redis.call('EXPIRE', KEYS[1], ARGV[2])
        return 1
    `

	// 执行 Lua 脚本
	_, err = c.cli.Eval(
		luaScript,
		[]string{key}, // KEYS[1]: key
		string(data),  // ARGV[1]: JSON 数据
		30*24*60*60,   // ARGV[2]: 过期时间（秒）
	).Result()

	if err != nil {
		return fmt.Errorf("failed to atomically add to recycle bin: %w", err)
	}
	return nil
}

func (c *Cache) RemoveClassIDFromRecycleBin(ctx context.Context, stuID, year, semester string, classID string) (bool, error) {
	key := c.generateRecycledBin(stuID, year, semester)
	//先检查是否有key
	exist, err := c.cli.Exists(key).Result()
	if err != nil {
		return false, err
	}
	if exist == 0 {
		return false, errors.New("no data in recycle bin")
	}

	//先把成员读取出来
	members, err := c.cli.SMembers(key).Result()
	if err != nil {
		return false, err
	}
	//然后搜索对应的class_id
	for _, member := range members {
		var recycleElem RecycleElem
		err = json.Unmarshal([]byte(member), &recycleElem)
		if err != nil {
			continue
		}
		if recycleElem.ClassID == classID {
			//找到就删除
			err = c.cli.SRem(key, member).Err()
			if err != nil {
				return false, err
			}

			return recycleElem.IsManuallyAdd, nil
		}
	}
	return false, errors.New("can not find the data in recycle bin")
}

func (c *Cache) GetRecycledClassIDs(ctx context.Context, stuID, year, semester string) ([]string, error) {
	key := c.generateRecycledBin(stuID, year, semester)

	//先把成员读取出来
	members, err := c.cli.SMembers(key).Result()
	if err != nil {
		return nil, err
	}

	var res = make([]string, 0, len(members))

	var waitForDelete []string

	//然后搜索对应的class_id
	for _, member := range members {
		var recycleElem RecycleElem
		err = json.Unmarshal([]byte(member), &recycleElem)
		if err != nil {
			continue
		}

		if recycleElem.ExpireTime.Before(time.Now()) {
			waitForDelete = append(waitForDelete, member)
			continue
		}

		res = append(res, recycleElem.ClassID)
	}

	if len(waitForDelete) > 0 {

		err := c.cli.SRem(key, waitForDelete).Err()
		if err != nil {
			classLog.LogPrinter.Warnf("delete class[%v] in recycle bin failed : %v", waitForDelete, err)
		}
	}
	return res, nil
}

func (c *Cache) CheckRecycleBinElementExist(ctx context.Context, stuID, year, semester string, classID string) bool {
	key := c.generateRecycledBin(stuID, year, semester)
	//先把成员读取出来
	members, err := c.cli.SMembers(key).Result()
	if err != nil {
		return false
	}
	//然后搜索对应的class_id
	for _, member := range members {
		var recycleElem RecycleElem
		err = json.Unmarshal([]byte(member), &recycleElem)
		if err != nil {
			continue
		}
		if recycleElem.ClassID == classID {
			return true
		}
	}
	return false
}

func (c *Cache) GetClassIDList(ctx context.Context, stuID, year, semester string) ([]string, error) {
	key := c.generateSCKey(stuID, year, semester)
	exist, err := c.cli.Exists(key).Result()
	if err != nil || exist == 0 {
		return nil, repo.ErrCacheMiss
	}

	res, err := c.cli.SMembers(key).Result()
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, nil
	}
	return res, nil
}

func (c *Cache) GetClassesByID(ctx context.Context, classids ...string) ([]*model.ClassDO, []string, error) {
	if len(classids) == 0 {
		return nil, nil, nil
	}

	keys := make([]string, 0, len(classids))
	for _, classid := range classids {
		keys = append(keys, c.generateClassKey(classid))
	}

	res, err := c.cli.MGet(keys...).Result()
	if err != nil {
		return nil, classids, err
	}

	// 存储有效的 ClassDO 和未找到的 ClassID
	var foundClasses []*model.ClassDO
	var missingClassIDs []string

	// 遍历结果判断哪些键没有获取到
	for i, value := range res {
		if value == nil || value == redis.Nil {
			// 如果值是 nil 或 redis.Nil，表示该键在 Redis 中不存在
			missingClassIDs = append(missingClassIDs, classids[i])
		} else {
			// 如果有值，假设值是有效的，转换并添加到 foundClasses 中
			classDO := &model.ClassDO{}
			err := json.Unmarshal([]byte(value.(string)), classDO) // 假设返回的是 JSON 格式的字符串
			if err != nil {
				missingClassIDs = append(missingClassIDs, classids[i])
			} else {
				foundClasses = append(foundClasses, classDO)
			}
		}
	}
	return foundClasses, missingClassIDs, nil

}

func (c *Cache) SetClassIDList(ctx context.Context, stuID, year, semester string, classids ...string) error {
	if len(classids) == 0 {
		return nil
	}

	key := c.generateSCKey(stuID, year, semester)

	// 修复后的 Lua 脚本
	const luaScript = `
        -- 将过期时间转换为数字
        local expire = tonumber(ARGV[1])
        
        -- 批量添加元素（需要遍历 ARGV）
        for i = 2, #ARGV do
            redis.call('SADD', KEYS[1], ARGV[i])
        end
        
        -- 设置过期时间
        redis.call('EXPIRE', KEYS[1], expire)
        return 1
    `

	// 构造参数（确保 expire 是数字）
	expireSeconds := int((7 * 24 * time.Hour).Seconds())
	args := []interface{}{expireSeconds}
	for _, id := range classids {
		args = append(args, id)
	}

	// 执行脚本
	if err := c.cli.Eval(luaScript, []string{key}, args...).Err(); err != nil {
		return fmt.Errorf("lua script failed: %w", err)
	}
	return nil
}

func (c *Cache) AddClass(ctx context.Context, classes ...*model.ClassDO) error {
	if len(classes) == 0 {
		return nil
	}

	// 准备键值对和参数
	var kvs []string
	for _, class := range classes {
		if class == nil {
			return errors.New("nil class object")
		}

		k := c.generateClassKey(class.ID)
		v, err := json.Marshal(class)
		if err != nil {
			// 处理序列化错误，建议添加日志记录
			continue
		}
		kvs = append(kvs, k, string(v))
	}

	if len(kvs) == 0 {
		return nil // 所有数据序列化都失败
	}

	// 构造 Lua 脚本
	const luaScript = `
        local expire = tonumber(ARGV[1])
        for i = 2, #ARGV, 2 do
            local key = ARGV[i]
            local value = ARGV[i+1]
            redis.call('SET', key, value)
            redis.call('EXPIRE', key, expire)
        end
        return 'OK'
    `

	// 准备脚本参数：过期时间 + 键值对
	expireSeconds := int((7 * 24 * time.Hour).Seconds())
	args := make([]interface{}, 0, 1+len(kvs))
	args = append(args, expireSeconds)
	for _, kv := range kvs {
		args = append(args, kv)
	}

	// 执行脚本
	if err := c.cli.Eval(luaScript, []string{}, args...).Err(); err != nil {
		return fmt.Errorf("lua script failed: %w", err)
	}
	return nil
}

func (c *Cache) DeleteClassIDList(ctx context.Context, stuID, year, semester string) error {
	key := c.generateSCKey(stuID, year, semester)

	// 使用 DEL 命令删除整个集合
	if err := c.cli.Del(key).Err(); err != nil {
		return fmt.Errorf("delete %s cache classid_list failed : %w", stuID, err)
	}
	return nil
}

func (c *Cache) generateSCKey(stuID, year, semester string) string {
	return "scr:" + stuID + ":" + year + ":" + semester
}

func (c *Cache) generateClassKey(claID string) string {
	return "class:" + claID
}
func (c *Cache) generateRecycledBin(stuID, year, semester string) string {
	return "recycle_bin:" + stuID + ":" + year + ":" + semester
}
