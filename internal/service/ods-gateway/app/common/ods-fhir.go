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
	PageSize        *int
}

type OdsFHIRClient interface {
	SearchOrganisations(ctx context.Context, request SeachOrganisationsRequest) (*fhirHTTP.OrganizationBundle, error)
	GetOrganisationById(ctx context.Context, organisationId string) (*fhirHTTP.OrganizationResource, error)
}
