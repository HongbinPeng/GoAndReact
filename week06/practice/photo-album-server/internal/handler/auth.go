package handler

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"photo-album-server/internal/model"
	"photo-album-server/internal/util"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Register(c *gin.Context) {
	username := strings.TrimSpace(c.PostForm("username"))
	password := c.PostForm("password")

	if username == "" || password == "" {
		util.Error(c, http.StatusBadRequest, "用户名和密码不能为空")
		return
	}

	avatarFile, err := c.FormFile("avatar")
	if err != nil {
		util.Error(c, http.StatusBadRequest, "头像文件不能为空")
		return
	}

	var existingUser model.User
	err = h.DB.Where("username = ?", username).First(&existingUser).Error
	if err == nil {
		util.Error(c, http.StatusConflict, "用户名已存在")
		return
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		util.Error(c, http.StatusInternalServerError, "检查用户名是否重复失败")
		return
	}

	avatarURL, _, err := util.SaveUploadedImage(avatarFile, h.Config.UploadDir, h.Config.MaxUploadSize, "avatar")
	if err != nil {
		util.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	hashedPassword, err := util.HashPassword(password)
	if err != nil {
		_ = os.Remove(strings.TrimPrefix(avatarURL, "/"))
		util.Error(c, http.StatusInternalServerError, "密码加密失败")
		return
	}

	user := model.User{
		Username:  username,
		Password:  hashedPassword,
		AvatarURL: avatarURL,
	}

	if err := h.DB.Create(&user).Error; err != nil {
		_ = os.Remove(strings.TrimPrefix(avatarURL, "/"))
		util.Error(c, http.StatusInternalServerError, "创建用户失败")
		return
	}

	util.Success(c, http.StatusCreated, "注册成功", nil)
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Error(c, http.StatusBadRequest, "请求体格式错误")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		util.Error(c, http.StatusBadRequest, "用户名和密码不能为空")
		return
	}

	var user model.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			util.Error(c, http.StatusUnauthorized, "用户名或密码错误")
			return
		}

		util.Error(c, http.StatusInternalServerError, "查询用户失败")
		return
	}

	if err := util.ComparePassword(user.Password, req.Password); err != nil {
		util.Error(c, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	token, err := util.GenerateToken(user.ID, user.Username, h.Config.JWTSecret)
	if err != nil {
		util.Error(c, http.StatusInternalServerError, "生成登录凭证失败")
		return
	}

	util.Success(c, http.StatusOK, "登录成功", gin.H{
		"token":      token,
		"token_type": "Bearer",
	})
}
