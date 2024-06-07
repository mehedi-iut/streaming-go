package middlewares

import (
	"context"
	"net/http"
	"video_stream/utils"
)

func Authenticate(next http.Handler) http.Handler {
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

		// Set user ID in context (optional, depending on your application logic)
		ctx := context.WithValue(r.Context(), "userId", userId)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

		// Call the next handler in the chain
		//next.ServeHTTP(w, r)
	})
}
