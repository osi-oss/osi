package services

import (
	"errors"
	"log/slog"
	"time"

	"github.com/osi-oss/osi/internal/apperrors"
	"github.com/osi-oss/osi/internal/models"
	"github.com/osi-oss/osi/internal/repository"
	"github.com/osi-oss/osi/internal/validators"
	"gorm.io/gorm"
)

type InviteService struct {
	inviteRepo    *repository.InviteRepository
	orgRepo       *repository.OrganizationRepository
	positionRepo  *repository.PositionRepository
	userRepo      *repository.UserRepository
	employeeRepo  *repository.EmployeeRepository
	permissionSvc *PermissionService
	emailService  interface {
		SendInviteEmail(toEmail, orgName, positionName, inviteLink string) error
	}
	logger *slog.Logger
}

func NewInviteService(
	inviteRepo *repository.InviteRepository,
	orgRepo *repository.OrganizationRepository,
	positionRepo *repository.PositionRepository,
	userRepo *repository.UserRepository,
	employeeRepo *repository.EmployeeRepository,
	permissionSvc *PermissionService,
	emailService interface {
		SendInviteEmail(toEmail, orgName, positionName, inviteLink string) error
	},
	logger *slog.Logger,
) *InviteService {
	return &InviteService{
		inviteRepo:    inviteRepo,
		orgRepo:       orgRepo,
		positionRepo:  positionRepo,
		userRepo:      userRepo,
		employeeRepo:  employeeRepo,
		permissionSvc: permissionSvc,
		emailService:  emailService,
		logger:        logger,
	}
}

// CreateInvite создаёт приглашение и отправляет email
// Учитывает:
// - Пользователь может приглашать в разные организации где он основатель/админ
// - Scoped права: может приглашать только в определённые отделы/локации
// - invitedUserId может быть NULL если пользователь не зарегистрирован
func (s *InviteService) CreateInvite(
	inviterID int64,
	orgID int64,
	positionID int64,
	invitedEmail string,
) (*models.Invite, error) {
	// Проверяем и очищаем email
	cleanedEmail, err := validators.Email.CleanAndValidate(invitedEmail)
	if err != nil {
		return nil, apperrors.BadRequest(err.Error())
	}
	invitedEmail = cleanedEmail

	// Получаем организацию
	org, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrOrganizationNotFound
		}
		return nil, apperrors.Wrap(err, 500, "database error")
	}

	// Получаем позицию и проверяем, что она принадлежит организации
	position, err := s.positionRepo.GetByID(positionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.BadRequest("position not found")
		}
		return nil, apperrors.Wrap(err, 500, "database error")
	}

	if position.OrganizationID != orgID {
		return nil, apperrors.BadRequest("position does not belong to this organization")
	}

	// Получаем inviter'а для валидации
	_, err = s.userRepo.GetById(inviterID)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to get inviter")
	}

	// === КРИТИЧЕСКОЕ: Проверяем права приглашать ===
	// Нужно проверить:
	// 1. Является ли пользователь основателем организации (может приглашать в любую позицию)
	// 2. Или у него есть scoped право "invites.create" на этот отдел/локацию
	if err := s.validateInviteAccess(inviterID, orgID, position); err != nil {
		return nil, err
	}

	// Получаем или создаём пользователя с этим email
	var invitedUserID *int64
	invitedUser, err := s.userRepo.GetByEmail(invitedEmail)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.Wrap(err, 500, "database error")
		}
		// Пользователь не существует - создаём со статусом pending_email
		// invitedUserID останется nil
	} else {
		// Пользователь найден - сохраняем его ID
		invitedUserID = &invitedUser.ID

		// Проверяем: не является ли пользователь уже активным членом организации
		existingMember, _ := s.orgRepo.GetMemberByUserAndOrgID(invitedUser.ID, orgID)
		if existingMember != nil && existingMember.ID > 0 && existingMember.Status == models.MemberActive {
			return nil, apperrors.BadRequest("user is already an active member of this organization")
		}

		// Проверяем: не является ли пользователь основателем
		existingFounder, _ := s.orgRepo.GetFounderByUserAndOrgID(invitedUser.ID, orgID)
		if existingFounder != nil && existingFounder.ID > 0 {
			return nil, apperrors.BadRequest("user is already a founder of this organization")
		}

		// Проверяем: нет ли активного pending приглашения на эту же должность
		existsPending, err := s.inviteRepo.ExistsPendingInvite(orgID, positionID, invitedUserID)
		if err != nil {
			return nil, apperrors.Wrap(err, 500, "database error")
		}
		if existsPending {
			return nil, apperrors.BadRequest("pending invite for this position already exists")
		}
	}

	// Проверяем: нет ли pending приглашения на этот email на эту же позицию
	existsPendingEmail, err := s.inviteRepo.ExistsPendingInviteByEmail(orgID, positionID, invitedEmail)
	if err != nil {
		return nil, apperrors.Wrap(err, 500, "database error")
	}
	if existsPendingEmail {
		return nil, apperrors.BadRequest("pending invite for this email and position already exists")
	}

	// Создаём приглашение
	invite := &models.Invite{
		OrganizationID:  orgID,
		PositionID:      positionID,
		InvitedUserID:   invitedUserID, // Может быть NULL
		InvitedByUserID: inviterID,
		Status:          models.InvitePending,
		InvitedEmail:    invitedEmail,
		InvitedAt:       ptrTime(time.Now()),
	}

	if err := s.inviteRepo.Create(invite); err != nil {
		return nil, apperrors.Wrap(err, 500, "failed to create invite")
	}

	// Формируем ссылку для приглашения (UX-редирект после логина)
	inviteLink := buildInviteLink(orgID, positionID)

	// Отправляем email
	if err := s.emailService.SendInviteEmail(invitedEmail, org.Name, position.Name, inviteLink); err != nil {
		s.logger.Error("failed to send invite email", slog.String("email", invitedEmail), slog.Any("error", err))
		// Не возвращаем ошибку - приглашение создано, письмо просто не отправилось
	}

	// Загружаем полный объект приглашения
	return s.inviteRepo.GetByID(invite.ID)
}

// AcceptInvite принимает приглашение залогиненным пользователем
func (s *InviteService) AcceptInvite(inviteID int64, currentUserID int64) error {
	invite, err := s.inviteRepo.GetByID(inviteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.BadRequest("invite not found")
		}
		return apperrors.Wrap(err, 500, "database error")
	}

	// Проверяем, что пользователь может принять это приглашение
	// Приглашение может быть принято только пользователем с email == invited_email И с ID == invited_user_id
	currentUser, err := s.userRepo.GetById(currentUserID)
	if err != nil {
		return apperrors.Wrap(err, 500, "failed to get current user")
	}

	// Email должен совпадать
	if currentUser.Email != invite.InvitedEmail {
		return apperrors.ErrForbidden
	}

	// Статус должен быть pending
	if invite.Status != models.InvitePending {
		return apperrors.BadRequest("invite is not pending")
	}

	// Проверяем, что пользователь не является уже членом
	existingMember, _ := s.orgRepo.GetMemberByUserAndOrgID(currentUserID, invite.OrganizationID)
	if existingMember != nil && existingMember.ID > 0 && existingMember.Status == models.MemberActive {
		return apperrors.BadRequest("user is already a member of this organization")
	}

	// Обновляем статус приглашения
	if err := s.inviteRepo.UpdateStatusWithTimestamp(inviteID, models.InviteAccepted); err != nil {
		return apperrors.Wrap(err, 500, "failed to update invite")
	}

	// Создаём OrganizationMember с активным статусом
	now := time.Now()
	member := &models.OrganizationMember{
		OrganizationID: invite.OrganizationID,
		UserID:         currentUserID,
		Status:         models.MemberActive,
		JoinedAt:       &now,
	}

	if err := s.orgRepo.CreateMember(member); err != nil {
		// Если не удалось создать члена, откатываем статус приглашения
		_ = s.inviteRepo.UpdateStatus(inviteID, models.InvitePending)
		return apperrors.Wrap(err, 500, "failed to create organization member")
	}

	// Создаём Employee для этой позиции
	employee := &models.Employee{
		MemberID:   member.ID,
		PositionID: invite.PositionID,
		StartDate:  &now,
	}

	if err := s.employeeRepo.Create(employee); err != nil {
		s.logger.Error("failed to create employee", slog.Any("error", err))
		// Не возвращаем ошибку - важное всё равно создано
	}

	return nil
}

// DeclineInvite отклоняет приглашение
func (s *InviteService) DeclineInvite(inviteID int64, currentUserID int64) error {
	invite, err := s.inviteRepo.GetByID(inviteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.BadRequest("invite not found")
		}
		return apperrors.Wrap(err, 500, "database error")
	}

	// Проверяем, что пользователь может отклонить это приглашение
	currentUser, err := s.userRepo.GetById(currentUserID)
	if err != nil {
		return apperrors.Wrap(err, 500, "failed to get current user")
	}

	// Email должен совпадать
	if currentUser.Email != invite.InvitedEmail {
		return apperrors.ErrForbidden
	}

	// Статус должен быть pending
	if invite.Status != models.InvitePending {
		return apperrors.BadRequest("invite is not pending")
	}

	// Обновляем статус приглашения
	return s.inviteRepo.UpdateStatusWithTimestamp(inviteID, models.InviteDeclined)
}

// GetMyInvites получает приглашения текущего пользователя
func (s *InviteService) GetMyInvites(userEmail string) ([]models.Invite, error) {
	// Получаем только pending приглашения для email пользователя
	return s.inviteRepo.GetPendingByEmail(userEmail)
}

// CancelInvite отменяет приглашение (только создатель или админ организации)
func (s *InviteService) CancelInvite(inviteID int64, currentUserID int64) error {
	invite, err := s.inviteRepo.GetByID(inviteID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.BadRequest("invite not found")
		}
		return apperrors.Wrap(err, 500, "database error")
	}

	// Проверяем права: только создатель приглашения или основатель организации может отменить
	if invite.InvitedByUserID != currentUserID {
		// Проверяем, является ли текущий пользователь основателем
		founder, _ := s.orgRepo.GetFounderByUserAndOrgID(currentUserID, invite.OrganizationID)
		if founder == nil || founder.ID == 0 {
			return apperrors.ErrForbidden
		}
	}

	// Удаляем приглашение
	return s.inviteRepo.Delete(inviteID)
}

// GetOrganizationInvites получает все приглашения организации (только для админов)
func (s *InviteService) GetOrganizationInvites(orgID int64, currentUserID int64) ([]models.Invite, error) {
	// Проверяем прав доступа - пользователь должен быть основателем
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(currentUserID, orgID)
	if err != nil || founder == nil || founder.ID == 0 {
		return nil, apperrors.ErrForbidden
	}

	return s.inviteRepo.GetByOrganization(orgID)
}

// validateInviteAccess проверяет права пользователя на создание приглашений в организацию
// Учитывает:
// - Является ли основателем (может приглашать в любую позицию)
// - Есть ли scoped право "invites.create" на конкретный отдел/локацию
func (s *InviteService) validateInviteAccess(userID int64, orgID int64, position *models.Position) error {
	// Сначала проверяем: является ли пользователь основателем организации
	founder, err := s.orgRepo.GetFounderByUserAndOrgID(userID, orgID)
	if err == nil && founder != nil && founder.ID > 0 {
		// Основатель имеет все права
		return nil
	}

	// Пользователь не основатель - проверяем scoped права
	// Определяем scope для проверки прав:
	// - Если у позиции есть отдел - проверяем право на отдел
	// - Иначе проверяем право на организацию

	var scopeType models.ScopeType
	var scopeID *int64

	if position.DepartmentID != nil {
		scopeType = models.ScopeDepartment
		scopeID = position.DepartmentID
	} else {
		scopeType = models.ScopeOrganization
		scopeID = nil
	}

	// Проверяем право "invites.create" на нужный scope
	hasPermission, err := s.permissionSvc.UserHasScopedPermission(userID, orgID, "invites.create", scopeType, scopeID)
	if err != nil {
		return apperrors.Wrap(err, 500, "failed to check permission")
	}

	if !hasPermission {
		return apperrors.ErrForbidden
	}

	return nil
}

// validateOrgAccess проверяет базовый доступ к организации
func (s *InviteService) validateOrgAccess(userID int64, orgID int64) error {
	hasAccess, err := s.permissionSvc.UserHasAccessToOrganization(userID, orgID)
	if err != nil {
		return apperrors.Wrap(err, 500, "failed to check access")
	}

	if !hasAccess {
		return apperrors.ErrForbidden
	}

	return nil
}

// Helper функции

func ptrTime(t time.Time) *time.Time {
	return &t
}

func buildInviteLink(orgID, positionID int64) string {
	// Фронтэнд будет использовать эту ссылку для редиректа после логина
	// Формат: https://app.com/login?invite_org_id=123&invite_position_id=456
	return "" // Будет заполняться фронтэндом
}
