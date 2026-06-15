package services

import (
	"context"
	"errors"
	"time"

	"github.com/githubi2/FbAi/art-design-server/crypto"
	"github.com/githubi2/FbAi/art-design-server/db"
	"github.com/githubi2/FbAi/art-design-server/models"
)

// UserService 用户服务（PostgreSQL 实现）
type UserService struct{}

var DefaultUserService = &UserService{}

// List 分页查询用户列表
func (s *UserService) List(page, size int, keyword string) ([]models.User, int64) {
	if db.Pool == nil {
		return s.listFallback(page, size, keyword)
	}

	ctx := context.Background()

	// 查询总数
	var total int64
	countSQL := `SELECT COUNT(*) FROM users WHERE ($1 = '' OR user_name ILIKE '%' || $1 || '%' OR nick_name ILIKE '%' || $1 || '%')`
	_ = db.Pool.QueryRow(ctx, countSQL, keyword).Scan(&total)

	// 分页查询
	offset := (page - 1) * size
	querySQL := `
		SELECT id, user_name, nick_name, email, phone, avatar, status, role_id, role_name, created_at, updated_at
		FROM users
		WHERE ($1 = '' OR user_name ILIKE '%' || $1 || '%' OR nick_name ILIKE '%' || $1 || '%')
		ORDER BY id ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := db.Pool.Query(ctx, querySQL, keyword, size, offset)
	if err != nil {
		return []models.User{}, 0
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.UserName, &u.NickName, &u.Email, &u.Phone,
			&u.Avatar, &u.Status, &u.RoleID, &u.RoleName, &u.CreatedAt, &u.UpdatedAt); err != nil {
			continue
		}
		u.Password = "" // 不返回密码
		users = append(users, u)
	}
	if users == nil {
		users = []models.User{}
	}

	return users, total
}

// GetByID 按ID获取用户
func (s *UserService) GetByID(id uint) (*models.User, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	querySQL := `
		SELECT id, user_name, nick_name, email, phone, avatar, status, role_id, role_name, created_at, updated_at
		FROM users WHERE id = $1
	`
	var u models.User
	err := db.Pool.QueryRow(ctx, querySQL, id).Scan(
		&u.ID, &u.UserName, &u.NickName, &u.Email, &u.Phone,
		&u.Avatar, &u.Status, &u.RoleID, &u.RoleName, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("用户不存在")
	}
	u.Password = "" // 不返回密码
	return &u, nil
}

// Create 创建用户（密码自动 bcrypt 哈希）
func (s *UserService) Create(req models.CreateUserRequest) (*models.User, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	// 哈希密码
	hashedPassword, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("密码加密失败")
	}

	querySQL := `
		INSERT INTO users (user_name, password, nick_name, email, phone, avatar, status, role_id, role_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, '', $9, $10)
		RETURNING id
	`
	var id uint
	err = db.Pool.QueryRow(ctx, querySQL,
		req.UserName, hashedPassword, req.NickName, req.Email, req.Phone,
		req.Avatar, req.Status, req.RoleID, now, now,
	).Scan(&id)
	if err != nil {
		return nil, errors.New("创建用户失败: " + err.Error())
	}

	return &models.User{
		ID:        id,
		UserName:  req.UserName,
		NickName:  req.NickName,
		Email:     req.Email,
		Phone:     req.Phone,
		Avatar:    req.Avatar,
		Status:    req.Status,
		RoleID:    req.RoleID,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update 更新用户
func (s *UserService) Update(req models.UpdateUserRequest) (*models.User, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	querySQL := `
		UPDATE users SET nick_name=$1, email=$2, phone=$3, avatar=$4, status=$5, role_id=$6, updated_at=$7
		WHERE id=$8
	`
	_, err := db.Pool.Exec(ctx, querySQL, req.NickName, req.Email, req.Phone, req.Avatar, req.Status, req.RoleID, now, req.ID)
	if err != nil {
		return nil, errors.New("更新用户失败")
	}

	return s.GetByID(req.ID)
}

// Delete 删除用户
func (s *UserService) Delete(id uint) error {
	if db.Pool == nil {
		return errors.New("数据库未连接")
	}

	ctx := context.Background()
	result, err := db.Pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("用户不存在")
	}
	return nil
}

// GetAuthInfo 获取用户认证信息（用于登录）
func (s *UserService) GetAuthInfo(userName string) (uint, string, string, bool, error) {
	if db.Pool == nil {
		return 0, "", "", false, errors.New("数据库未连接")
	}

	ctx := context.Background()
	querySQL := `SELECT id, password, role_name, status FROM users WHERE user_name = $1`

	var id uint
	var password, roleName string
	var status int
	err := db.Pool.QueryRow(ctx, querySQL, userName).Scan(&id, &password, &roleName, &status)
	if err != nil {
		return 0, "", "", false, errors.New("用户不存在")
	}

	return id, password, roleName, status == 1, nil
}

// GetPasswordHash 获取用户密码哈希（用于登录验证）
func (s *UserService) GetPasswordHash(userName string) (uint, string, string, error) {
	if db.Pool == nil {
		return 0, "", "", errors.New("数据库未连接")
	}

	ctx := context.Background()
	querySQL := `SELECT id, password, role_name FROM users WHERE user_name = $1 AND status = 1`

	var id uint
	var password, roleName string
	err := db.Pool.QueryRow(ctx, querySQL, userName).Scan(&id, &password, &roleName)
	if err != nil {
		return 0, "", "", errors.New("用户不存在或已被禁用")
	}

	return id, password, roleName, nil
}

// --- 内存 fallback（数据库不可用时使用）---

var fallbackUsers = map[uint]*models.User{
	1: {ID: 1, UserName: "admin", NickName: "管理员", Status: 1, RoleName: "超级管理员"},
}

func (s *UserService) listFallback(page, size int, keyword string) ([]models.User, int64) {
	var result []models.User
	for _, u := range fallbackUsers {
		if keyword != "" && u.UserName != keyword && u.NickName != keyword {
			continue
		}
		result = append(result, *u)
	}
	total := int64(len(result))
	start := (page - 1) * size
	if start < 0 {
		start = 0
	}
	if start >= len(result) {
		return []models.User{}, total
	}
	end := start + size
	if end > len(result) {
		end = len(result)
	}
	return result[start:end], total
}
