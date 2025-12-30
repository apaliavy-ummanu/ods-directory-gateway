package common

import (
	"context"

	fhirHTTP "github.com/Cleo-Systems/ods-fhir-gateway/pkg/ods-fhir-api/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

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

//counterfeiter:generate -o ./mocks/fake_ods_fhir_client.gen.go . OdsFHIRClient
type OdsFHIRClient interface {
	SearchOrganisations(ctx context.Context, request SeachOrganisationsRequest) (*fhirHTTP.OrganizationBundle, error)
	GetOrganisationByID(ctx context.Context, organisationID string) (*fhirHTTP.OrganizationResource, error)
}
