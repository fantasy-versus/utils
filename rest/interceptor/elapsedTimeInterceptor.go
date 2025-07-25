package interceptor

import (
	"context"
	"encoding/base64"
	"math"
	"net"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/fantasy-versus/utils/crypto"
	"github.com/fantasy-versus/utils/log"
	"github.com/gorilla/mux"
)

type requestInfo struct {
	url  string
	hash string
}
type key int

const requestIDKey key = 0

var (
	totalRequests         uint64
	maxConcurrentRequests int32
	concurrentRequests    int32
	ch                    chan requestInfo
)

func LaunchMemStats() {
	ch = make(chan requestInfo)
	if log.LogLevel == log.TRACE {
		go memStats()
	}
}

// Intercepts the request and calculates the total run time from start to finish
func NewElapsedTimeInterceptor() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := float64(time.Now().UnixNano()) / float64(time.Millisecond)
			idRequest, _ := crypto.GetDataHash(time.Now())
			idRequestStr := base64.RawStdEncoding.EncodeToString(idRequest[:])
			ctx := context.WithValue(r.Context(), requestIDKey, idRequestStr)
			requestInfo := requestInfo{url: r.URL.Path, hash: base64.RawStdEncoding.EncodeToString(idRequest[:])}

			if log.LogLevel == log.TRACE {

				ch <- requestInfo

				atomic.AddInt32(&concurrentRequests, 1)
				if atomic.LoadInt32(&concurrentRequests) > atomic.LoadInt32(&maxConcurrentRequests) {
					atomic.StoreInt32(&maxConcurrentRequests, atomic.LoadInt32(&concurrentRequests))
				}
				atomic.AddUint64(&totalRequests, 1)
			}
			if r == nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Errorf("%s - Missing Request", requestInfo.hash)
				return
			}

			remoteAddress := GetRequesterIp(r)
			ip, _, _ := net.SplitHostPort(remoteAddress)
			log.Infof("**** %s - New Request Arrived: Requester ip is %s; Request info: [%s %s%s]", requestInfo.hash, ip, r.Method, r.Host, r.URL)

			defer func() {
				endTime := float64(time.Now().UnixNano()) / float64(time.Millisecond)
				elapsed := float64((endTime - startTime) / 1000)
				log.Infof("**** %s - Time consumed for query to %s is %.2f seconds", requestInfo.hash, r.URL.Path, math.Round(elapsed*100)/100)
				if log.LogLevel == log.TRACE {

					atomic.AddInt32(&concurrentRequests, -1)
					ch <- requestInfo
				}
			}()
			r.Header.Add(HEADER_REMOTE_IP, ip)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return "UNKNOWN"
}

func memStats() {
	var m runtime.MemStats
	for {
		requestInfo := <-ch
		runtime.ReadMemStats(&m)
		log.Debugf(
			"**** %s - Request to: %s - Total connections count: %d; Current connections count: %d; Max concurrent connections count: %d; Alloc = %v MiB; TotalAlloc = %v MiB; Sys = %v MiB; Num gc cycles = %v",
			requestInfo.hash,
			requestInfo.url,
			atomic.LoadUint64(&totalRequests),
			atomic.LoadInt32(&concurrentRequests),
			atomic.LoadInt32(&maxConcurrentRequests),
			m.Alloc/1024/1024,
			m.TotalAlloc/1024/1024,
			m.Sys/1024/1024,
			m.NumGC,
		)

	}
}
