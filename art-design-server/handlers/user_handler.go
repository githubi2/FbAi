package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/githubi2/FbAi/art-design-server/middleware"
	"github.com/githubi2/FbAi/art-design-server/models"
	"github.com/githubi2/FbAi/art-design-server/services"
)

// UserHandler 用户处理器
type UserHandler struct{}

var DefaultUserHandler = &UserHandler{}

// List GET /api/v1/users
func (h *UserHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	list, total := services.DefaultUserService.List(page, size, keyword)
	c.JSON(http.StatusOK, models.PageSuccess(list, total, page, size))
}

// GetByID GET /api/v1/users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的ID"))
		return
	}

	user, err := services.DefaultUserService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.Success(user))
}

// Create POST /api/v1/users
func (h *UserHandler) Create(c *gin.Context) {
	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	user, err := services.DefaultUserService.Create(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, err.Error()))
		return
	}

	c.JSON(http.StatusCreated, models.Success(user))
}

// Update PUT /api/v1/users/:id
func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的ID"))
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}
	req.ID = uint(id)

	user, err := services.DefaultUserService.Update(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, err.Error()))
		return
	}

	// 如果管理员修改了密码，强制该用户退出登录
	if req.Password != "" {
		middleware.InvalidateUserSessions(uint(id))
	}

	c.JSON(http.StatusOK, models.Success(user))
}

// Delete DELETE /api/v1/users/:id
func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "无效的ID"))
		return
	}

	if err := services.DefaultUserService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, models.Error(models.CodeNotFound, err.Error()))
		return
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("删除成功", nil))
}

// BatchDelete POST /api/v1/users/batch-delete
func (h *UserHandler) BatchDelete(c *gin.Context) {
	var ids []uint
	if err := c.ShouldBindJSON(&ids); err != nil {
		c.JSON(http.StatusBadRequest, models.Error(models.CodeBadRequest, "参数错误: "+err.Error()))
		return
	}

	for _, id := range ids {
		services.DefaultUserService.Delete(id)
	}

	c.JSON(http.StatusOK, models.SuccessWithMsg("批量删除成功", nil))
}
