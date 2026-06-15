package services

import (
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/githubi2/FbAi/art-design-server/models"
)

// MenuService 菜单服务
type MenuService struct {
	mu    sync.RWMutex
	menus map[uint]*models.Menu
	seq   uint
}

var DefaultMenuService = NewMenuService()

func NewMenuService() *MenuService {
	svc := &MenuService{
		menus: make(map[uint]*models.Menu),
		seq:   6,
	}
	// 初始化默认菜单（匹配前端路由结构）
	now := time.Now()
	menuList := []models.Menu{
		{ID: 1, ParentID: 0, Name: "Dashboard", Path: "/dashboard", Component: "LAYOUT", Icon: "dashboard", Sort: 1, Type: 1, Status: 1, Locale: "menus.dashboard", CreatedAt: now, UpdatedAt: now},
		{ID: 2, ParentID: 1, Name: "控制台", Path: "console", Component: "/dashboard/console", Icon: "", Sort: 1, Type: 2, Status: 1, Locale: "menus.dashboard.console", CreatedAt: now, UpdatedAt: now},
		{ID: 3, ParentID: 0, Name: "系统管理", Path: "/system", Component: "LAYOUT", Icon: "system", Sort: 2, Type: 1, Status: 1, Locale: "menus.system", CreatedAt: now, UpdatedAt: now},
		{ID: 4, ParentID: 3, Name: "用户管理", Path: "user", Component: "/system/user", Icon: "", Sort: 1, Type: 2, Status: 1, AuthMark: "system:user", Locale: "menus.system.user", CreatedAt: now, UpdatedAt: now},
		{ID: 5, ParentID: 3, Name: "角色管理", Path: "role", Component: "/system/role", Icon: "", Sort: 2, Type: 2, Status: 1, AuthMark: "system:role", Locale: "menus.system.role", CreatedAt: now, UpdatedAt: now},
		{ID: 6, ParentID: 3, Name: "菜单管理", Path: "menu", Component: "/system/menu", Icon: "", Sort: 3, Type: 2, Status: 1, AuthMark: "system:menu", Locale: "menus.system.menu", CreatedAt: now, UpdatedAt: now},
	}
	for i := range menuList {
		m := &menuList[i]
		svc.menus[m.ID] = m
	}
	return svc
}

// List 获取所有菜单平铺列表
func (s *MenuService) List() []models.Menu {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Menu, 0, len(s.menus))
	for _, m := range s.menus {
		result = append(result, *m)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Sort < result[j].Sort
	})
	return result
}

// Tree 获取菜单树
func (s *MenuService) Tree() []models.MenuTree {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 收集所有菜单
	all := make([]models.Menu, 0, len(s.menus))
	for _, m := range s.menus {
		all = append(all, *m)
	}

	// 构建树
	return buildTree(all, 0)
}

// TreeByIDs 按角色菜单ID列表获取菜单树
func (s *MenuService) TreeByIDs(menuIDs []uint) []models.MenuTree {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idSet := make(map[uint]bool)
	for _, id := range menuIDs {
		idSet[id] = true
	}

	var filtered []models.Menu
	for _, m := range s.menus {
		if idSet[m.ID] {
			filtered = append(filtered, *m)
		}
	}

	return buildTree(filtered, 0)
}

// GetByID 按ID获取菜单
func (s *MenuService) GetByID(id uint) (*models.Menu, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	m, ok := s.menus[id]
	if !ok {
		return nil, errors.New("菜单不存在")
	}
	return m, nil
}

// Create 创建菜单
func (s *MenuService) Create(req models.CreateMenuRequest) (*models.Menu, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.seq++
	now := time.Now()
	menu := &models.Menu{
		ID:        s.seq,
		ParentID:  req.ParentID,
		Name:      req.Name,
		Path:      req.Path,
		Component: req.Component,
		Icon:      req.Icon,
		Sort:      req.Sort,
		Type:      req.Type,
		Status:    req.Status,
		AuthMark:  req.AuthMark,
		Locale:    req.Locale,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.menus[s.seq] = menu
	return menu, nil
}

// Update 更新菜单
func (s *MenuService) Update(req models.UpdateMenuRequest) (*models.Menu, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	m, ok := s.menus[req.ID]
	if !ok {
		return nil, errors.New("菜单不存在")
	}

	m.ParentID = req.ParentID
	m.Name = req.Name
	m.Path = req.Path
	m.Component = req.Component
	m.Icon = req.Icon
	m.Sort = req.Sort
	m.Type = req.Type
	m.Status = req.Status
	m.AuthMark = req.AuthMark
	m.Locale = req.Locale
	m.UpdatedAt = time.Now()

	return m, nil
}

// Delete 删除菜单
func (s *MenuService) Delete(id uint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.menus[id]; !ok {
		return errors.New("菜单不存在")
	}

	// 级联删除子菜单
	for cid, cm := range s.menus {
		if cm.ParentID == id {
			delete(s.menus, cid)
		}
	}
	delete(s.menus, id)
	return nil
}

func buildTree(menus []models.Menu, parentID uint) []models.MenuTree {
	var result []models.MenuTree
	for _, m := range menus {
		if m.ParentID == parentID {
			node := models.MenuTree{
				ID:        m.ID,
				ParentID:  m.ParentID,
				Name:      m.Name,
				Path:      m.Path,
				Component: m.Component,
				Icon:      m.Icon,
				Sort:      m.Sort,
				Type:      m.Type,
				AuthMark:  m.AuthMark,
				Locale:    m.Locale,
				Children:  buildTree(menus, m.ID),
			}
			result = append(result, node)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Sort < result[j].Sort
	})
	return result
}
