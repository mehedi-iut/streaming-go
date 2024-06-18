package middlewares

import (
   "net/http"
   "net/http/httptest"
   "testing"
   "time"
   "context"

   "github.com/redis/go-redis/v9"
   "github.com/stretchr/testify/assert"
)

func TestRateLimiter(t *testing.T) {
   // Initialize Redis client
   client := redis.NewClient(&redis.Options{
      Addr: "localhost:6379",
   })
   defer client.Close()

   // Clear any existing data
   client.FlushDB(context.Background())

   // Create a new rate limiter
   rateLimiter := NewRedisRateLimiter(client, time.Second*10, 5) // 5 requests per 10 seconds

   handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      w.WriteHeader(http.StatusOK)
      w.Write([]byte("OK"))
   })

   // Create a test server with the rate-limited middleware
   testServer := httptest.NewServer(Authenticate(handler, rateLimiter))
   defer testServer.Close()

   // Helper function to make requests
   makeRequest := func() *http.Response {
      req, _ := http.NewRequest("GET", testServer.URL, nil)
      req.Header.Set("Authorization", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im1laGVkaTE2NzBAZ21haWwuY29tIiwiZXhwIjoxNzE4NzEwNTcyLCJ1c2VySWQiOjF9.GQ28gUGlNiM21TnlMKTZc9NOJsfoQvzc3yozXXTV2iQ")
      client := &http.Client{}
      resp, _ := client.Do(req)
      return resp
   }

   // Test within rate limit
   for i := 0; i < 150; i++ {
      resp := makeRequest()
      assert.Equal(t, http.StatusOK, resp.StatusCode)
   }

   // Test exceeding rate limit
   resp := makeRequest()
   assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

   // Wait for the window to reset
   time.Sleep(10 * time.Second)

   // Test after rate limit window reset
   resp = makeRequest()
   assert.Equal(t, http.StatusOK, resp.StatusCode)
}
