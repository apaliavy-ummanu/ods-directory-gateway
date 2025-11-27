package ods_fhir

import (
	"context"
	"fmt"

	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/utils"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/common"
	http "github.com/Cleo-Systems/ods-directory-gateway/pkg/ods-fhir-api/client"
	"github.com/pkg/errors"
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
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.Status())
	}

	return resp.ApplicationfhirJSON200, nil
}

func (c *Client) GetOrganisationById(ctx context.Context, organisationId string) (*http.OrganizationResource, error) {
	resp, err := c.apiClient.GetSingleOrganizationWithResponse(ctx, organisationId)
	if err != nil {
		return nil, errors.Wrap(err, "error getting organization by id")
	}

	if resp.StatusCode() != 200 {
		return nil, errors.New(resp.Status())
	}

	return resp.ApplicationfhirJSON200, nil
}
