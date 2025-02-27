package cache

import (
	"context"
	"encoding/json"
	"github.com/alicebob/miniredis/v2"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/model"
	"github.com/asynccnu/Muxi_ClassList/internal/data/class/repo"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func initRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	s, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	//defer s.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})
	return rdb, s
}

func TestCache_generateSCKey(t *testing.T) {
	type args struct {
		stuID    string
		year     string
		semester string
	}

	tests := []struct {
		name string
		arg  args
		want string
	}{
		{"success", args{"123", "2023", "1"}, "scr:123:2023:1"},
	}

	cache := new(Cache)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := cache.generateSCKey(tt.arg.stuID, tt.arg.year, tt.arg.semester)
			assert.Equal(t, res, tt.want)
		})
	}
}
func TestCache_generateRecycledBin(t *testing.T) {
	type args struct {
		stuID    string
		year     string
		semester string
	}

	tests := []struct {
		name string
		arg  args
		want string
	}{
		{"success", args{"123", "2023", "1"}, "recycle_bin:123:2023:1"},
	}

	cache := new(Cache)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := cache.generateRecycledBin(tt.arg.stuID, tt.arg.year, tt.arg.semester)
			assert.Equal(t, res, tt.want)
		})
	}
}

func TestCache_AddClassIDToRecycleBin(t *testing.T) {
	rdb, s := initRedis(t)
	defer s.Close()

	cache := &Cache{cli: rdb}

	type args struct {
		stuID    string
		year     string
		semester string
		classID  string
		added    bool
	}
	tests := []struct {
		name    string
		args    args
		mock    func(*miniredis.Miniredis) // 用于前置条件或错误模拟
		wantErr assert.ErrorAssertionFunc
		// 新增验证函数
		check func(t *testing.T, s *miniredis.Miniredis, cache *Cache, args args)
	}{
		{
			name: "success",
			args: args{"123", "2023", "1", "1", false},
			mock: func(m *miniredis.Miniredis) {
				// 无需前置操作，或模拟其他初始状态
			},
			wantErr: assert.NoError,
			check: func(t *testing.T, s *miniredis.Miniredis, cache *Cache, args args) {
				key := cache.generateRecycledBin(args.stuID, args.year, args.semester)

				// 验证 ClassID 是否在集合中
				members, err1 := s.SMembers(key)
				assert.NoError(t, err1)

				var ids []string

				for _, member := range members {
					var tmp RecycleElem
					err := json.Unmarshal([]byte(member), &tmp)
					assert.NoError(t, err)
					ids = append(ids, tmp.ClassID)
				}
				assert.Contains(t, ids, args.classID)

				// 验证 TTL 是否正确（允许1秒误差）
				expectedTTL := 30 * 24 * time.Hour
				actualTTL := s.TTL(key)
				assert.True(t, actualTTL <= expectedTTL && actualTTL >= expectedTTL-time.Second,
					"Expected TTL ~15 days, got %v", actualTTL)
			},
		},
		// 可添加其他用例（如 Redis 错误、重复添加等）
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll() // 确保测试独立性
			if tt.mock != nil {
				tt.mock(s) // 执行前置条件设置
			}

			err := cache.AddClassIDToRecycleBin(context.Background(),
				tt.args.stuID, tt.args.year, tt.args.semester, tt.args.classID, tt.args.added)

			tt.wantErr(t, err)
			if tt.check != nil {
				tt.check(t, s, cache, tt.args) // 执行后置验证
			}
		})
	}
}

func TestCache_RemoveClassIDFromRecycleBin(t *testing.T) {
	rdb, s := initRedis(t)
	defer s.Close()

	cache := &Cache{cli: rdb}

	type args struct {
		stuID      string
		year       string
		semester   string
		classID    string
		expireTime time.Time
		added      bool
	}
	tests := []struct {
		name    string
		args    args
		prepare func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args)
		check   func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "classID exists and the set has one member",
			args: args{
				stuID:      "123",
				year:       "2023",
				semester:   "1",
				classID:    "1",
				expireTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				added:      true,
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args) {
				key := cache.generateRecycledBin(args2.stuID, args2.year, args2.semester)

				data, err := json.Marshal(&RecycleElem{
					ClassID:       args2.classID,
					ExpireTime:    args2.expireTime,
					IsManuallyAdd: args2.added,
				})
				assert.NoError(tt, err)

				num, err := miniredis2.SAdd(key, string(data))
				assert.NoError(tt, err)

				//检查是否添加成功
				assert.Equal(tt, 1, num)
			},
			check: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args) {
				key := cache.generateRecycledBin(args2.stuID, args2.year, args2.semester)
				exist := miniredis2.Exists(key)
				assert.Equal(tt, false, exist)

				//members, err := miniredis2.SMembers(key)
				//assert.NoError(tt, err)
				//
				//assert.NotContains(tt, members, args2.classID, "ClassID should be deleted from the set")
			},
			wantErr: assert.NoError,
		},
		{
			name: "class_id exist but the set has two members",
			args: args{
				stuID:      "123",
				year:       "2023",
				semester:   "1",
				classID:    "1",
				expireTime: time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC),
				added:      true,
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args) {
				key := cache.generateRecycledBin(args2.stuID, args2.year, args2.semester)
				data1, err := json.Marshal(&RecycleElem{
					ClassID:       args2.classID,
					ExpireTime:    args2.expireTime,
					IsManuallyAdd: args2.added,
				})

				assert.NoError(tt, err)

				data2, err := json.Marshal(&RecycleElem{
					ClassID:       args2.classID + args2.classID,
					ExpireTime:    args2.expireTime,
					IsManuallyAdd: !args2.added,
				})
				assert.NoError(tt, err)

				num, err := miniredis2.SAdd(key, string(data1), string(data2))
				assert.NoError(tt, err)

				//检查是否添加成功
				assert.Equal(tt, 2, num)
			},
			check: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args) {
				key := cache.generateRecycledBin(args2.stuID, args2.year, args2.semester)

				members, err := miniredis2.SMembers(key)
				assert.NoError(tt, err)

				data1, err := json.Marshal(&RecycleElem{
					ClassID:       args2.classID,
					ExpireTime:    args2.expireTime,
					IsManuallyAdd: args2.added,
				})
				assert.NoError(tt, err)

				data2, err := json.Marshal(&RecycleElem{
					ClassID:       args2.classID + args2.classID,
					ExpireTime:    args2.expireTime,
					IsManuallyAdd: !args2.added,
				})
				assert.NoError(tt, err)

				assert.NotContains(tt, members, string(data1), "ClassID should be deleted from the set")
				assert.Contains(tt, members, string(data2), "ClassID should be deleted from the set")
			},
			wantErr: assert.NoError,
		},
		{
			name: "the set don't exist the key",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classID:  "1",
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args) {},
			check:   func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args) {},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			tt.prepare(t, s, tt.args)
			res, err := cache.RemoveClassIDFromRecycleBin(context.Background(), tt.args.stuID, tt.args.year, tt.args.semester, tt.args.classID)
			assert.Equal(t, tt.args.added, res)
			tt.wantErr(t, err)
			tt.check(t, s, tt.args)
		})
	}
}

func TestCache_GetRecycledClassIDs(t *testing.T) {
	rdb, s := initRedis(t)
	defer s.Close()

	cache := &Cache{cli: rdb}

	type args struct {
		stuID      string
		year       string
		semester   string
		classIDs   []string
		expireTime []time.Time
	}
	tests := []struct {
		name    string
		args    args
		prepare func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args)
		check   func(tt *testing.T, arg args, res []string)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "don't have members",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classIDs: []string{},
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args) {
				//key := cache.generateRecycledBin(args2.stuID,args2.year,args2.semester)
			},
			check: func(tt *testing.T, arg args, res []string) {
				assert.ElementsMatch(tt, arg.classIDs, res)
			},
			wantErr: assert.NoError,
		},
		{
			name: "have members but someone expires",
			args: args{
				stuID:      "123",
				year:       "2023",
				semester:   "1",
				classIDs:   []string{"1", "2", "3"},
				expireTime: []time.Time{time.Now().Add(-1 * time.Hour), time.Now().Add(1 * time.Hour), time.Now().Add(1 * time.Hour)},
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args) {
				key := cache.generateRecycledBin(args2.stuID, args2.year, args2.semester)

				var data []string
				for i := 0; i < len(args2.classIDs); i++ {
					tmp, err := json.Marshal(&RecycleElem{
						ClassID:       args2.classIDs[i],
						ExpireTime:    args2.expireTime[i],
						IsManuallyAdd: false,
					})
					assert.NoError(tt, err)
					data = append(data, string(tmp))
				}
				num, err := miniredis2.SAdd(key, data...)
				assert.NoError(tt, err)
				assert.Equal(tt, len(args2.classIDs), num, "Set members should be added correctly")
			},
			check: func(tt *testing.T, arg args, res []string) {
				assert.ElementsMatch(tt, []string{"2", "3"}, res)
			},
			wantErr: assert.NoError,
		},
		{
			name: "have members but everyone alive",
			args: args{
				stuID:      "123",
				year:       "2023",
				semester:   "1",
				classIDs:   []string{"1", "2", "3"},
				expireTime: []time.Time{time.Now().Add(1 * time.Hour), time.Now().Add(1 * time.Hour), time.Now().Add(1 * time.Hour)},
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args2 args) {
				key := cache.generateRecycledBin(args2.stuID, args2.year, args2.semester)

				var data []string
				for i := 0; i < len(args2.classIDs); i++ {
					tmp, err := json.Marshal(&RecycleElem{
						ClassID:       args2.classIDs[i],
						ExpireTime:    args2.expireTime[i],
						IsManuallyAdd: false,
					})
					assert.NoError(tt, err)
					data = append(data, string(tmp))
				}
				num, err := miniredis2.SAdd(key, data...)
				assert.NoError(tt, err)
				assert.Equal(tt, len(args2.classIDs), num, "Set members should be added correctly")
			},
			check: func(tt *testing.T, arg args, res []string) {
				assert.ElementsMatch(tt, []string{"1", "2", "3"}, res)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			tt.prepare(t, s, tt.args)
			res, err := cache.GetRecycledClassIDs(context.Background(), tt.args.stuID, tt.args.year, tt.args.semester)
			tt.wantErr(t, err)
			tt.check(t, tt.args, res)
		})
	}
}

func TestCache_CheckRecycleBinElementExist(t *testing.T) {
	rdb, s := initRedis(t)
	defer s.Close()

	cache := &Cache{cli: rdb}

	type args struct {
		stuID    string
		year     string
		semester string
		classID  string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(tt *testing.T, miniredis2 *miniredis.Miniredis, args args)
		check   func(tt *testing.T, arg args, res bool)
	}{
		{
			name: "key is not exist",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classID:  "1",
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args args) {
				//key := cache.generateRecycledBin(args.stuID,args.year,args.semester)
			},
			check: func(tt *testing.T, arg args, res bool) {
				assert.False(tt, res)
			},
		},
		{
			name: "key is  exist",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classID:  "1",
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args args) {
				key := cache.generateRecycledBin(args.stuID, args.year, args.semester)

				data, err := json.Marshal(&RecycleElem{
					ClassID:       args.classID,
					ExpireTime:    time.Now(),
					IsManuallyAdd: false,
				})
				assert.NoError(tt, err)

				num, err := miniredis2.SAdd(key, string(data))
				assert.NoError(tt, err)
				assert.Equal(tt, 1, num, "Set members should be added correctly")
			},
			check: func(tt *testing.T, arg args, res bool) {
				assert.True(tt, res)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			tt.prepare(t, s, tt.args)
			res := cache.CheckRecycleBinElementExist(context.Background(), tt.args.stuID, tt.args.year, tt.args.semester, tt.args.classID)
			tt.check(t, tt.args, res)
		})
	}
}

func TestCache_GetClassIDList(t *testing.T) {
	rdb, s := initRedis(t)
	defer s.Close()

	cache := &Cache{cli: rdb}

	type args struct {
		stuID    string
		year     string
		semester string
		classIDs []string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(tt *testing.T, miniredis2 *miniredis.Miniredis, args args)
		check   func(tt *testing.T, arg args, res []string, err error)
	}{
		{
			name: "key is not exist",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classIDs: []string{},
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args args) {
				//key := cache.generateRecycledBin(args.stuID,args.year,args.semester)
			},
			check: func(tt *testing.T, arg args, res []string, err error) {
				assert.Nil(tt, res)
				assert.ErrorIs(tt, err, repo.ErrCacheMiss)
			},
		},
		{
			name: "key is exist but is empty",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classIDs: []string{},
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args args) {
				key := cache.generateSCKey(args.stuID, args.year, args.semester)
				_, err := miniredis2.SAdd(key, args.classIDs...)
				assert.NoError(tt, err)

				exist := miniredis2.Exists(key)
				assert.True(tt, exist)

				members, err := miniredis2.SMembers(key)
				assert.NoError(tt, err)
				assert.Equal(tt, 0, len(members))
				//assert.ElementsMatch(tt, members, args.classIDs)
			},
			check: func(tt *testing.T, arg args, res []string, err error) {
				assert.Nil(tt, res)
				assert.Nil(tt, err)
			},
		},
		{
			name: "key is exist but is logically empty",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classIDs: []string{},
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args args) {
				key := cache.generateSCKey(args.stuID, args.year, args.semester)
				_, err := miniredis2.SAdd(key, RedisNULL)
				assert.NoError(tt, err)

				exist := miniredis2.Exists(key)
				assert.True(tt, exist)

				members, err := miniredis2.SMembers(key)
				assert.NoError(tt, err)
				assert.Contains(tt, members, RedisNULL)
				//assert.ElementsMatch(tt, members, args.classIDs)
			},
			check: func(tt *testing.T, arg args, res []string, err error) {
				assert.Nil(tt, res)
				assert.Nil(tt, err)
			},
		},
		{
			name: "key is exist but is not empty",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classIDs: []string{"1", "2"},
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, args args) {
				key := cache.generateSCKey(args.stuID, args.year, args.semester)
				_, err := miniredis2.SAdd(key, args.classIDs...)
				assert.NoError(tt, err)

				exist := miniredis2.Exists(key)
				assert.True(tt, exist)

				members, err := miniredis2.SMembers(key)
				assert.NoError(tt, err)
				assert.Equal(tt, len(args.classIDs), len(members))
				//assert.ElementsMatch(tt, members, args.classIDs)
			},
			check: func(tt *testing.T, arg args, res []string, err error) {
				assert.ElementsMatch(tt, res, arg.classIDs)
				assert.NoError(tt, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			tt.prepare(t, s, tt.args)
			res, err := cache.GetClassIDList(context.Background(), tt.args.stuID, tt.args.year, tt.args.semester)
			tt.check(t, tt.args, res, err)
		})
	}
}

func TestCache_GetClassesByID(t *testing.T) {
	rdb, s := initRedis(t)
	defer s.Close()

	cache := &Cache{cli: rdb}

	type args struct {
		classids []string
	}

	var class1 = &model.ClassDO{
		ID: "1",
	}
	var class2 = &model.ClassDO{
		ID: "2",
	}

	tests := []struct {
		name    string
		args    args
		prepare func(tt *testing.T, miniredis2 *miniredis.Miniredis, arg args)
		want    []*model.ClassDO
		want1   []string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: args{
				classids: []string{"1", "2"},
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, arg args) {
				key1 := cache.generateClassKey(class1.ID)
				key2 := cache.generateClassKey(class2.ID)

				data1, err := json.Marshal(class1)
				assert.NoError(tt, err)
				data2, err := json.Marshal(class2)
				assert.NoError(tt, err)
				err = miniredis2.Set(key1, string(data1))
				assert.NoError(tt, err)
				err = miniredis2.Set(key2, string(data2))
			},
			want:    []*model.ClassDO{class1, class2},
			want1:   nil,
			wantErr: assert.NoError,
		},
		{
			name: "There are courses that do not exist",
			args: args{
				classids: []string{"1", "2"},
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, arg args) {
				key1 := cache.generateClassKey(class1.ID)
				//key2 := cache.generateClassKey(class2.ID)

				data1, err := json.Marshal(class1)
				assert.NoError(tt, err)
				//data2, err := json.Marshal(class2)
				//assert.NoError(tt, err)
				err = miniredis2.Set(key1, string(data1))
				assert.NoError(tt, err)
				//err = miniredis2.Set(key2, string(data2))
			},
			want:    []*model.ClassDO{class1},
			want1:   []string{"2"},
			wantErr: assert.NoError,
		},
		{
			name: "len(class_id)==0",
			args: args{
				classids: []string{},
			},
			prepare: func(tt *testing.T, miniredis2 *miniredis.Miniredis, arg args) {
			},
			want:    nil,
			want1:   nil,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			tt.prepare(t, s, tt.args)
			got, got1, err := cache.GetClassesByID(context.Background(), tt.args.classids...)
			tt.wantErr(t, err)
			assert.ElementsMatch(t, got, tt.want)
			assert.ElementsMatch(t, got1, tt.want1)
		})
	}
}

func TestCache_SetClassIDList(t *testing.T) {
	rdb, s := initRedis(t)
	defer s.Close()

	cache := &Cache{cli: rdb}

	type args struct {
		stuID    string
		year     string
		semester string
		classids []string
	}
	tests := []struct {
		name    string
		args    args
		check   func(t *testing.T, s *miniredis.Miniredis, args2 args)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classids: []string{"1"},
			},
			check: func(t *testing.T, s *miniredis.Miniredis, args2 args) {
				key := cache.generateSCKey(args2.stuID, args2.year, args2.semester)

				members, err := s.SMembers(key)
				assert.NoError(t, err)

				assert.ElementsMatch(t, args2.classids, members)
				// 验证 TTL 是否正确（允许1秒误差）
				expectedTTL := 7 * 24 * time.Hour
				actualTTL := s.TTL(key)
				assert.True(t, actualTTL <= expectedTTL && actualTTL >= expectedTTL-time.Second,
					"Expected TTL ~7 days, got %v", actualTTL)
			},
			wantErr: assert.NoError,
		},
		{
			name: "length is 0",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classids: nil,
			},
			check: func(t *testing.T, s *miniredis.Miniredis, args2 args) {
				key := cache.generateSCKey(args2.stuID, args2.year, args2.semester)

				members, err := s.SMembers(key)
				assert.NoError(t, err)

				assert.Contains(t, members, RedisNULL)
				// 验证 TTL 是否正确（允许1秒误差）
				expectedTTL := 7 * 24 * time.Hour
				actualTTL := s.TTL(key)
				assert.True(t, actualTTL <= expectedTTL && actualTTL >= expectedTTL-time.Second,
					"Expected TTL ~7 days, got %v", actualTTL)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			err := cache.SetClassIDList(context.Background(), tt.args.stuID, tt.args.year, tt.args.semester, tt.args.classids...)
			tt.wantErr(t, err)
			tt.check(t, s, tt.args)
		})
	}
}

func TestCache_AddClass(t *testing.T) {
	rdb, s := initRedis(t)
	defer s.Close()

	cache := &Cache{cli: rdb}

	class1 := &model.ClassDO{ID: "1"}
	class2 := &model.ClassDO{ID: "2"}

	type args struct {
		classes []*model.ClassDO
	}

	tests := []struct {
		name    string
		args    args
		check   func(*testing.T, *miniredis.Miniredis, args)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "add one",
			args: args{
				classes: []*model.ClassDO{class1},
			},
			check: func(t *testing.T, m *miniredis.Miniredis, args args) {
				key := cache.generateClassKey(args.classes[0].ID)

				val, err := m.Get(key)
				assert.NoError(t, err)
				var class model.ClassDO
				err = json.Unmarshal([]byte(val), &class)
				assert.NoError(t, err)
				assert.Equal(t, &class, args.classes[0])
			},
			wantErr: assert.NoError,
		},
		{
			name: "add two,but someone is nil",
			args: args{
				classes: []*model.ClassDO{class1, nil},
			},
			check: func(t *testing.T, m *miniredis.Miniredis, args args) {
				key := cache.generateClassKey(args.classes[0].ID)

				_, err := m.Get(key)
				assert.Error(t, err)
			},
			wantErr: assert.Error,
		},
		{
			name: "add two",
			args: args{
				classes: []*model.ClassDO{class1, class2},
			},
			check: func(t *testing.T, m *miniredis.Miniredis, args args) {
				key1 := cache.generateClassKey(args.classes[0].ID)
				key2 := cache.generateClassKey(args.classes[1].ID)
				val, err := m.Get(key1)
				assert.NoError(t, err)

				var class model.ClassDO
				err = json.Unmarshal([]byte(val), &class)
				assert.NoError(t, err)
				assert.Equal(t, &class, args.classes[0])

				val, err = m.Get(key2)
				assert.NoError(t, err)
				err = json.Unmarshal([]byte(val), &class)
				assert.NoError(t, err)
				assert.Equal(t, &class, args.classes[1])
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			err := cache.AddClass(context.Background(), tt.args.classes...)
			tt.wantErr(t, err)
			tt.check(t, s, tt.args)
		})
	}
}

func TestCache_DeleteClassIDList(t *testing.T) {
	rdb, s := initRedis(t)
	defer s.Close()

	cache := &Cache{cli: rdb}

	type args struct {
		stuID    string
		year     string
		semester string

		classIDs []string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(t *testing.T, s *miniredis.Miniredis, args2 args)
		check   func(t *testing.T, s *miniredis.Miniredis, args2 args)
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "key is not exist",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classIDs: nil,
			},
			prepare: func(t *testing.T, s *miniredis.Miniredis, args2 args) {
			},
			check: func(t *testing.T, s *miniredis.Miniredis, args2 args) {
			},
			wantErr: assert.NoError,
		},
		{
			name: "key is  exist",
			args: args{
				stuID:    "123",
				year:     "2023",
				semester: "1",
				classIDs: []string{"1", "2", "3"},
			},
			prepare: func(t *testing.T, s *miniredis.Miniredis, args2 args) {
				key := cache.generateSCKey(args2.stuID, args2.year, args2.semester)
				num, err := s.SAdd(key, args2.classIDs...)
				assert.NoError(t, err)
				assert.Equal(t, len(args2.classIDs), num)

				exist := s.Exists(key)
				assert.True(t, exist)
			},
			check: func(t *testing.T, s *miniredis.Miniredis, args2 args) {
				key := cache.generateSCKey(args2.stuID, args2.year, args2.semester)
				exist := s.Exists(key)
				assert.False(t, exist)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.FlushAll()
			tt.prepare(t, s, tt.args)
			err := cache.DeleteClassIDList(context.Background(), tt.args.stuID, tt.args.year, tt.args.semester)
			tt.wantErr(t, err)
			tt.check(t, s, tt.args)
		})
	}
}
