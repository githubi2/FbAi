package services

import (
	"context"
	"errors"
	"time"

	"github.com/githubi2/FbAi/art-design-server/db"
	"github.com/githubi2/FbAi/art-design-server/models"
	"github.com/jackc/pgx/v5"
)

// RoleService 角色服务（PostgreSQL 实现）
type RoleService struct{}

var DefaultRoleService = &RoleService{}

// List 获取所有角色列表
func (s *RoleService) List() []models.Role {
	if db.Pool == nil {
		return s.listFallback()
	}

	ctx := context.Background()
	querySQL := `SELECT id, role_name, role_code, description, menu_ids, status, created_at, updated_at 
		FROM roles ORDER BY id ASC`

	rows, err := db.Pool.Query(ctx, querySQL)
	if err != nil {
		return s.listFallback()
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var r models.Role
		if err := rows.Scan(&r.ID, &r.RoleName, &r.RoleCode, &r.Description,
			&r.MenuIDs, &r.Status, &r.CreatedAt, &r.UpdatedAt); err != nil {
			continue
		}
		roles = append(roles, r)
	}
	if roles == nil {
		roles = []models.Role{}
	}
	return roles
}

// GetByID 按ID获取角色
func (s *RoleService) GetByID(id uint) (*models.Role, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	querySQL := `SELECT id, role_name, role_code, description, menu_ids, status, created_at, updated_at 
		FROM roles WHERE id = $1`

	var r models.Role
	err := db.Pool.QueryRow(ctx, querySQL, id).Scan(
		&r.ID, &r.RoleName, &r.RoleCode, &r.Description,
		&r.MenuIDs, &r.Status, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("角色不存在")
	}
	return &r, nil
}

// GetByCode 按角色编码获取角色
func (s *RoleService) GetByCode(code string) (*models.Role, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	querySQL := `SELECT id, role_name, role_code, description, menu_ids, status, created_at, updated_at 
		FROM roles WHERE role_code = $1`

	var r models.Role
	err := db.Pool.QueryRow(ctx, querySQL, code).Scan(
		&r.ID, &r.RoleName, &r.RoleCode, &r.Description,
		&r.MenuIDs, &r.Status, &r.CreatedAt, &r.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("角色不存在")
	}
	return &r, nil
}

// Create 创建角色
func (s *RoleService) Create(req models.CreateRoleRequest) (*models.Role, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	// 检查角色编码是否已存在
	var exists int
	_ = db.Pool.QueryRow(ctx, `SELECT 1 FROM roles WHERE role_code = $1`, req.RoleCode).Scan(&exists)
	if exists == 1 {
		return nil, errors.New("角色编码已存在")
	}

	menuIDs := req.MenuIDs
	if menuIDs == nil {
		menuIDs = []int64{}
	}

	querySQL := `INSERT INTO roles (role_name, role_code, description, menu_ids, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	var id uint
	err := db.Pool.QueryRow(ctx, querySQL,
		req.RoleName, req.RoleCode, req.Description, menuIDs, req.Status, now, now,
	).Scan(&id)
	if err != nil {
		return nil, errors.New("创建角色失败: " + err.Error())
	}

	return &models.Role{
		ID:          id,
		RoleName:    req.RoleName,
		RoleCode:    req.RoleCode,
		Description: req.Description,
		MenuIDs:     menuIDs,
		Status:      req.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update 更新角色
func (s *RoleService) Update(req models.UpdateRoleRequest) (*models.Role, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	// 检查角色是否存在
	_, err := s.GetByID(req.ID)
	if err != nil {
		return nil, err
	}

	menuIDs := req.MenuIDs
	if menuIDs == nil {
		menuIDs = []int64{}
	}

	querySQL := `UPDATE roles SET role_name=$1, role_code=$2, description=$3, menu_ids=$4, status=$5, updated_at=$6
		WHERE id=$7`

	_, err = db.Pool.Exec(ctx, querySQL,
		req.RoleName, req.RoleCode, req.Description, menuIDs, req.Status, now, req.ID,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, errors.New("角色编码已存在")
		}
		return nil, errors.New("更新角色失败")
	}

	return s.GetByID(req.ID)
}

// Delete 删除角色
func (s *RoleService) Delete(id uint) error {
	if db.Pool == nil {
		return errors.New("数据库未连接")
	}

	ctx := context.Background()
	result, err := db.Pool.Exec(ctx, `DELETE FROM roles WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("角色不存在")
	}
	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, pgx.ErrNoRows) || errors.Is(err, context.DeadlineExceeded)
}

// --- 内存 fallback（数据库不可用时使用）---

var fallbackRoles = map[uint]*models.Role{
	1: {ID: 1, RoleName: "超级管理员", RoleCode: "R_SUPER", Description: "拥有所有权限", Status: 1, MenuIDs: []int64{}},
	2: {ID: 2, RoleName: "管理员", RoleCode: "R_ADMIN", Description: "管理员权限", Status: 1, MenuIDs: []int64{}},
	3: {ID: 3, RoleName: "普通用户", RoleCode: "R_USER", Description: "普通用户权限", Status: 1, MenuIDs: []int64{}},
}

func (s *RoleService) listFallback() []models.Role {
	result := make([]models.Role, 0, len(fallbackRoles))
	for _, r := range fallbackRoles {
		result = append(result, *r)
	}
	return result
}
