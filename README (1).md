# source-asia-backend

A single Go service that handles two things: rate-limited request ingestion and a product catalog with media support. No external dependencies, no database — just the standard library and in-memory state.

---

## How to Run

 from the project root:

```bash
go run cmd/server/main.go
```

The server starts on `:8080`. You'll see a confirmation in the terminal once it's listening.

Open a second terminal window to run the commands below while keeping the server running.

---

## Part 1 — Rate Limiter

### Endpoints

**POST /request**

Accepts a `user_id` and any JSON payload. Returns `201 Created` on success — this signals that a new timestamped entry was written to memory, not just acknowledged.

Returns `400 Bad Request` if `user_id` is missing or empty, or if the request body isn't valid JSON.
Returns `429 Too Many Requests` once a user exceeds 5 requests in the current rolling window.

```bash
curl -X POST "http://localhost:8080/request" \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user_123", "payload": {"action": "test_ping"}}'
```

Run that 6 times in a row — the first 5 go through, the 6th gets a 429.

**GET /stats**

Returns per-user metrics. Query by `user_id`:

```bash
curl -X GET "http://localhost:8080/stats?user_id=user_123"
```

Response shape:

```json
{
  "accepted_current_window": 3,
  "rejected_cumulative": 14
}
```

- `accepted_current_window` — how many requests this user has had accepted in the last 60 seconds
- `rejected_cumulative` — lifetime total of rate-limited rejections for this user (not reset per window, by design — useful for spotting persistent abusers)

### Rate Limiting Approach

Uses a **rolling window** (not a fixed window). Each accepted request stores an exact timestamp. On every new request, the system counts how many timestamps fall within the last 60 seconds from right now. This avoids the burst problem you get with fixed windows (where a user can fire 5 requests at 11:59 and another 5 at 12:00 and stay "within limits").

Every window evaluation is wrapped in a `sync.Mutex` to prevent race conditions — parallel requests for the same `user_id` cannot both read a count of 4 and both get approved, pushing the user to 6.

### PowerShell equivalents (Windows)

```powershell
# Check stats
Invoke-RestMethod -Uri "http://localhost:8080/stats?user_id=user_123" -Method Get

# Send a request
Invoke-RestMethod -Uri "http://localhost:8080/request" -Method Post `
  -ContentType "application/json" `
  -Body '{"user_id": "user_123", "payload": {"action": "test_ping"}}'
```

---

## Part 2 — Product Catalog

### Endpoints

**POST /products**

Creates a new product. SKUs must be unique — duplicates return `409 Conflict`.

```bash
curl -X POST "http://localhost:8080/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Gaming Mouse",
    "sku": "MOUSE-999",
    "image_urls": ["https://links.com/img1.jpg"],
    "video_urls": []
  }'
```

Returns `201 Created` with the full product including the server-assigned `id`.

**GET /products**

Returns a paginated list. Deliberately excludes full media arrays — only counts are returned here.

```bash
curl -X GET "http://localhost:8080/products?page=1&limit=10"
```

Default: `page=1`, `limit=10`. Maximum `limit` is `100`.

List item shape:

```json
{
  "id": "prod_1",
  "name": "Gaming Mouse",
  "sku": "MOUSE-999",
  "image_count": 1,
  "video_count": 0
}
```

**GET /products/{id}**

Full product detail including all media URLs. Returns `404 Not Found` for unknown IDs.

**POST /products/{id}/media**

Appends additional image or video URLs to an existing product. At least one of `image_urls` or `video_urls` must be present and non-empty, otherwise returns `400`.

### Validation Rules

- `name` and `sku` cannot be empty
- URLs must start with `http://` or `https://` and be 2048 characters or under
- No binary uploads, no base64 — URL strings only
- Max 20 URLs per array per request (21+ returns `400`)
- Duplicate `sku` on create returns `409 Conflict`

### Why the list endpoint doesn't return full media arrays

With 1,000 products and 10 images each, returning everything on `GET /products` means serializing 10,000 URLs for a page that only needs 20 items. Instead, the list reads only from a lightweight metadata map — no media is loaded at all. The full arrays are only fetched on `GET /products/{id}`, where one product is requested explicitly.

This is enforced at the storage layer, not just the handler. The metadata and media are stored separately in memory and only joined on detail lookups.

---

## Seed Script (optional, for load testing)

To populate the catalog with 100 products quickly (PowerShell):

```powershell
for ($i = 1; $i -le 100; $i++) {
    $body = @{
        name       = "Automated Toy Asset $i"
        sku        = "SKU-SEED-VAL-$i"
        image_urls = @("https://cdn.example.com/img_$i.jpg", "https://cdn.example.com/backup_$i.jpg")
        video_urls = @("https://cdn.example.com/video_$i.mp4")
    } | ConvertTo-Json

    Invoke-RestMethod -Uri "http://localhost:8080/products" -Method Post -ContentType "application/json" -Body $body | Out-Null
}
Write-Host "Done — 100 products created." -ForegroundColor Green
```

After running this, `GET /products?page=1&limit=20` should return quickly, pulling only summary data for 20 items while 1,000 media URLs sit untouched in memory.

---

## Project Structure

```
source-asia-backend/
├── cmd/server/main.go        # Entry point — wires everything together, starts server
├── internal/
│   ├── model/                # Struct definitions for requests, products, responses
│   ├── store/                # In-memory state — maps, slices, mutex locks
│   ├── handler/              # HTTP handlers — decode input, call store, write response
│   ├── middleware/           # Rate limit check sits here, runs before handlers
│   └── validation/           # Input validation — URLs, required fields, array limits
```

---

## Production Limitations

**This is a single-instance, in-memory service.** That's fine for the assignment, but here's what would need to change in production:

**State is lost on restart.** Everything lives in RAM. A crash or deploy wipes all products and rate-limit history.

**Can't scale horizontally.** If you run two instances behind a load balancer, each has its own state. A user could hit instance A four times and instance B four times and bypass the rate limit entirely. Rate-limit data needs to live somewhere shared.

**Mutex contention under heavy load.** A global lock works fine at low traffic. At scale, it becomes a bottleneck as every concurrent request queues up to acquire it. Approaches to fix this include lock striping (split the map into N buckets, each with its own lock), atomic counters, or offloading state to a dedicated store.

**What a real production stack would look like:**

- **PostgreSQL** for products — a `products` table for core fields and a `product_media` table with a foreign key back to it. `GET /products` runs a query on `products` only; `GET /products/{id}` joins `product_media`. The split-memory design maps cleanly to this.
- **Redis** for rate limiting — rolling window timestamps stored in a sorted set per user. Redis handles atomic operations natively, so you get consistency across multiple server instances without application-level locks.
- **CDN (e.g. S3 + CloudFront)** for media — instead of storing third-party URLs directly, the backend would accept file uploads, push them to S3, and store the resulting CDN URLs. The validation logic stays the same, just pointed at trusted internal URLs.

---

## AI Tools

Used AI assistance to help draft sections of this README and structure the seed script. Core logic — the rolling window implementation, mutex handling, storage layer design, and validation rules — was written and verified manually.
