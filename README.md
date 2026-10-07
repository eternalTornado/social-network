# GopherSocial API

REST API cho một mạng xã hội nhỏ, viết bằng Go và PostgreSQL. Dự án đang trong quá trình xây dựng: hiện có health check, tạo bài viết và xem bài viết theo id.

## Công nghệ

| Thành phần | Dùng gì |
| --- | --- |
| Ngôn ngữ | Go 1.25 |
| Router | [chi v5](https://github.com/go-chi/chi) |
| Database | PostgreSQL 16 (chạy bằng Docker Compose) |
| Driver | [lib/pq](https://github.com/lib/pq) qua `database/sql` |
| Migration | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Live reload | [air](https://github.com/air-verse/air) |
| Đọc `.env` | [godotenv](https://github.com/joho/godotenv) |

## Cần cài trước

- **Go 1.25+**
- **Docker Desktop** (để chạy Postgres)
- **golang-migrate**: `brew install golang-migrate`
- **air**: `go install github.com/air-verse/air@latest`
  (đừng dùng `brew install air`, formula đó là một công cụ format code R khác tên trùng). Nhớ thêm `$(go env GOPATH)/bin` vào `PATH`.

## Chạy thử

```bash
git clone https://github.com/eternalTornado/social-network.git
cd social-network

cp .env.example .env   # cấu hình local, đã có sẵn giá trị khớp docker-compose.yml
make db-up             # bật Postgres và đợi tới khi nhận kết nối
make migrate-up        # tạo bảng
make dev               # chạy API với live reload (hoặc: make run)
```

Kiểm tra server:

```bash
curl localhost:8080/v1/health
```

```json
{"env":"development","status":"ok","version":"0.0.1"}
```

## API

Mọi route nằm dưới tiền tố `/v1`.

| Method | Path | Mô tả |
| --- | --- | --- |
| `GET` | `/v1/health` | Trạng thái server, môi trường và version |
| `POST` | `/v1/posts` | Tạo bài viết |
| `GET` | `/v1/posts/{postID}` | Lấy một bài viết theo id |

### Tạo bài viết

```bash
curl -X POST localhost:8080/v1/posts \
  -H 'Content-Type: application/json' \
  -d '{"title":"Hello","content":"Bài viết đầu tiên","tags":["go","intro"]}'
```

Body chỉ nhận ba field `title`, `content`, `tags`; field lạ hoặc body lớn hơn 1MB sẽ bị trả `400`. Chưa có đăng nhập nên bài viết được gán cho một user cố định.

### Lấy bài viết

```bash
curl localhost:8080/v1/posts/1
```

```json
{
  "id": 1,
  "content": "Bài viết đầu tiên",
  "title": "Hello",
  "user_id": "1",
  "tags": ["go", "intro"],
  "created_at": "...",
  "updated_at": "..."
}
```

Id không tồn tại trả `404`, id không phải số trả `400`.

### Định dạng lỗi

Mọi lỗi đều trả JSON dạng:

```json
{"Error": "not found"}
```

## Cấu hình

Server đọc biến môi trường, và tự nạp file `.env` ở thư mục gốc nếu có.

| Biến | Mặc định | Ý nghĩa |
| --- | --- | --- |
| `ADDR` | `:8080` | Địa chỉ server lắng nghe |
| `DB_ADDR` | — | Chuỗi kết nối Postgres, ví dụ `postgres://admin:hotdog123@localhost:5432/socialnetwork?sslmode=disable` |
| `DB_MAX_OPEN_CONNS` | `30` | Số kết nối tối đa tới DB |
| `DB_MAX_IDLE_CONNS` | `30` | Số kết nối rảnh được giữ lại |
| `DB_MAX_IDLE_TIME` | `15m` | Thời gian tối đa một kết nối được rảnh trước khi bị đóng |
| `ENV` | `development` | Tên môi trường, hiện trong `/v1/health` |

`DB_ADDR` phải được đặt (qua `.env` hoặc biến môi trường), giá trị mặc định trong code không kết nối được. `make migrate-*` cũng dùng `DB_ADDR` này.

## Lệnh Make

Chạy `make` (không tham số) để xem toàn bộ danh sách.

| Lệnh | Việc làm |
| --- | --- |
| `make dev` | Chạy API với air, tự build lại khi sửa code |
| `make run` | Chạy API một lần |
| `make build` | Build binary ra `./bin/main` |
| `make test` / `make fmt` / `make vet` | Test, format, soát lỗi |
| `make db-up` / `make db-down` | Bật / tắt Postgres (giữ dữ liệu) |
| `make db-reset` | **Xoá hết dữ liệu**, bật lại DB trắng rồi chạy migration |
| `make psql` | Mở `psql` bên trong container |
| `make migration <tên>` | Tạo cặp file migration mới, ví dụ `make migration add_comments` |
| `make migrate-up` | Chạy mọi migration chưa áp dụng |
| `make migrate-down <n>` | Undo `n` migration gần nhất |
| `make migrate-version` | Xem version hiện tại của DB |
| `make migrate-force <v>` | Gỡ trạng thái dirty sau khi migration lỗi giữa chừng |

## Database

Hai bảng, tạo bởi các migration trong [cmd/migrate/migrations](cmd/migrate/migrations):

- **users**: `id`, `email` (citext, unique), `username` (unique), `password` (bytea), `created_at`
- **posts**: `id`, `title`, `content`, `user_id` (khoá ngoại tới `users.id`), `tags` (mảng `varchar(100)`), `created_at`, `updated_at`

## Cấu trúc thư mục

```
cmd/
  api/                 # HTTP server: main, router, handler, helper JSON
  migrate/migrations/  # File SQL up/down cho golang-migrate
internal/
  db/                  # Mở connection pool tới Postgres
  env/                 # Đọc biến môi trường có giá trị mặc định
  store/               # Tầng truy cập dữ liệu (repository) cho posts, users
docker-compose.yml     # Postgres 16 cho môi trường local
Makefile               # Lệnh tắt cho dev, DB và migration
.air.toml              # Cấu hình live reload
```

Handler trong `cmd/api` không gọi SQL trực tiếp mà đi qua `store.Storage`, một struct gom các interface `Posts` và `Users`. Nhờ vậy tầng HTTP không phụ thuộc vào cách lưu trữ cụ thể.
