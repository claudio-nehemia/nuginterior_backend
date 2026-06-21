package handler

import (
	"strconv"

	"github.com/claudio-nehemia/interior_backend/internal/helper"
	"github.com/claudio-nehemia/interior_backend/internal/service"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	svc service.NotificationService
}

func NewNotificationHandler(svc service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

func (h *NotificationHandler) Index(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		helper.Unauthorized(c, "User ID tidak ditemukan")
		return
	}
	userID := userIDVal.(uint)

	data, err := h.svc.GetNotificationsForUser(c.Request.Context(), userID)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Daftar notifikasi berhasil dimuat", data)
}

func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		helper.BadRequest(c, "ID notifikasi tidak valid", nil)
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		helper.Unauthorized(c, "User ID tidak ditemukan")
		return
	}
	userID := userIDVal.(uint)

	err = h.svc.MarkAsRead(c.Request.Context(), uint(id), userID)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Notifikasi telah ditandai dibaca", nil)
}

func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		helper.Unauthorized(c, "User ID tidak ditemukan")
		return
	}
	userID := userIDVal.(uint)

	err := h.svc.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Semua notifikasi telah ditandai dibaca", nil)
}

func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		helper.Unauthorized(c, "User ID tidak ditemukan")
		return
	}
	userID := userIDVal.(uint)

	count, err := h.svc.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		helper.InternalError(c, err.Error())
		return
	}

	helper.OK(c, "Jumlah notifikasi unread berhasil dimuat", gin.H{
		"unread_count": count,
	})
}
