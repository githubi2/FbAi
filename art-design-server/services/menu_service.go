package services

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/githubi2/FbAi/art-design-server/db"
	"github.com/githubi2/FbAi/art-design-server/models"
)

// MenuService 菜单服务（PostgreSQL 实现）
type MenuService struct{}

var DefaultMenuService = &MenuService{}

// List 获取所有菜单平铺列表
func (s *MenuService) List() []models.Menu {
	if db.Pool == nil {
		return s.listFallback()
	}

	ctx := context.Background()
	querySQL := `SELECT id, parent_id, title, name, path, component, icon, sort_order, menu_type, hidden, status, created_at, updated_at
		FROM menus ORDER BY sort_order ASC, id ASC`

	rows, err := db.Pool.Query(ctx, querySQL)
	if err != nil {
		return s.listFallback()
	}
	defer rows.Close()

	var menus []models.Menu
	for rows.Next() {
		var m models.Menu
		if err := rows.Scan(&m.ID, &m.ParentID, &m.Title, &m.Name, &m.Path, &m.Component,
			&m.Icon, &m.SortOrder, &m.MenuType, &m.Hidden, &m.Status, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		menus = append(menus, m)
	}
	if menus == nil {
		menus = []models.Menu{}
	}
	return menus
}

// Tree 获取菜单树
func (s *MenuService) Tree() []models.MenuTree {
	all := s.List()
	return buildMenuTree(all, 0)
}

// TreeByIDs 按角色菜单ID列表获取菜单树
func (s *MenuService) TreeByIDs(menuIDs []int64) []models.MenuTree {
	all := s.List()
	idSet := make(map[uint]bool)
	for _, id := range menuIDs {
		idSet[uint(id)] = true
	}

	var filtered []models.Menu
	for _, m := range all {
		if idSet[m.ID] {
			filtered = append(filtered, m)
		}
	}

	return buildMenuTree(filtered, 0)
}

// GetByID 按ID获取菜单
func (s *MenuService) GetByID(id uint) (*models.Menu, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	querySQL := `SELECT id, parent_id, title, name, path, component, icon, sort_order, menu_type, hidden, status, created_at, updated_at
		FROM menus WHERE id = $1`

	var m models.Menu
	err := db.Pool.QueryRow(ctx, querySQL, id).Scan(
		&m.ID, &m.ParentID, &m.Title, &m.Name, &m.Path, &m.Component,
		&m.Icon, &m.SortOrder, &m.MenuType, &m.Hidden, &m.Status, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("菜单不存在")
	}
	return &m, nil
}

// Create 创建菜单
func (s *MenuService) Create(req models.CreateMenuRequest) (*models.Menu, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	hidden := false
	if req.Hidden != nil {
		hidden = *req.Hidden
	}
	menuType := req.MenuType
	if menuType == "" {
		menuType = "menu"
	}

	querySQL := `INSERT INTO menus (parent_id, title, name, path, component, icon, sort_order, menu_type, hidden, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id`

	var id uint
	err := db.Pool.QueryRow(ctx, querySQL,
		req.ParentID, req.Title, req.Name, req.Path, req.Component, req.Icon,
		req.SortOrder, menuType, hidden, req.Status, now, now,
	).Scan(&id)
	if err != nil {
		return nil, errors.New("创建菜单失败: " + err.Error())
	}

	return &models.Menu{
		ID:        id,
		ParentID:  req.ParentID,
		Title:     req.Title,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Icon:      req.Icon,
		SortOrder: req.SortOrder,
		MenuType:  menuType,
		Hidden:    hidden,
		Status:    req.Status,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// Update 更新菜单
func (s *MenuService) Update(req models.UpdateMenuRequest) (*models.Menu, error) {
	if db.Pool == nil {
		return nil, errors.New("数据库未连接")
	}

	ctx := context.Background()
	now := time.Now()

	// 检查菜单是否存在
	_, err := s.GetByID(req.ID)
	if err != nil {
		return nil, err
	}

	hidden := false
	if req.Hidden != nil {
		hidden = *req.Hidden
	}
	menuType := req.MenuType
	if menuType == "" {
		menuType = "menu"
	}

	querySQL := `UPDATE menus SET parent_id=$1, title=$2, name=$3, path=$4, component=$5, icon=$6, 
		sort_order=$7, menu_type=$8, hidden=$9, status=$10, updated_at=$11 WHERE id=$12`

	_, err = db.Pool.Exec(ctx, querySQL,
		req.ParentID, req.Title, req.Name, req.Path, req.Component, req.Icon,
		req.SortOrder, menuType, hidden, req.Status, now, req.ID,
	)
	if err != nil {
		return nil, errors.New("更新菜单失败")
	}

	return s.GetByID(req.ID)
}

// Delete 删除菜单
func (s *MenuService) Delete(id uint) error {
	if db.Pool == nil {
		return errors.New("数据库未连接")
	}

	ctx := context.Background()

	// 级联删除子菜单
	_, _ = db.Pool.Exec(ctx, `DELETE FROM menus WHERE parent_id = $1`, id)

	result, err := db.Pool.Exec(ctx, `DELETE FROM menus WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("菜单不存在")
	}
	return nil
}

func buildMenuTree(menus []models.Menu, parentID uint) []models.MenuTree {
	var result []models.MenuTree
	for _, m := range menus {
		if m.ParentID == parentID {
			node := models.MenuTree{
				ID:        m.ID,
				ParentID:  m.ParentID,
				Title:     m.Title,
				Name:      m.Name,
				Path:      m.Path,
				Component: m.Component,
				Icon:      m.Icon,
				SortOrder: m.SortOrder,
				MenuType:  m.MenuType,
				Hidden:    m.Hidden,
				Children:  buildMenuTree(menus, m.ID),
			}
			result = append(result, node)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].SortOrder < result[j].SortOrder
	})
	return result
}

// --- 内存 fallback（数据库不可用时使用）---

func (s *MenuService) listFallback() []models.Menu {
	now := time.Now()
	return []models.Menu{
		{ID: 1, ParentID: 0, Title: "Dashboard", Name: "Dashboard", Path: "/dashboard", Component: "LAYOUT", Icon: "dashboard", SortOrder: 1, MenuType: "directory", Status: 1, CreatedAt: now, UpdatedAt: now},
		{ID: 2, ParentID: 1, Title: "控制台", Name: "Console", Path: "console", Component: "/dashboard/console", SortOrder: 1, MenuType: "menu", Status: 1, CreatedAt: now, UpdatedAt: now},
		{ID: 3, ParentID: 0, Title: "系统管理", Name: "System", Path: "/system", Component: "LAYOUT", Icon: "system", SortOrder: 2, MenuType: "directory", Status: 1, CreatedAt: now, UpdatedAt: now},
		{ID: 4, ParentID: 3, Title: "用户管理", Name: "User", Path: "user", Component: "/system/user", SortOrder: 1, MenuType: "menu", Status: 1, CreatedAt: now, UpdatedAt: now},
		{ID: 5, ParentID: 3, Title: "角色管理", Name: "Role", Path: "role", Component: "/system/role", SortOrder: 2, MenuType: "menu", Status: 1, CreatedAt: now, UpdatedAt: now},
		{ID: 6, ParentID: 3, Title: "菜单管理", Name: "Menu", Path: "menu", Component: "/system/menu", SortOrder: 3, MenuType: "menu", Status: 1, CreatedAt: now, UpdatedAt: now},
	}
}
