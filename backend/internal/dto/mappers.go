package dto

import "github.com/osi-oss/osi/internal/models"

// ToUserResponse преобразует User в UserResponse
func ToUserResponse(u *models.User) UserResponse {
	return UserResponse{
		ID:            u.ID,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		MiddleName:    u.MiddleName,
		Status:        string(u.Status),
		EmailVerified: u.IsEmailVerified,
		HasPassword:   u.HasPassword(),
		CreatedAt:     u.CreatedAt,
	}
}

// ToOrganizationResponse преобразует Organization в OrganizationResponse
func ToOrganizationResponse(o *models.Organization) OrganizationResponse {
	return OrganizationResponse{
		ID:           o.ID,
		Name:         o.Name,
		LegalName:    o.LegalName,
		INN:          o.INN,
		OGRN:         o.OGRN,
		KPP:          o.KPP,
		LegalAddress: o.LegalAddress,
		Status:       string(o.Status),
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}
}

// ToOrganizationResponses преобразует slice Organization в slice OrganizationResponse
func ToOrganizationResponses(orgs []models.Organization) []OrganizationResponse {
	result := make([]OrganizationResponse, len(orgs))
	for i, o := range orgs {
		result[i] = ToOrganizationResponse(&o)
	}
	return result
}

// ToLocationResponse преобразует Location в LocationResponse
func ToLocationResponse(l *models.Location) LocationResponse {
	return LocationResponse{
		ID:             l.ID,
		OrganizationID: l.OrganizationID,
		Name:           l.Name,
		Address:        l.Address,
		Source:         l.Source,
		IsVerified:     l.IsVerified,
		IsActive:       l.IsActive,
		CreatedAt:      l.CreatedAt,
		UpdatedAt:      l.UpdatedAt,
	}
}

// ToLocationResponses преобразует slice Location в slice LocationResponse
func ToLocationResponses(locs []models.Location) []LocationResponse {
	result := make([]LocationResponse, len(locs))
	for i, l := range locs {
		result[i] = ToLocationResponse(&l)
	}
	return result
}

// ToDepartmentResponse преобразует Department в DepartmentResponse
func ToDepartmentResponse(d *models.Department) DepartmentResponse {
	return DepartmentResponse{
		ID:          d.ID,
		LocationID:  d.LocationID,
		ParentID:    d.ParentID,
		Name:        d.Name,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// ToDepartmentResponses преобразует slice Department в slice DepartmentResponse
func ToDepartmentResponses(depts []models.Department) []DepartmentResponse {
	result := make([]DepartmentResponse, len(depts))
	for i, d := range depts {
		result[i] = ToDepartmentResponse(&d)
	}
	return result
}

// ToPositionResponse преобразует Position в PositionResponse
func ToPositionResponse(p *models.Position) PositionResponse {
	return PositionResponse{
		ID:             p.ID,
		OrganizationID: p.OrganizationID,
		DepartmentID:   p.DepartmentID,
		Name:           p.Name,
		IsAdmin:        p.IsAdmin,
		Description:    p.Description,
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}
}

// ToPositionResponses преобразует slice Position в slice PositionResponse
func ToPositionResponses(positions []models.Position) []PositionResponse {
	result := make([]PositionResponse, len(positions))
	for i, p := range positions {
		result[i] = ToPositionResponse(&p)
	}
	return result
}

// ToMemberResponse преобразует OrganizationMember в MemberResponse
func ToMemberResponse(m *models.OrganizationMember) MemberResponse {
	resp := MemberResponse{
		ID:             m.ID,
		OrganizationID: m.OrganizationID,
		UserID:         m.UserID,
		Status:         string(m.Status),
		JoinedAt:       m.JoinedAt,
	}
	if m.User.ID != 0 {
		resp.User = ToUserResponse(&m.User)
	}
	return resp
}

// ToEmployeeResponse преобразует Employee в EmployeeResponse
func ToEmployeeResponse(e *models.Employee) EmployeeResponse {
	resp := EmployeeResponse{
		ID:         e.ID,
		MemberID:   e.MemberID,
		PositionID: e.PositionID,
		IsIntern:   e.IsIntern,
		StartDate:  e.StartDate,
		EndDate:    e.EndDate,
	}
	if e.Position.ID != 0 {
		resp.Position = ToPositionResponse(&e.Position)
	}
	return resp
}

// ToInviteResponse преобразует Invite в InviteResponse
func ToInviteResponse(i *models.Invite) InviteResponse {
	resp := InviteResponse{
		ID:              i.ID,
		OrganizationID:  i.OrganizationID,
		PositionID:      i.PositionID,
		InvitedEmail:    i.InvitedEmail,
		Status:          string(i.Status),
		InvitedUserID:   i.InvitedUserID,
		InvitedByUserID: i.InvitedByUserID,
		CreatedAt:       i.CreatedAt,
		AcceptedAt:      i.AcceptedAt,
		DeclinedAt:      i.DeclinedAt,
	}

	if i.Organization.ID != 0 {
		orgResp := ToOrganizationResponse(&i.Organization)
		resp.Organization = &orgResp
	}

	if i.Position.ID != 0 {
		posResp := ToPositionResponse(&i.Position)
		resp.Position = &posResp
	}

	if i.InvitedByUser.ID != 0 {
		userResp := ToUserResponse(&i.InvitedByUser)
		resp.InvitedByUser = &userResp
	}

	return resp
}

// ToInviteResponses преобразует slice Invite в slice InviteResponse
func ToInviteResponses(invites []models.Invite) []InviteResponse {
	result := make([]InviteResponse, len(invites))
	for i, inv := range invites {
		result[i] = ToInviteResponse(&inv)
	}
	return result
}
