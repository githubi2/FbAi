package services

import (
	"errors"
	"sync"
	"time"

	"github.com/githubi2/FbAi/art-design-server/models"
)

// UserService 用户服务
type UserService struct {
	mu    sync.RWMutex
	users map[uint]*models.User
	seq   uint
}

var DefaultUserService = NewUserService()

func NewUserService() *UserService {
	svc := &UserService{
		users: make(map[uint]*models.User),
		seq:   2,
	}
	// 初始化默认用户
	svc.users[1] = &models.User{
		ID:        1,
		UserName:  "admin",
		Password:  "",
		NickName:  "超级管理员",
		Email:     "admin@art-design.com",
		Phone:     "13800000001",
		Avatar:    "",
		Status:    1,
		RoleID:    1,
		RoleName:  "超级管理员",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	svc.users[2] = &models.User{
		ID:        2,
		UserName:  "user",
		Password:  "",
		NickName:  "普通用户",
		Email:     "user@art-design.com",
		Phone:     "13800000002",
		Avatar:    "",
		Status:    1,
		RoleID:    2,
		RoleName:  "普通用户",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return svc
}

// List 分页查询用户列表
func (s *UserService) List(page, size int, keyword string) ([]models.User, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.User
	for _, u := range s.users {
		if keyword != "" && !contains(u.UserName, keyword) && !contains(u.NickName, keyword) {
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

// GetByID 按ID获取用户
func (s *UserService) GetByID(id uint) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, ok := s.users[id]
	if !ok {
		return nil, errors.New("用户不存在")
	}
	return u, nil
}

// Create 创建用户
func (s *UserService) Create(req models.CreateUserRequest) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查用户名是否已存在
	for _, u := range s.users {
		if u.UserName == req.UserName {
			return nil, errors.New("用户名已存在")
		}
	}

	s.seq++
	now := time.Now()
	user := &models.User{
		ID:        s.seq,
		UserName:  req.UserName,
		Password:  "",
		NickName:  req.NickName,
		Email:     req.Email,
		Phone:     req.Phone,
		Avatar:    req.Avatar,
		Status:    req.Status,
		RoleID:    req.RoleID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.users[s.seq] = user
	return user, nil
}

// Update 更新用户
func (s *UserService) Update(req models.UpdateUserRequest) (*models.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[req.ID]
	if !ok {
		return nil, errors.New("用户不存在")
	}

	if req.NickName != "" {
		u.NickName = req.NickName
	}
	if req.Email != "" {
		u.Email = req.Email
	}
	if req.Phone != "" {
		u.Phone = req.Phone
	}
	if req.Avatar != "" {
		u.Avatar = req.Avatar
	}
	u.Status = req.Status
	u.RoleID = req.RoleID
	u.UpdatedAt = time.Now()

	return u, nil
}

// Delete 删除用户
func (s *UserService) Delete(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.users[id]; !ok {
		return errors.New("用户不存在")
	}
	delete(s.users, id)
	return nil
}

// GetAuthInfo 获取用户认证信息（用于登录）
func (s *UserService) GetAuthInfo(userName string) (uint, string, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.UserName == userName {
			return u.ID, u.RoleName, u.Status == 1, nil
		}
	}
	return 0, "", false, errors.New("用户不存在")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
