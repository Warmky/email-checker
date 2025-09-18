package cache

// go mod tidy
// cache.go

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
	ctx    context.Context
	ttl    time.Duration
}

type CacheEntry struct {
	TimeStamp time.Time              `json:"timeStamp"`
	Response  map[string]interface{} `json:"response"`
}

func NewRedisCache(addr, password string, db int, ttl time.Duration) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		ctx:    ctx,
		ttl:    ttl,
	}, nil
}

func (r *RedisCache) Get(domain string) (CacheEntry, bool) {
	data, err := r.client.Get(r.ctx, domain).Bytes()
	if err != nil {
		return CacheEntry{}, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return CacheEntry{}, false
	}

	return entry, true
}

func (r *RedisCache) Set(domain string, entry CacheEntry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return r.client.Set(r.ctx, domain, data, r.ttl).Err()
}

func (r *RedisCache) Delete(domain string) error {
	return r.client.Del(r.ctx, domain).Err()
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

/*
func main() {
	// 创建Redis缓存实例（假设Redis运行在localhost:6379）
	cache, err := NewRedisCache("localhost:6379", "", 0, 0) // 1*time.Hour 缓存存在的时间
	if err != nil {
		log.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer cache.Close()

	// 测试数据
	testDomain := "163.com"
	testResponse := map[string]interface{}{
		"autodiscover":  "autodiscover_result",
		"autoconfig":    "autoconfig_result",
		"srv":           "srv_result",
		"guess":         "guess_result",
		"recentResults": []string{"scan1", "scan2", "scan3"},
	}

	testEntry := CacheEntry{
		TimeStamp: time.Now(),
		Response:  testResponse,
	}

	fmt.Println("=== Redis缓存测试 ===")

	// 测试1: 设置缓存
	fmt.Println("1. 设置缓存...")
	err = cache.Set(testDomain, testEntry)
	if err != nil {
		log.Fatalf("设置缓存失败: %v", err)
	}
	fmt.Println("✅ 缓存设置成功")

	// 测试2: 获取缓存
	fmt.Println("2. 获取缓存...")
	entry, exists := cache.Get(testDomain)
	if !exists {
		log.Fatalf("获取缓存失败: 缓存不存在")
	}
	fmt.Printf("✅ 缓存获取成功:\n")
	fmt.Printf("   - 时间戳: %v\n", entry.TimeStamp)
	fmt.Printf("   - autodiscover: %v\n", entry.Response["autodiscover"])
	fmt.Printf("   - autoconfig: %v\n", entry.Response["autoconfig"])
	fmt.Printf("   - recentResults: %v\n", entry.Response["recentResults"])

	testEntry = CacheEntry{
		TimeStamp: time.Now(),
		Response:  testResponse,
	}
	err = cache.Set(testDomain, testEntry)
	if err != nil {
		log.Fatalf("设置缓存失败: %v", err)
	}
	fmt.Println("✅ 缓存设置成功")

	entry, exists = cache.Get(testDomain)
	if !exists {
		log.Fatalf("获取缓存失败: 缓存不存在")
	}
	fmt.Printf("✅ 缓存获取成功:\n")
	fmt.Printf("   - 时间戳: %v\n", entry.TimeStamp)
	fmt.Printf("   - autodiscover: %v\n", entry.Response["autodiscover"])
	fmt.Printf("   - autoconfig: %v\n", entry.Response["autoconfig"])
	fmt.Printf("   - recentResults: %v\n", entry.Response["recentResults"])

}
*/
