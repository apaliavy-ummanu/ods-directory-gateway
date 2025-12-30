package queries

import (
	"context"
	"strconv"

	"github.com/pkg/errors"

	"github.com/Cleo-Systems/ods-fhir-gateway/internal/service/common/utils"
	"github.com/Cleo-Systems/ods-fhir-gateway/internal/service/ods-gateway/app/common"
	"github.com/Cleo-Systems/ods-fhir-gateway/internal/service/ods-gateway/app/domain"
)

type SearchOrganisationsQuery struct {
	Name            *string
	City            *string
	Postcode        *string
	RoleCode        *string
	Active          *bool
	PrimaryRoleOnly *bool
	PageSize        int
	Page            int
}

type SearchOrganisationsResponse struct {
	Organisations []domain.Organisation
	TotalCount    int
}

type SearchOrganisationsQueryHandler interface {
	Handle(ctx context.Context, query SearchOrganisationsQuery) (SearchOrganisationsResponse, error)
}

func NewSearchOrganisationsQueryHandler(fhirClient common.OdsFHIRClient) SearchOrganisationsQueryHandler {
	return &searchOrganisationsQueryHandlerImpl{
		fhirClient: fhirClient,
	}
}

type searchOrganisationsQueryHandlerImpl struct {
	fhirClient common.OdsFHIRClient
}

func (h *searchOrganisationsQueryHandlerImpl) Handle(
	ctx context.Context,
	query SearchOrganisationsQuery,
) (SearchOrganisationsResponse, error) {
	organisationBundle, err := h.fhirClient.SearchOrganisations(ctx, common.SeachOrganisationsRequest{
		Name:            query.Name,
		City:            query.City,
		Postcode:        query.Postcode,
		RoleCode:        query.RoleCode,
		Active:          query.Active,
		PrimaryRoleOnly: query.PrimaryRoleOnly,
		PageSize:        query.PageSize,
		Page:            query.Page,
	})
	if err != nil {
		return SearchOrganisationsResponse{}, errors.Wrap(err, "error getting organisation from ODS API")
	}

	orgs := make([]domain.Organisation, 0)
	for _, entry := range utils.Deref(organisationBundle.Entry) {
		if entry.Resource != nil {
			orgs = append(orgs, mapOrganisationToDomain(*entry.Resource))
		}
	}

	total, err := strconv.Atoi(utils.Deref(organisationBundle.Total))
	if err != nil {
		return SearchOrganisationsResponse{}, err
	}

	return SearchOrganisationsResponse{
		Organisations: orgs,
		TotalCount:    total,
	}, nil
}
