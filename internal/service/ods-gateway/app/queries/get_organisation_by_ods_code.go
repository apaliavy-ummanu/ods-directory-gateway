package queries

import (
	"context"

	"github.com/pkg/errors"

	"github.com/Cleo-Systems/ods-fhir-gateway/internal/service/ods-gateway/app/common"
	"github.com/Cleo-Systems/ods-fhir-gateway/internal/service/ods-gateway/app/domain"
)

type GetOrganisationByODSCodeQuery struct {
	ODSCode string
}

type GetOrganisationByODSCodeQueryHandler interface {
	Handle(ctx context.Context, query GetOrganisationByODSCodeQuery) (domain.Organisation, error)
}

func NewGetOrganisationByODSCodeQueryHandler(fhirClient common.OdsFHIRClient) GetOrganisationByODSCodeQueryHandler {
	return &getOrganisationByODSCodeQueryHandlerImpl{
		fhirClient: fhirClient,
	}
}

type getOrganisationByODSCodeQueryHandlerImpl struct {
	fhirClient common.OdsFHIRClient
}

func (h *getOrganisationByODSCodeQueryHandlerImpl) Handle(
	ctx context.Context,
	query GetOrganisationByODSCodeQuery,
) (domain.Organisation, error) {
	if query.ODSCode == "" {
		return domain.Organisation{}, errors.New("ODS code is required")
	}

	organisation, err := h.fhirClient.GetOrganisationByID(ctx, query.ODSCode)
	if err != nil {
		return domain.Organisation{}, errors.Wrap(err, "error getting organisation from ODS API")
	}

	if organisation == nil {
		return domain.Organisation{}, errors.New("no data received from ODS API")
	}

	return mapOrganisationToDomain(*organisation), nil
}
