package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/githubi2/FbAi/art-design-server/crypto"
	"github.com/githubi2/FbAi/art-design-server/db"
	"github.com/githubi2/FbAi/art-design-server/models"
)

// TenantService 租户服务
type TenantService struct{}

var DefaultTenantService = &TenantService{}

// List 获取所有租户列表
func (s *TenantService) List() []models.Tenant {
	if db.Pool == nil {
		return []models.Tenant{}
	}

	ctx := context.Background()
	querySQL := `SELECT id, name, code, status, contact_name, contact_phone, contact_email, description, created_at, updated_at
		FROM tenants ORDER BY id ASC`

	rows, err := db.Pool.Query(ctx, querySQL)
	if err != nil {
		return []models.Tenant{}
	}
	defer rows.Close()

	var tenants []models.Tenant
	for rows.Next() {
		var t models.Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Code, &t.Status, &t.ContactName,
			&t.ContactPhone, &t.ContactEmail, &t.Description, &t.CreatedAt, &t.UpdatedAt); err != nil {
			continue
		}
		tenants = append(tenants, t)
	}
	if tenants == nil {
		tenants = []models.Tenant{}
	}
	return tenants
}

// GetByID 按ID获取租户
func (s *TenantService) GetByID(id uint) (*models.Tenant, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	querySQL := `SELECT id, name, code, status, contact_name, contact_phone, contact_email, description, created_at, updated_at
		FROM tenants WHERE id = $1`

	var t models.Tenant
	err := db.Pool.QueryRow(ctx, querySQL, id).Scan(
		&t.ID, &t.Name, &t.Code, &t.Status, &t.ContactName,
		&t.ContactPhone, &t.ContactEmail, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("租户不存在")
	}
	return &t, nil
}

// GetByCode 按编码获取租户
func (s *TenantService) GetByCode(code string) (*models.Tenant, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	querySQL := `SELECT id, name, code, status, contact_name, contact_phone, contact_email, description, created_at, updated_at
		FROM tenants WHERE code = $1`

	var t models.Tenant
	err := db.Pool.QueryRow(ctx, querySQL, code).Scan(
		&t.ID, &t.Name, &t.Code, &t.Status, &t.ContactName,
		&t.ContactPhone, &t.ContactEmail, &t.Description, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("租户不存在")
	}
	return &t, nil
}

// Create 创建租户（事务：创建租户 + 创建角色 + 创建管理员账号）
func (s *TenantService) Create(req models.CreateTenantRequest) (*models.Tenant, *models.User, error) {
	if db.Pool == nil {
		return nil, nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	// 开启事务
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, nil, errors.New("开启事务失败")
	}
	defer tx.Rollback(ctx)

	// 1. 创建租户
	var tenantID uint
	err = tx.QueryRow(ctx,
		`INSERT INTO tenants (name, code, status, contact_name, contact_phone, contact_email, description, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		req.Name, req.Code, 1, req.ContactName, req.ContactPhone, req.ContactEmail, req.Description, now, now,
	).Scan(&tenantID)
	if err != nil {
		return nil, nil, errors.New("创建租户失败: " + err.Error())
	}

	// 2. 创建租户管理员角色（role_code 包含租户ID避免冲突）
	// 分配菜单: Dashboard(1), System(2), Console(3), User(4)
	adminRoleCode := fmt.Sprintf("T%d_R_ADMIN", tenantID)
	adminMenuIDs := "{1,2,3,4}"
	var adminRoleID uint
	err = tx.QueryRow(ctx,
		`INSERT INTO roles (role_name, role_code, description, menu_ids, status, tenant_id, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		"租户管理员", adminRoleCode, "租户管理员角色", adminMenuIDs, 1, tenantID, now, now,
	).Scan(&adminRoleID)
	if err != nil {
		return nil, nil, errors.New("创建租户管理员角色失败: " + err.Error())
	}

	// 3. 创建普通用户角色
	// 分配菜单: Dashboard(1), Console(3)
	userRoleCode := fmt.Sprintf("T%d_R_USER", tenantID)
	userMenuIDs := "{1,3}"
	var userRoleID uint
	err = tx.QueryRow(ctx,
		`INSERT INTO roles (role_name, role_code, description, menu_ids, status, tenant_id, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		"普通用户", userRoleCode, "普通用户角色", userMenuIDs, 1, tenantID, now, now,
	).Scan(&userRoleID)
	if err != nil {
		return nil, nil, errors.New("创建普通用户角色失败: " + err.Error())
	}

	// 4. 为租户管理员角色分配权限
	_, err = tx.Exec(ctx,
		`INSERT INTO role_permissions (role_id, permission_id)
		 SELECT $1, id FROM permissions WHERE code IN (
		 	'system:user:list','system:user:create','system:user:edit','system:user:delete',
		 	'system:role:list','system:role:create','system:role:edit','system:role:delete',
		 	'dashboard:view','dashboard:export'
		 )`,
		adminRoleID,
	)
	if err != nil {
		return nil, nil, errors.New("分配租户管理员权限失败: " + err.Error())
	}

	// 5. 为普通用户角色分配权限
	_, err = tx.Exec(ctx,
		`INSERT INTO role_permissions (role_id, permission_id)
		 SELECT $1, id FROM permissions WHERE code IN ('dashboard:view')`,
		userRoleID,
	)
	if err != nil {
		return nil, nil, errors.New("分配普通用户权限失败: " + err.Error())
	}

	// 6. 创建租户管理员账号
	hashedPassword, hashErr := crypto.HashPassword(req.AdminPassword)
	if hashErr != nil {
		return nil, nil, errors.New("密码加密失败")
	}

	var adminUserID uint
	err = tx.QueryRow(ctx,
		`INSERT INTO users (user_name, password, nick_name, status, role_id, role_name, tenant_id, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		req.AdminUserName, hashedPassword, req.AdminNickName, 1, adminRoleID, "租户管理员", tenantID, now, now,
	).Scan(&adminUserID)
	if err != nil {
		return nil, nil, errors.New("创建租户管理员账号失败: " + err.Error())
	}

	// 提交事务
	if err := tx.Commit(ctx); err != nil {
		return nil, nil, errors.New("提交事务失败")
	}

	tenant := &models.Tenant{
		ID:           tenantID,
		Name:         req.Name,
		Code:         req.Code,
		Status:       1,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		Description:  req.Description,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	adminUser := &models.User{
		ID:        adminUserID,
		UserName:  req.AdminUserName,
		NickName:  req.AdminNickName,
		Status:    1,
		RoleID:    adminRoleID,
		RoleName:  "租户管理员",
		TenantID:  &tenantID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return tenant, adminUser, nil
}

// Update 更新租户信息
func (s *TenantService) Update(req models.UpdateTenantRequest) (*models.Tenant, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	querySQL := `UPDATE tenants SET name=$1, contact_name=$2, contact_phone=$3, contact_email=$4, description=$5, status=$6, updated_at=$7
		WHERE id=$8`

	_, err := db.Pool.Exec(ctx, querySQL,
		req.Name, req.ContactName, req.ContactPhone, req.ContactEmail, req.Description, req.Status, now, req.ID,
	)
	if err != nil {
		return nil, errors.New("更新租户失败")
	}

	return s.GetByID(req.ID)
}

// Delete 删除租户（级联删除租户下用户、角色、会话）
func (s *TenantService) Delete(id uint) error {
	if db.Pool == nil {
		return errors.New("数据库未连接")
	}

	ctx := context.Background()
	result, err := db.Pool.Exec(ctx, `DELETE FROM tenants WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("租户不存在")
	}
	return nil
}

// GetUserPermissions 获取用户的所有权限点编码
func (s *TenantService) GetUserPermissions(userID uint) []string {
	if db.Pool == nil {
		return []string{}
	}

	ctx := context.Background()
	querySQL := `
		SELECT DISTINCT p.code FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN users u ON u.role_id = rp.role_id
		WHERE u.id = $1
		ORDER BY p.code
	`

	rows, err := db.Pool.Query(ctx, querySQL, userID)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			continue
		}
		perms = append(perms, code)
	}
	if perms == nil {
		perms = []string{}
	}
	return perms
}

// GetUserTenant 获取用户所属租户ID
func (s *TenantService) GetUserTenant(userID uint) (*uint, string, error) {
	if db.Pool == nil {
		return nil, "", errors.New("数据库未连接")
	}

	ctx := context.Background()
	var tenantID *uint
	var tenantName string

	err := db.Pool.QueryRow(ctx,
		`SELECT u.tenant_id, COALESCE(t.name, '') FROM users u
		 LEFT JOIN tenants t ON u.tenant_id = t.id
		 WHERE u.id = $1`, userID,
	).Scan(&tenantID, &tenantName)
	if err != nil {
		return nil, "", err
	}

	return tenantID, tenantName, nil
}
