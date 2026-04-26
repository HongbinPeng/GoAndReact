package handler

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"photo-album-server/internal/middleware"
	"photo-album-server/internal/model"
	"photo-album-server/internal/util"
)

func (h *Handler) GetProtectedFile(c *gin.Context) {
	currentUser, ok := middleware.GetCurrentUser(c)
	if !ok {
		util.Error(c, http.StatusUnauthorized, "未获取到当前登录用户信息")
		return
	}

	requestPath := "/uploads" + c.Param("filepath")
	if requestPath == "/uploads" || requestPath == "/uploads/" {
		util.Error(c, http.StatusBadRequest, "文件路径不能为空")
		return
	}

	switch {
	case strings.HasPrefix(requestPath, "/uploads/avatar/"):
		if err := h.checkAvatarAccess(requestPath); err != nil {
			h.handleFileAccessError(c, err)
			return
		}
	case strings.HasPrefix(requestPath, "/uploads/photos/"):
		if err := h.checkPhotoAccess(requestPath, currentUser.ID); err != nil {
			h.handleFileAccessError(c, err)
			return
		}
	default:
		util.Error(c, http.StatusForbidden, "无权访问该文件")
		return
	}

	localPath, err := h.resolveLocalUploadPath(requestPath)
	if err != nil {
		util.Error(c, http.StatusBadRequest, "文件路径不合法")
		return
	}

	if _, err := os.Stat(localPath); err != nil {
		if os.IsNotExist(err) {
			util.Error(c, http.StatusNotFound, "文件不存在")
			return
		}

		util.Error(c, http.StatusInternalServerError, "读取文件失败")
		return
	}

	c.File(localPath)
}

func (h *Handler) checkAvatarAccess(requestPath string) error {
	var user model.User
	if err := h.DB.Where("avatar_url = ?", requestPath).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errFileNotFound
		}

		return errQueryFileFailed
	}

	return nil
}

func (h *Handler) checkPhotoAccess(requestPath string, currentUserID uint) error {
	var photo model.Photo
	if err := h.DB.Where("file_path = ?", requestPath).First(&photo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errFileNotFound
		}

		return errQueryFileFailed
	}

	var album model.Album
	if err := h.DB.First(&album, photo.AlbumID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errFileNotFound
		}

		return errQueryFileFailed
	}

	if !album.IsPublic && album.UserID != currentUserID {
		return errFileForbidden
	}

	return nil
}

func (h *Handler) resolveLocalUploadPath(requestPath string) (string, error) {
	cleanedRequestPath := filepath.Clean(strings.TrimPrefix(requestPath, "/"))
	uploadRoot := filepath.Clean(h.Config.UploadDir)

	relativePath, err := filepath.Rel(uploadRoot, cleanedRequestPath)
	if err != nil || relativePath == "." || strings.HasPrefix(relativePath, "..") {
		return "", errInvalidFilePath
	}

	return cleanedRequestPath, nil
}

func (h *Handler) handleFileAccessError(c *gin.Context, err error) {
	switch err {
	case errFileNotFound:
		util.Error(c, http.StatusNotFound, "文件不存在")
	case errFileForbidden:
		util.Error(c, http.StatusForbidden, "无权访问该文件")
	default:
		util.Error(c, http.StatusInternalServerError, "校验文件访问权限失败")
	}
}
