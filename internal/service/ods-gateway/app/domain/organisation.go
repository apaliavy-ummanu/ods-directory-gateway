package domain

import (
	"time"
)

type Organisation struct {
	ID                string
	OdsCode           string
	Name              string
	IsActive          bool
	Metadata          OrganisationMetadata
	Address           Address
	OperationalPeriod *OperationalPeriod
	RecordClass       string
	Roles             []OrganisationRole
}

type OrganisationMetadata struct {
	LastUpdated time.Time
}

type Address struct {
	City       *string
	Country    *string
	Lines      *[]string
	PostalCode *string
}

type OperationalPeriod struct {
	DateType string
	End      *time.Time
	Start    time.Time
}

type OrganisationRole struct {
	Code              string `json:"code"`
	Display           string
	OperationalPeriod *OperationalPeriod
	Primary           bool `json:"primary"`
	Status            string
}
