package server

import (
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rs/zerolog/log"

	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/ports/http"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/utils"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/domain"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/queries"
)

type ODSGatewayServer struct {
	app app.ODSGatewayApp
}

func NewODSGateway(gwApp app.ODSGatewayApp) (*ODSGatewayServer, error) {
	return &ODSGatewayServer{gwApp}, nil
}

func (s *ODSGatewayServer) SearchOrganisations(ctx echo.Context, params http.SearchOrganisationsParams) error {
	result, err := s.app.Queries.SearchOrganisations.Handle(ctx.Request().Context(), queries.SearchOrganisationsQuery{
		Name:            params.Name,
		City:            params.City,
		Postcode:        params.Postcode,
		RoleCode:        params.RoleCode,
		Active:          params.Active,
		PrimaryRoleOnly: params.PrimaryRoleOnly,
		PageSize:        params.PageSize,
		Page:            params.Page,
	})
	if err != nil {
		log.Err(err).Msg("error searching organisations")
		return ctx.JSON(500, err.Error())
	}

	items := make([]http.Organisation, 0)
	for _, r := range result.Organisations {
		items = append(items, mapGetOrganisationResponse(r))
	}

	searchResult := http.OrganisationSearchResponse{
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    result.TotalCount,
		Items:    items,
	}

	return ctx.JSON(200, searchResult)
}

func (s *ODSGatewayServer) GetOrganisationByOdsCode(
	ctx echo.Context,
	odsCode string,
	_ http.GetOrganisationByOdsCodeParams,
) error {
	result, err := s.app.Queries.GetOrganisationByODSCode.Handle(
		ctx.Request().Context(),
		queries.GetOrganisationByODSCodeQuery{
			ODSCode: odsCode,
		},
	)
	if err != nil {
		log.Err(err).Msg("error getting organisation by ODS code")
		return ctx.JSON(500, err.Error())
	}

	return ctx.JSON(200, mapGetOrganisationResponse(result))
}

func mapGetOrganisationResponse(org domain.Organisation) http.Organisation {
	return http.Organisation{
		Id:       org.ID,
		IsActive: org.IsActive,
		Metadata: http.OrganisationMetadata{
			LastUpdated: org.Metadata.LastUpdated,
		},
		Name:              org.Name,
		OdsCode:           org.ODSCode,
		RecordClass:       org.RecordClass,
		OperationalPeriod: mapOperationalPeriod(org.OperationalPeriod),
		Address:           mapOrganisationAddress(org.Address),
		Roles:             mapOrganisationRoles(org.Roles),
	}
}

func mapOperationalPeriod(period *domain.OperationalPeriod) *http.OperationalPeriod {
	if period == nil {
		return nil
	}
	end := &openapi_types.Date{}
	if period.End != nil {
		end = &openapi_types.Date{
			Time: *period.End,
		}
	}
	return &http.OperationalPeriod{
		DateType: utils.Ref(period.DateType),
		End:      end,
		Start:    openapi_types.Date{Time: period.Start},
	}
}

func mapOrganisationRoles(roles []domain.OrganisationRole) *[]http.OrganisationRole {
	orgRoles := make([]http.OrganisationRole, 0)
	for _, role := range roles {
		orgRole := http.OrganisationRole{
			Code:    role.Code,
			Display: role.Display,
			Primary: role.Primary,
			Status:  http.OrganisationRoleStatus(role.Status),
		}
		if role.OperationalPeriod != nil {
			orgRole.OperationalPeriod = mapOperationalPeriod(role.OperationalPeriod)
		}
		orgRoles = append(orgRoles, orgRole)
	}
	return &orgRoles
}

func mapOrganisationAddress(address domain.Address) *http.Address {
	return &http.Address{
		City:       address.City,
		Country:    address.Country,
		Lines:      address.Lines,
		PostalCode: address.PostalCode,
	}
}
