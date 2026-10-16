package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Version   int       `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int `json:"comments_count"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	query := `
		INSERT INTO posts (content, title, user_id, tags)
		VALUES($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := s.db.QueryRowContext(ctx, query, post.Content, post.Title, post.UserID, pq.Array(post.Tags)).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, user_id, title, content, created_at, updated_at, tags, version
		FROM posts
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		&post.CreatedAt,
		&post.UpdatedAt,
		pq.Array(&post.Tags),
		&post.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2, version = version + 1, updated_at = NOW()
		WHERE id = $3 AND version = $4
		RETURNING version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			//Nếu sửa cùng 1 post, user khác sửa trước thì xem như post đã thành post khác
			//không tìm thấy post hiện tại nữa
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *PostStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]PostWithMetadata, error) {
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	sort := "DESC"
	if fq.Sort == "asc" {
		sort = "ASC"
	}

	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags,
		       u.username, COUNT(c.id) AS comments_count
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		WHERE (p.user_id = $1 OR p.user_id IN (SELECT f.user_id FROM followers f WHERE f.follower_id = $1))
		  AND (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
		  AND (p.tags @> $5 OR $5 = '{}')
		  AND ($6::timestamptz IS NULL OR p.created_at >= $6::timestamptz)
		  AND ($7::timestamptz IS NULL OR p.created_at <= $7::timestamptz)
		GROUP BY p.id, u.username
		ORDER BY p.created_at ` + sort + `
		LIMIT $2 OFFSET $3
	`

	tags := fq.Tags
	if tags == nil {
		tags = []string{}
	}

	var since, until any
	if fq.Since != "" {
		since = fq.Since
	}
	if fq.Until != "" {
		until = fq.Until
	}

	rows, err := s.db.QueryContext(
		ctx,
		query,
		userID,
		fq.Limit,
		fq.Offset,
		fq.Search,
		pq.Array(tags),
		since,
		until,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	feed := []PostWithMetadata{}
	for rows.Next() {
		var p PostWithMetadata

		if err := rows.Scan(
			&p.ID,
			&p.UserID,
			&p.Title,
			&p.Content,
			&p.CreatedAt,
			&p.Version,
			pq.Array(&p.Tags),
			&p.User.Username,
			&p.CommentsCount,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}

		feed = append(feed, p)
	}

	//rows.Err có value nếu rows.Next() phát sinh lỗi
	//nên việc check được để sau
	//rows.Next trả false trong 2 case: hết element, hoặc bị lỗi
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feed, nil
}
