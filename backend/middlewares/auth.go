package middlewares

import (
	"context"
	"net/http"
	"strconv"
	"video_stream/utils"
)

func Authenticate(next http.Handler, rateLimiter *RedisRateLimiter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if token == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		userId, err := utils.VerifyToken(token)
		if err != nil {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		// Check rate limit
		allowed, err := rateLimiter.Allow(strconv.FormatInt(userId, 10))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !allowed {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
			return
		}

		// Set user ID in context (optional, depending on your application logic)
		ctx := context.WithValue(r.Context(), "userId", userId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

		// Call the next handler in the chain
		//next.ServeHTTP(w, r)
	})
}
