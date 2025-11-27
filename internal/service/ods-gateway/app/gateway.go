package app

import (
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/queries"
)

type Queries struct {
	GetOrganisationByODSCode queries.GetOrganisationByODSCodeQueryHandler
	SearchOrganisations      queries.SearchOrganisationsQueryHandler
}

type ODSGatewayApp struct {
	Queries Queries
}
