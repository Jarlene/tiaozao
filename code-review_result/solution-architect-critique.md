# Solution Architect Critique: Phase 2 (Product System)

## Executive Summary

The proposed architecture follows the existing layered pattern and would work as a first iteration, but contains several structural weaknesses in the data models, image upload strategy, and overall extensibility. The design underutilizes existing infrastructure (Redis, MinIO capabilities) and misses architectural safeguards that will become expensive to retrofit later.

---

## 1. Data Model Assessment

### 1.1 Product.Price as float64 -- HIGH SEVERITY

```
Price float64 `gorm:"not null" json:"price"`
```

Floating-point arithmetic introduces rounding errors for monetary values. A product priced at 0.1 yuan stored and summed repeatedly will produce inexact totals.

**Recommendation**: Use `int64` (unit = cents/fen) or a `decimal` type. GORM + PostgreSQL supports `decimal(10,2)` via `gorm:"type:decimal(10,2)"` (scanner/valuer included). The frontend should divide by 100 on display.

### 1.2 Product.Status as untyped int -- MEDIUM SEVERITY

```
Status int `gorm:"default:1;not null" json:"status"`
```

Magic numbers (`1 = active`, `0 = disabled`, `2 = sold`, `3 = removed`) scattered across the codebase will cause bugs.

**Recommendation**: Define typed constants:
```go
type ProductStatus int
const (
    ProductStatusDraft     ProductStatus = 0
    ProductStatusActive    ProductStatus = 1
    ProductStatusSold      ProductStatus = 2
    ProductStatusRemoved   ProductStatus = 3
)
```
This matches the existing pattern in `User.Status` (1=active, 0=disabled) and costs zero at runtime.

### 1.3 Category tree with `gorm:"-"` Children -- MEDIUM SEVERITY

```
Children []Category `gorm:"-" json:"children"`
```

Marking `Children` as `gorm:"-"` means the service layer must build the tree manually with N+1 queries or a single full-table load + in-memory tree building. The first approach is slow for deep trees; the second works only for small category sets (<500 rows). Neither is documented.

**Recommendation**: 
- Add `ParentID` index in GORM: `gorm:"index"`
- Pre-load the entire category table on app startup or cache it in Redis (which already exists but is unused).
- Add a `Level` column (tinyint) to simplify depth-limited queries.
- The API should offer both flat list and tree endpoints.

### 1.4 Category.Name uniqueness

No unique constraint on Category.Name. Duplicate categories will appear in the UI.

**Recommendation**: `gorm:"uniqueIndex;size:100;not null"`.

### 1.5 Product.Description as unbounded text

```
Description string `gorm:"type:text" json:"description"`
```

No length validation on the service layer. A user could submit megabytes of HTML.

**Recommendation**: Add a `binding:"max=5000"` on the service-layer request struct (existing pattern: `RegisterReq.Password` has `max=72`).

---

## 2. Image Upload Strategy Assessment

### 2.1 Pre-upload each image individually before product creation

This is the most consequential architectural decision in the proposal, and the reasoning is not articulated.

**Strengths**:
- Simple frontend: upload buttons fire independently, collect URLs, include in product payload
- Image validation (size, format) happens before form submission
- Works with existing Gin JSON handlers -- no multipart form parsing changes

**Weaknesses**:

| Issue | Severity | Detail |
|---|---|---|
| Orphan images | HIGH | User uploads 5 images, abandons the form, never creates product. Files live in MinIO forever. |
| UX degradation | MEDIUM | Upload begins before user commits to the product. Slow uploads on mobile networks waste user time on a form they may not complete. |
| Race conditions | LOW | Image uploaded -- product creation fails due to validation -- image is orphaned. |
| No cleanup mechanism | HIGH | No garbage collection for orphaned images. Bucket will accumulate dead files. |
| Extra network round-trips | MEDIUM | Each image = one HTTP POST + one or both of `POST upload` / `DELETE image`. |

**Recommendation**: Use a **two-phase approach** or a **multipart single-request approach**:

- **Phase 1 upload (optional)**: Support pre-upload (current proposal) for UX with preview, but also support:
- **Phase 2 commit**: Product creation receives either pre-uploaded image IDs OR raw files in a single multipart request. Only on successful product creation are the images permanently associated.
- **Cleanup goroutine**: A background goroutine (or cron endpoint) that deletes MinIO objects whose associated ProductID was never assigned to a real product after TTL (e.g., 24 hours). This can use a `pending_images` table.

### 2.2 MinIO bucket strategy

Current config creates `flea-avatars` bucket. Product images should not share this bucket without a prefix strategy.

**Recommendation**: Use a separate `flea-products` bucket, or use prefixes within a single bucket: `avatars/{user_id}/{uuid}.jpg` and `products/{product_id}/{sort_order}.jpg`. If using a single bucket, named differently from `flea-avatars` to avoid confusion (e.g., `flea-images`).

### 2.3 Missing MinIO utility methods

The current `storage.MinIOClient` only exposes `Client()` and `Bucket()`. Every upload handler would need to replicate boilerplate.

**Recommendation**: Add to `storage.MinIOClient`:
```go
func (m *MinIOClient) UploadFile(ctx context.Context, prefix string, filename string, reader io.Reader, size int64, contentType string) (string, error)
func (m *MinIOClient) DeleteFile(ctx context.Context, objectName string) error
func (m *MinIOClient) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
```
These three methods cover every future upload need (avatars, products, messages) and avoid handler-level duplication.

### 2.4 Image access URL generation

The proposal's `ProductImage.URL` field stores the full URL as a string. If MinIO is behind an S3-compatible gateway or the endpoint changes in the future, all stored URLs become invalid.

**Recommendation**: Store only the **object key** (e.g., `products/{id}/1.jpg`) in the `URL` column. Generate the public/presigned URL at serialization time. This is future-proof for CDN insertion and bucket migration.

---

## 3. API Design Assessment

### 3.1 Route ordering for `/api/v1/products/mine`

```
GET /api/v1/products/:id           // show detail
GET /api/v1/products/mine           // my products
```

In Gin, `/mine` would be caught by `/:id` if `:id` is registered first. The `/mine` route must be registered *before* `/:id`.

**Recommendation**: Document or restructure:
```go
mine := products.Group("/mine")  // must register first
mine.GET("", productHandler.GetMyProducts)
mine.PUT("/:id/status", productHandler.UpdateMyProductStatus)

detail := products.Group("")
detail.GET("/:id", productHandler.GetProduct)
```

Or use a query parameter: `GET /api/v1/products?mine=true`.

### 3.2 PATCH vs PUT for status update

```
PUT /api/v1/products/:id/status
```

This is semantically a partial update. `PATCH` is more RESTful.

**Recommendation**: Use `PATCH /api/v1/products/:id/status` with body `{"status": 2}`.

### 3.3 Search is mentioned but not defined

The proposal says "search params" on `GET /api/v1/products` but does not define the mechanism.

**Recommendation**: Define the search approach explicitly:
- **Short-term**: GORM `Where("title ILIKE ?", "%query%")` with full-text PostgreSQL search (`tsvector`/`tsquery`)
- **Medium-term**: Add a `search_vector` column with a GIN index for `to_tsvector('simple', title || ' ' || COALESCE(description, ''))`
- This avoids introducing Elasticsearch prematurely but provides meaningful search.

### 3.4 Pagination needs a strategy

Offset-based pagination (`page`/`page_size`) is the simplest starting point but breaks under high write volume (duplicate/missing rows when rows are inserted between pages).

**Recommendation**: Start with offset pagination but design the response to support cursor pagination later:
```go
type PaginatedResponse struct {
    Items      interface{} `json:"items"`
    Total      int64       `json:"total"`
    Page       int         `json:"page"`
    PageSize   int         `json:"page_size"`
    HasMore    bool        `json:"has_more"`
    NextCursor string      `json:"next_cursor,omitempty"`
}
```

---

## 4. Missing Error Codes

The proposal adds no new error codes to `pkg/errors`. At minimum:

```go
const (
    ErrProductNotFound   = 40401  // product not found
    ErrCategoryNotFound  = 40402  // category not found
    ErrImageTooLarge     = 40004  // image exceeds size limit
    ErrImageFormat       = 40005  // unsupported image format
    ErrImageLimitExceeded = 40006 // too many images per product
    ErrProductForbidden  = 40301  // not the owner
    ErrCategoryNotEmpty  = 40007  // category has children
)
```

---

## 5. Underutilized Infrastructure

### 5.1 Redis (currently initialized, `_ = rdb` discarded)

Redis is a running dependency with zero production usage. The product system should use it for:

1. **Category tree cache**: Refresh on category write, serve from Redis on read.
2. **Product view counter**: `INCR product:views:{id}` with periodic persistence to PostgreSQL.
3. **Image upload rate limiting**: Per-user rate limit to prevent abuse of the image upload endpoint.

Not using Redis for anything in Phase 2 after the infrastructure cost of running it is a missed opportunity.

### 5.2 MinIO supports presigned URLs

If images should not be fully public, use presigned URLs with short expiry (e.g., 1 hour). This requires storing object keys rather than full URLs in the database (see section 2.4).

---

## 6. Frontend Architecture Recommendations

### 6.1 File upload component pattern

Naive UI provides `NUpload` component. The proposed pre-upload strategy would require:

```
Upload (NUpload action=/api/v1/images) -> receives URL -> store in form state
Product create POST -> sends list of URLs
```

This couples the product form to the upload API. An alternative using the two-phase approach:

```
Upload (NUpload with custom handler) -> receives image ID -> store ID in form state
Product create POST -> sends list of image IDs
Product creation service -> batch-associates images to product
```

### 6.2 Store pattern

Same pattern as `useAuthStore`. A `useProductStore` and `useCategoryStore` following the same composition API pattern is the correct approach.

---

## 7. Scalability and Maintainability

| Concern | Assessment |
|---|---|
| Product listing with images | Potential N+1 query problem. Must use `Preload("Images")` or a batch image fetch. |
| Category tree | Full-table load every request is fine for <500 categories. Beyond that, caching is mandatory. |
| Image cleanup | No mechanism proposed. Without it, MinIO bucket grows unbounded. |
| Search | Undefined. Post-MVP search refactors are expensive. |
| Caching strategy | Zero. Redis sits idle. |
| Concurrent operations | No row-level locking or optimistic concurrency for product updates. |

---

## 8. Verification Questions and Answers

**Q1: Does the design correctly handle the orphan image problem?**

A1: No. Pre-upload without a cleanup mechanism creates unbounded storage waste. This is the single highest-severity issue in the proposal.

**Q2: Is the data model future-proof for Phase 3+ (transactions, messaging, reviews)?**

A2: Partially. The `UserID` foreign key on Product is correct and aligns with existing patterns. However, `Price float64` will cause problems when Phase 4 needs aggregation/sum queries. The `Status int` without typed constants will cause confusion when Phase 3 adds negotiation states.

**Q3: Does the design properly leverage existing infrastructure?**

A3: No. Redis is completely unused despite being a running dependency. The MinIO client wrapper is minimal and forces each handler to duplicate upload logic. The MinIO bucket is misconfigured (shared with avatars, no prefix strategy).

**Q4: Is the search strategy sufficiently defined?**

A4: No. "Search params" is too vague. The design should commit to at least a PostgreSQL full-text search approach with the appropriate schema (GIN indexes, tsvector column) even if it defers the frontend implementation.

**Q5: Is the API consistent with Phase 1's patterns?**

A5: Yes. The route structure, response format, and auth middleware usage all follow Phase 1 conventions. This is a clear strength of the proposal.

---

## 9. Final Score

**Solution Optimality Score: 6.5/10**

The proposal correctly follows the existing architectural conventions but has several structural gaps that will cause rework:

- Price as float64 needs changing before any financial reporting is built
- Image upload strategy needs a cleanup mechanism or a two-phase commit approach
- Redis should be used for at least category caching
- Search needs a concrete implementation plan
- MinIO utility methods need to be added to avoid handler-level duplication

The core architecture (layered Go + Gin + GORM, layered Vue 3 + Pinia) is sound. The issues are in the details of the data model and the upload strategy -- both are fixable within the same architectural approach.

---

*Generated using Solution Architect evaluation pattern*
