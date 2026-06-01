# Rate limiter HTTP (Middleware + Token Bucket)

**A ideia.** Proteger uma API limitando requisições por segundo.

**A técnica.** O **Middleware (Decorator) pattern**: uma função que recebe um `http.Handler` e devolve outro `http.Handler`, adicionando comportamento. É a base de logging, auth, CORS, tracing — tudo em Go. Por baixo, use **Token Bucket** via `golang.org/x/time/rate`.

**Refinamento técnico.**

- Comece com um limiter global. Depois evolua para **um limiter por IP**, guardado num mapa com `sync.Mutex` e uma goroutine que limpa visitantes inativos (senão é vazamento de memória).
- Encadeie middlewares: `Logging(RateLimit(Auth(handler)))`. Repare como a ordem importa.

**Snippet-guia:**

```go
func RateLimit(limiter *rate.Limiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// limiter por IP
type visitors struct {
    mu  sync.Mutex
    m   map[string]*rate.Limiter
}
func (v *visitors) get(ip string) *rate.Limiter {
    v.mu.Lock()
    defer v.mu.Unlock()
    l, ok := v.m[ip]
    if !ok {
        l = rate.NewLimiter(rate.Every(time.Second), 5) // 5 req/s
        v.m[ip] = l
    }
    return l
}
```
