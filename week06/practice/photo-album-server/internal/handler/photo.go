package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"photo-album-server/internal/middleware"
	"photo-album-server/internal/model"
	"photo-album-server/internal/util"
)

type photoItemResponse struct {
	ID        uint   `json:"id"`
	FilePath  string `json:"file_path"`
	FileSize  int64  `json:"file_size"`
	CreatedAt string `json:"created_at"`
}

func (h *Handler) UploadPhoto(c *gin.Context) {
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok {
		util.Error(c, http.StatusUnauthorized, "未获取到当前登录用户信息")
		return
	}

	album, err := h.findAlbumByParamID(c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.Error(c, http.StatusNotFound, "相册不存在")
			return
		}

		util.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if album.UserID != currentUser.ID {
		util.Error(c, http.StatusForbidden, "无权向该相册上传照片")
		return
	}

	photoFile, err := c.FormFile("photo")
	if err != nil {
		util.Error(c, http.StatusBadRequest, "照片文件不能为空")
		return
	}

	filePath, fileSize, err := util.SaveUploadedImage(photoFile, h.Config.UploadDir, h.Config.MaxUploadSize, "photos", fmt.Sprintf("album_%d", album.ID))
	if err != nil {
		util.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	photo := model.Photo{
		AlbumID:  album.ID,
		FilePath: filePath,
		FileSize: fileSize,
	}

	if err := h.DB.Create(&photo).Error; err != nil {
		_ = os.Remove(filePath[1:])
		util.Error(c, http.StatusInternalServerError, "保存照片记录失败")
		return
	}

	util.Success(c, http.StatusCreated, "上传照片成功", nil)
}

func (h *Handler) GetAlbumPhotos(c *gin.Context) {
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok {
		util.Error(c, http.StatusUnauthorized, "未获取到当前登录用户信息")
		return
	}

	album, err := h.findAlbumByParamID(c.Param("id"))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.Error(c, http.StatusNotFound, "相册不存在")
			return
		}

		util.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if !album.IsPublic && album.UserID != currentUser.ID {
		util.Error(c, http.StatusForbidden, "无权查看该私有相册")
		return
	}

	page, pageSize := util.ParsePagination(c)

	var total int64
	if err := h.DB.Model(&model.Photo{}).Where("album_id = ?", album.ID).Count(&total).Error; err != nil {
		util.Error(c, http.StatusInternalServerError, "统计照片数量失败")
		return
	}

	var photos []model.Photo
	if err := h.DB.Where("album_id = ?", album.ID).
		Order("created_at DESC").
		Offset(util.Offset(page, pageSize)).
		Limit(pageSize).
		Find(&photos).Error; err != nil {
		util.Error(c, http.StatusInternalServerError, "查询照片列表失败")
		return
	}

	items := make([]photoItemResponse, 0, len(photos))
	for _, photo := range photos {
		items = append(items, photoItemResponse{
			ID:        photo.ID,
			FilePath:  photo.FilePath,
			FileSize:  photo.FileSize,
			CreatedAt: photo.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	util.Success(c, http.StatusOK, "获取相册照片列表成功", util.BuildListData(items, page, pageSize, total))
}

func (h *Handler) findAlbumByParamID(idParam string) (*model.Album, error) {
	albumID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || albumID == 0 {
		return nil, fmt.Errorf("相册ID不合法")
	}

	var album model.Album
	if err := h.DB.First(&album, albumID).Error; err != nil {
		return nil, err
	}

	return &album, nil
}
