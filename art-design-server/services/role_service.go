package services

import (
	"errors"
	"sync"
	"time"

	"github.com/githubi2/FbAi/art-design-server/models"
)

// RoleService 角色服务
type RoleService struct {
	mu    sync.RWMutex
	roles map[uint]*models.Role
	seq   uint
}

var DefaultRoleService = NewRoleService()

func NewRoleService() *RoleService {
	svc := &RoleService{
		roles: make(map[uint]*models.Role),
		seq:   2,
	}
	svc.roles[1] = &models.Role{
		ID: 1, Name: "超级管理员", Code: "R_SUPER",
		Desc: "拥有所有权限", Status: 1, MenuIDs: []uint{1, 2, 3, 4, 5, 6},
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	svc.roles[2] = &models.Role{
		ID: 2, Name: "普通用户", Code: "R_USER",
		Desc: "基础权限", Status: 1, MenuIDs: []uint{1},
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	return svc
}

// List 获取角色列表
func (s *RoleService) List() []models.Role {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Role, 0, len(s.roles))
	for _, r := range s.roles {
		result = append(result, *r)
	}
	return result
}

// GetByID 按ID获取角色
func (s *RoleService) GetByID(id uint) (*models.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, ok := s.roles[id]
	if !ok {
		return nil, errors.New("角色不存在")
	}
	return r, nil
}

// Create 创建角色
func (s *RoleService) Create(req models.CreateRoleRequest) (*models.Role, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, r := range s.roles {
		if r.Code == req.Code {
			return nil, errors.New("角色编码已存在")
		}
	}

	s.seq++
	now := time.Now()
	role := &models.Role{
		ID:        s.seq,
		Name:      req.Name,
		Code:      req.Code,
		Desc:      req.Desc,
		Status:    req.Status,
		MenuIDs:   req.MenuIDs,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.roles[s.seq] = role
	return role, nil
}

// Update 更新角色
func (s *RoleService) Update(req models.UpdateRoleRequest) (*models.Role, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	r, ok := s.roles[req.ID]
	if !ok {
		return nil, errors.New("角色不存在")
	}

	r.Name = req.Name
	r.Code = req.Code
	r.Desc = req.Desc
	r.Status = req.Status
	r.MenuIDs = req.MenuIDs
	r.UpdatedAt = time.Now()

	return r, nil
}

// Delete 删除角色
func (s *RoleService) Delete(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.roles[id]; !ok {
		return errors.New("角色不存在")
	}
	delete(s.roles, id)
	return nil
}
