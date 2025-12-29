package common

import (
	"context"

	fhirHTTP "github.com/Cleo-Systems/ods-directory-gateway/pkg/ods-fhir-api/client"
)

type SeachOrganisationsRequest struct {
	Name            *string
	City            *string
	Postcode        *string
	RoleCode        *string
	Active          *bool
	PrimaryRoleOnly *bool
	PageSize        int
	Page            int
}

type OdsFHIRClient interface {
	SearchOrganisations(ctx context.Context, request SeachOrganisationsRequest) (*fhirHTTP.OrganizationBundle, error)
	GetOrganisationByID(ctx context.Context, organisationID string) (*fhirHTTP.OrganizationResource, error)
}
