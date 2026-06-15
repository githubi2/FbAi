package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/models"
	"github.com/githubi2/FbAi/art-design-server/services"
)

// MenuHandler 菜单处理器
type MenuHandler struct{}

var DefaultMenuHandler = &MenuHandler{}

// List GET /api/v1/menus — 平铺列表
func (h *MenuHandler) List(c *gin.Context) {
	list := services.DefaultMenuService.List()
	c.JSON(http.StatusOK, models.Success(list))
}

// Tree GET /api/v1/menus/tree — 树形结构
func (h *MenuHandler) Tree(c *gin.Context) {
	tree := services.DefaultMenuService.Tree()
	c.JSON(http.StatusOK, models.Success(tree))
}

// GetByID GET /api/v1/menus/:id
func (h *MenuHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的ID"))
		return
	}

	menu, err := services.DefaultMenuService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(menu))
}

// Create POST /api/v1/menus
func (h *MenuHandler) Create(c *gin.Context) {
	var req models.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	menu, err := services.DefaultMenuService.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.Success(menu))
}

// Update PUT /api/v1/menus/:id
func (h *MenuHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的ID"))
		return
	}

	var req models.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}
	req.ID = uint(id)

	menu, err := services.DefaultMenuService.Update(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(menu))
}

// Delete DELETE /api/v1/menus/:id
func (h *MenuHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的ID"))
		return
	}

	if err := services.DefaultMenuService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("删除成功", nil))
}
