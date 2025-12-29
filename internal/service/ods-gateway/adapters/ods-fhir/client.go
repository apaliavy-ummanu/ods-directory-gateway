package odsfhir

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/utils"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/common"
	http "github.com/Cleo-Systems/ods-directory-gateway/pkg/ods-fhir-api/client"
)

type Client struct {
	apiClient http.ClientWithResponsesInterface
}

func NewClient(apiClient http.ClientWithResponsesInterface) *Client {
	return &Client{
		apiClient: apiClient,
	}
}

func (c *Client) SearchOrganisations(
	ctx context.Context,
	req common.SeachOrganisationsRequest,
) (*http.OrganizationBundle, error) {
	params := http.GetOrganizationResourcesParams{
		NameContains:              req.Name,
		Active:                    req.Active,
		AddressPostalcodeContains: req.Postcode,
		AddressCityContains:       req.City,
		OdsOrgRole:                req.RoleCode,
		OdsOrgPrimaryRole:         req.PrimaryRoleOnly,
		UnderscoreCount:           utils.Ref(fmt.Sprintf("%d", utils.Deref(req.PageSize))),
	}
	resp, err := c.apiClient.GetOrganizationResourcesWithResponse(ctx, &params)
	if err != nil {
		log.Err(err).Msg("error getting organisations from ODS API")
		return nil, err
	}

	if resp.StatusCode() != 200 {
		log.Err(errors.New(resp.Status())).Msg("error getting organisations from ODS API")
		return nil, errors.New(resp.Status())
	}

	return resp.ApplicationfhirJSON200, nil
}

func (c *Client) GetOrganisationByID(ctx context.Context, organisationID string) (*http.OrganizationResource, error) {
	resp, err := c.apiClient.GetSingleOrganizationWithResponse(ctx, organisationID)
	if err != nil {
		log.Err(err).Msg("error getting organisation from ODS API")
		return nil, errors.Wrap(err, "error getting organisation by id")
	}

	if resp.StatusCode() != 200 {
		log.Err(errors.New(resp.Status())).Msg("error getting organisation from ODS API")
		return nil, errors.New(resp.Status())
	}

	return resp.ApplicationfhirJSON200, nil
}
