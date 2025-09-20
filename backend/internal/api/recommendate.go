package api

import (
	"backend/internal/cache"
	"encoding/json"
	"net/http"
)

func RecommendedDomainsHandler(w http.ResponseWriter, r *http.Request, redisCache *cache.RedisCache) {
	// 这里可以指定几个热门域名
	domains := []string{"qq.com", "outlook.com", "gmail.com"}
	results := make([]map[string]interface{}, 0)

	for _, domain := range domains {
		if redisCache != nil {
			if entry, exists := redisCache.Get(domain); exists {
				results = append(results, map[string]interface{}{
					"domain":   domain,
					"response": entry.Response, // 直接把缓存内容返回
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
