package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"photo-album-server/internal/middleware"
	"photo-album-server/internal/model"
	"photo-album-server/internal/util"
)

type createAlbumRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

type albumItemResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	CreatedAt   string `json:"created_at"`
}

type publicAlbumResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
	CreatedAt   string `json:"created_at"`
	Creator     gin.H  `json:"creator"`
}

func (h *Handler) CreateAlbum(c *gin.Context) {
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok {
		util.Error(c, http.StatusUnauthorized, "未获取到当前登录用户信息")
		return
	}

	var req createAlbumRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, http.StatusBadRequest, "请求体格式错误")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		util.Error(c, http.StatusBadRequest, "相册名称不能为空")
		return
	}

	album := model.Album{
		UserID:      currentUser.ID,
		Name:        req.Name,
		Description: strings.TrimSpace(req.Description),
		IsPublic:    req.IsPublic,
	}

	if err := h.DB.Create(&album).Error; err != nil {
		util.Error(c, http.StatusInternalServerError, "创建相册失败")
		return
	}

	util.Success(c, http.StatusCreated, "创建相册成功", nil)
}

func (h *Handler) GetMyAlbums(c *gin.Context) {
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok {
		util.Error(c, http.StatusUnauthorized, "未获取到当前登录用户信息")
		return
	}

	page, pageSize := util.ParsePagination(c)

	var total int64
	if err := h.DB.Model(&model.Album{}).Where("user_id = ?", currentUser.ID).Count(&total).Error; err != nil {
		util.Error(c, http.StatusInternalServerError, "统计我的相册数量失败")
		return
	}

	var albums []model.Album
	if err := h.DB.Where("user_id = ?", currentUser.ID).
		Order("created_at DESC").
		Offset(util.Offset(page, pageSize)).
		Limit(pageSize).
		Find(&albums).Error; err != nil {
		util.Error(c, http.StatusInternalServerError, "查询我的相册失败")
		return
	}

	items := make([]albumItemResponse, 0, len(albums))
	for _, album := range albums {
		items = append(items, buildAlbumItemResponse(album))
	}

	util.Success(c, http.StatusOK, "获取我的相册列表成功", util.BuildListData(items, page, pageSize, total))
}

func (h *Handler) GetPublicAlbums(c *gin.Context) {
	page, pageSize := util.ParsePagination(c)

	var total int64
	if err := h.DB.Model(&model.Album{}).Where("is_public = ?", true).Count(&total).Error; err != nil {
		util.Error(c, http.StatusInternalServerError, "统计公开相册数量失败")
		return
	}

	var albums []model.Album
	if err := h.DB.Preload("User").
		Where("is_public = ?", true).
		Order("created_at DESC").
		Offset(util.Offset(page, pageSize)).
		Limit(pageSize).
		Find(&albums).Error; err != nil {
		util.Error(c, http.StatusInternalServerError, "查询公开相册失败")
		return
	}

	items := make([]publicAlbumResponse, 0, len(albums))
	for _, album := range albums {
		items = append(items, publicAlbumResponse{
			ID:          album.ID,
			Name:        album.Name,
			Description: album.Description,
			IsPublic:    album.IsPublic,
			CreatedAt:   album.CreatedAt.Format("2006-01-02 15:04:05"),
			Creator: gin.H{
				"username":   album.User.Username,
				"avatar_url": album.User.AvatarURL,
			},
		})
	}

	util.Success(c, http.StatusOK, "获取公开相册列表成功", util.BuildListData(items, page, pageSize, total))
}

func buildAlbumItemResponse(album model.Album) albumItemResponse {
	return albumItemResponse{
		ID:          album.ID,
		Name:        album.Name,
		Description: album.Description,
		IsPublic:    album.IsPublic,
		CreatedAt:   album.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
