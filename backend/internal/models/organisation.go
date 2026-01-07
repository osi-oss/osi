package models

type Organization struct {
	BaseModel

	Name         string `gorm:"not null"`
	LegalName    *string
	INN          *string
	OGRN         *string
	KPP          *string
	LegalAddress *string

	Status OrgStatus `gorm:"type:org_status;not null;default:'draft'"`

	Founders  []OrganizationFounder `gorm:"foreignKey:OrganizationID"`
	Members   []OrganizationMember  `gorm:"foreignKey:OrganizationID"`
	Locations []Location            `gorm:"foreignKey:OrganizationID"`
	Positions []Position            `gorm:"foreignKey:OrganizationID"`
}
