package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/osi-oss/osi/internal/dto"
	"github.com/osi-oss/osi/internal/helpers"
	"github.com/osi-oss/osi/internal/services"
)

type InviteController struct {
	inviteService *services.InviteService
}

func NewInviteController(inviteService *services.InviteService) *InviteController {
	return &InviteController{
		inviteService: inviteService,
	}
}

// CreateInvite godoc
// @Summary      Создать приглашение в организацию
// @Description  Пригласить пользователя на должность в организацию по email.
// @Description  Требует право "invites.create" на уровне организации или отдела, где находится позиция.
// @Description  Если пользователь с таким email не существует, он будет автоматически создан со статусом pending_email.
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Param        request body dto.CreateInviteRequest true "Данные приглашения"
// @Success      201 {object} dto.InviteResponse "Приглашение успешно создано"
// @Failure      400 {object} dto.ErrorResponse "Ошибка валидации"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет прав на создание приглашения"
// @Failure      404 {object} dto.ErrorResponse "Организация или должность не найдена"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/invites [post]
func (ctrl *InviteController) CreateInvite(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	orgID, err := strconv.ParseInt(c.Param("orgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	var req dto.CreateInviteRequest
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	invite, err := ctrl.inviteService.CreateInvite(userID, orgID, req.PositionID, req.Email)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondCreated(c, dto.ToInviteResponse(invite))
}

// GetMyInvites godoc
// @Summary      Получить мои приглашения
// @Description  Возвращает все pending приглашения, отправленные на email текущего пользователя.
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200 {array} dto.InviteResponse "Список приглашений"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /invites/my [get]
func (ctrl *InviteController) GetMyInvites(c *gin.Context) {
	// Получаем email текущего пользователя из контекста (установлено auth middleware)
	userEmail, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "email not found in context"})
		return
	}

	invites, err := ctrl.inviteService.GetMyInvites(userEmail.(string))
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToInviteResponses(invites))
}

// AcceptInvite godoc
// @Summary      Принять приглашение
// @Description  Принять приглашение в организацию на должность.
// @Description  Автоматически добавляет пользователя в организацию и создаёт запись сотрудника.
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        inviteId path int true "ID приглашения" minimum(1) example(1)
// @Success      200 {object} dto.AcceptInviteResponse "Приглашение принято"
// @Failure      400 {object} dto.ErrorResponse "Приглашение не найдено или уже обработано"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет прав на принятие этого приглашения"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /invites/{inviteId}/accept [post]
func (ctrl *InviteController) AcceptInvite(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	inviteID, err := strconv.ParseInt(c.Param("inviteId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite id"})
		return
	}

	if err := ctrl.inviteService.AcceptInvite(inviteID, userID); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.AcceptInviteResponse{
		Message: "Invite accepted successfully",
	})
}

// DeclineInvite godoc
// @Summary      Отклонить приглашение
// @Description  Отклонить приглашение в организацию.
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        inviteId path int true "ID приглашения" minimum(1) example(1)
// @Success      200 {object} dto.DeclineInviteResponse "Приглашение отклонено"
// @Failure      400 {object} dto.ErrorResponse "Приглашение не найдено или уже обработано"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет прав на отклонение этого приглашения"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /invites/{inviteId}/decline [post]
func (ctrl *InviteController) DeclineInvite(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	inviteID, err := strconv.ParseInt(c.Param("inviteId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite id"})
		return
	}

	if err := ctrl.inviteService.DeclineInvite(inviteID, userID); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.DeclineInviteResponse{
		Message: "Invite declined successfully",
	})
}

// CancelInvite godoc
// @Summary      Отменить приглашение
// @Description  Отменить выданное приглашение.
// @Description  Может отменить только создатель приглашения или основатель организации.
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        inviteId path int true "ID приглашения" minimum(1) example(1)
// @Success      200 {object} dto.AcceptInviteResponse "Приглашение отменено"
// @Failure      400 {object} dto.ErrorResponse "Приглашение не найдено"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет прав на отмену этого приглашения"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /invites/{inviteId} [delete]
func (ctrl *InviteController) CancelInvite(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	inviteID, err := strconv.ParseInt(c.Param("inviteId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite id"})
		return
	}

	if err := ctrl.inviteService.CancelInvite(inviteID, userID); err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.AcceptInviteResponse{
		Message: "Invite cancelled successfully",
	})
}

// GetOrganizationInvites godoc
// @Summary      Получить приглашения организации
// @Description  Возвращает все приглашения организации.
// @Description  Доступно только основателям/администраторам организации.
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        orgId path int true "ID организации" minimum(1) example(1)
// @Success      200 {array} dto.InviteResponse "Список приглашений"
// @Failure      400 {object} dto.ErrorResponse "Неверный формат ID организации"
// @Failure      401 {object} dto.ErrorResponse "Отсутствует или невалидный токен авторизации"
// @Failure      403 {object} dto.ErrorResponse "Нет прав на просмотр приглашений"
// @Failure      500 {object} dto.ErrorResponse "Внутренняя ошибка сервера"
// @Router       /organizations/{orgId}/invites [get]
func (ctrl *InviteController) GetOrganizationInvites(c *gin.Context) {
	userID, err := helpers.GetUserID(c)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	orgID, err := strconv.ParseInt(c.Param("orgId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid organization id"})
		return
	}

	invites, err := ctrl.inviteService.GetOrganizationInvites(orgID, userID)
	if err != nil {
		helpers.RespondError(c, err)
		return
	}

	helpers.RespondOK(c, dto.ToInviteResponses(invites))
}
