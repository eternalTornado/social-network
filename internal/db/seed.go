package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"

	"github.com/truongndv/social/internal/store"
)

var usernames = []string{
	"alice", "bob", "carol", "dave", "eve",
	"frank", "grace", "heidi", "ivan", "judy",
	"mallory", "nathan", "olivia", "peggy", "quentin",
}

var titles = []string{
	"Học Go trong 30 ngày", "Vì sao nên dùng context.Context", "Postgres index cơ bản",
	"chi router vs net/http", "Docker compose cho dev local", "JWT là gì",
	"Clean architecture cho REST API", "Rate limiting đơn giản", "Structured logging với zap",
	"Testing handler HTTP trong Go",
}

var contents = []string{
	"Bài viết chia sẻ kinh nghiệm thực tế khi làm dự án Go.",
	"Một vài lưu ý nhỏ nhưng quan trọng khi thiết kế API.",
	"Tổng hợp lỗi hay gặp và cách xử lý.",
	"So sánh ưu nhược điểm giữa các cách tiếp cận.",
	"Ghi chú nhanh sau khi đọc tài liệu chính thức.",
	"Ví dụ code tối giản, dễ chạy thử.",
	"Checklist trước khi deploy production.",
	"Giải thích từng bước, kèm sơ đồ.",
	"Bài học rút ra sau một buổi debug dài.",
	"Tips tối ưu performance cho service Go.",
}

var tagsPool = []string{
	"go", "postgres", "docker", "api", "backend",
	"testing", "chi", "sql", "auth", "devops",
}

var commentContents = []string{
	"Bài viết hay quá, cảm ơn bạn!",
	"Mình cũng gặp lỗi này, giờ mới hiểu.",
	"Có thể nói rõ hơn phần này không?",
	"Đã áp dụng và chạy được, cảm ơn.",
	"Theo mình nên thêm ví dụ thực tế.",
	"Chỗ này hình như thiếu index thì phải.",
	"Ủng hộ bạn viết thêm phần tiếp theo.",
	"Đơn giản mà hiệu quả.",
	"Mình thắc mắc về timeout ở đây.",
	"Rất chi tiết, dễ hiểu.",
}

func Seed(store store.Storage) {
	ctx := context.Background()

	users := generateUsers(100)
	for _, user := range users {
		if err := store.Users.Create(ctx, user); err != nil {
			log.Println("Lỗi tạo user:", err)
			return
		}
	}

	posts := generatePosts(200, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("lỗi tạo post:", err)
			return
		}
	}

	comments := generateComments(500, users, posts)
	for _, comment := range comments {
		if err := store.Comments.Create(ctx, comment); err != nil {
			log.Println("lỗi tạo comment:", err)
			return
		}
	}

	log.Println("seeding completed")
}

func generateUsers(n int) []*store.User {
	users := make([]*store.User, n)

	for i := 0; i < n; i++ {
		name := fmt.Sprintf("%s%d", usernames[i%len(usernames)], i)
		users[i] = &store.User{
			Username: name,
			Email:    name + "@example.com",
			Password: "123123",
		}
	}

	return users
}

func generatePosts(n int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, n)

	for i := 0; i < n; i++ {
		user := users[rand.Intn(len(users))]

		posts[i] = &store.Post{
			Title:   titles[rand.Intn(len(titles))],
			Content: contents[rand.Intn(len(contents))],
			UserID:  user.ID,
			Tags:    randomTags(),
		}
	}

	return posts
}

func randomTags() []string {
	n := rand.Intn(3) + 1

	tags := make([]string, n)
	for i := 0; i < n; i++ {
		tags[i] = tagsPool[rand.Intn(len(tagsPool))]
	}

	return tags
}

func generateComments(n int, users []*store.User, posts []*store.Post) []*store.Comment {
	comments := make([]*store.Comment, n)

	for i := 0; i < n; i++ {
		comments[i] = &store.Comment{
			PostID:  posts[rand.Intn(len(posts))].ID,
			UserID:  users[rand.Intn(len(users))].ID,
			Content: commentContents[rand.Intn(len(commentContents))],
		}
	}

	return comments
}
