package queries_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/utils"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/common/mocks"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/queries"
	http "github.com/Cleo-Systems/ods-directory-gateway/pkg/ods-fhir-api/client"
)

// helper to create a handler with a mocked ODS client.
func newHandlerWithMock(t *testing.T) (queries.GetOrganisationByODSCodeQueryHandler, *mocks.FakeOdsFHIRClient) {
	t.Helper()

	mockODS := &mocks.FakeOdsFHIRClient{}

	h := queries.NewGetOrganisationByODSCodeQueryHandler(
		mockODS,
	)

	return h, mockODS
}

func TestGetOrganisationByODSCode_Success(t *testing.T) {
	t.Parallel()

	handler, mockODS := newHandlerWithMock(t)

	ctx := context.Background()
	odsCode := "ABC123"

	// --- helpers for FHIR date wrappers (adjust if your types differ) ---
	dt := func(t time.Time) *types.Date {
		return &types.Date{Time: t}
	}

	operStart := time.Date(2025, 1, 10, 9, 0, 0, 0, time.UTC)
	operEnd := time.Date(2025, 12, 31, 18, 0, 0, 0, time.UTC)

	roleStart := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	roleEnd := time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)

	lastUpdated := time.Date(2025, 2, 3, 10, 11, 12, 0, time.UTC)

	expectedOrg := &http.OrganizationResource{
		Active: utils.Ref(true),
		Address: &http.Address{
			City:       utils.Ref("Warsaw"),
			Country:    utils.Ref("Poland"),
			District:   utils.Ref("Mazowieckie"),
			Line:       utils.Ref([]string{"Line 1", "Line 2"}),
			PostalCode: utils.Ref("02-659"),
		},
		Id:   "ID",
		Name: "MyName",
		Identifier: &http.Identifier{
			System: utils.Ref("https://fhir.nhs.uk/Id/ods-organization-code"),
			Value:  utils.Ref(odsCode),
		},
		Meta: &http.Meta{
			LastUpdated: &lastUpdated,
		},

		// Extensions to cover:
		// - queries.ActivePeriodURL with Operational DateType (should be selected)
		// - queries.ActivePeriodURL with non-operational DateType (should be ignored)
		// - OrgRoleURL with one valid role (has Code) + one invalid role (missing Code -> filtered out)
		Extension: utils.Ref([]http.Extension{
			// Non-operational period (ignored)
			{
				Url: utils.Ref(queries.ActivePeriodURL),
				ValuePeriod: &http.Period{
					Start: dt(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)),
					End:   dt(time.Date(2020, 2, 1, 0, 0, 0, 0, time.UTC)),
					Extension: utils.Ref([]http.Extension{
						{
							Url:         utils.Ref(queries.DateTypeURL),
							ValueString: utils.Ref("Legal"),
						},
					}),
				},
			},
			// Operational period (selected)
			{
				Url: utils.Ref(queries.ActivePeriodURL),
				ValuePeriod: &http.Period{
					Start: dt(operStart),
					End:   dt(operEnd),
					Extension: utils.Ref([]http.Extension{
						{
							Url:         utils.Ref(queries.DateTypeURL),
							ValueString: utils.Ref("Operational"),
						},
					}),
				},
			},

			// Valid role
			{
				Url: utils.Ref(queries.OrgRoleURL),
				Extension: utils.Ref([]http.Extension{
					{
						Url: utils.Ref(queries.ExtensionRole),
						ValueCoding: &http.Coding{
							Code:    utils.Ref("ROLE-1"),
							Display: utils.Ref("Role One"),
						},
					},
					{
						Url:          utils.Ref(queries.ExtensionPrimaryRole),
						ValueBoolean: utils.Ref(true),
					},
					{
						Url:         utils.Ref(queries.ExtensionStatus),
						ValueString: utils.Ref("Active"),
					},
					{
						Url: utils.Ref(queries.ExtensionActivePeriod),
						ValuePeriod: &http.Period{
							Start: dt(roleStart),
							End:   dt(roleEnd),
							Extension: utils.Ref([]http.Extension{
								{
									Url:         utils.Ref(queries.DateTypeURL),
									ValueString: utils.Ref("Operational"),
								},
							}),
						},
					},
				}),
			},

			{
				Url: utils.Ref(queries.OrgRoleURL),
				Extension: utils.Ref([]http.Extension{
					{
						Url:         utils.Ref(queries.ExtensionStatus),
						ValueString: utils.Ref("Inactive"),
					},
				}),
			},
		}),
	}

	mockODS.GetOrganisationByIDReturns(expectedOrg, nil)

	query := queries.GetOrganisationByODSCodeQuery{
		ODSCode: odsCode,
	}

	result, err := handler.Handle(ctx, query)
	require.NoError(t, err)

	require.NotNil(t, result)
	assert.Equal(t, expectedOrg.Id, result.ID)
	assert.Equal(t, expectedOrg.Name, result.Name)
	assert.Equal(t, *expectedOrg.Active, result.IsActive)

	require.NotNil(t, result.Address)
	assert.Equal(t, *expectedOrg.Address.City, *result.Address.City)
	assert.Equal(t, *expectedOrg.Address.Country, *result.Address.Country)
	assert.Equal(t, *expectedOrg.Address.PostalCode, *result.Address.PostalCode)
	assert.Equal(t, expectedOrg.Address.Line, result.Address.Lines)

	// ODSCode mapping: Identifier.System matches ODSCodeURL => use Identifier.Value (not org.Id)
	assert.Equal(t, *expectedOrg.Identifier.Value, result.ODSCode)

	// Metadata mapping
	assert.Equal(t, *expectedOrg.Meta.LastUpdated, result.Metadata.LastUpdated)

	// OperationalPeriod mapping: should pick only DateType == "Operational"
	require.NotNil(t, result.OperationalPeriod)
	assert.Equal(t, operStart, result.OperationalPeriod.Start)
	require.NotNil(t, result.OperationalPeriod.End)
	assert.Equal(t, operEnd, *result.OperationalPeriod.End)
	assert.Equal(t, "Operational", result.OperationalPeriod.DateType)

	// Roles mapping: one valid role only (second should be filtered out because Code == "")
	require.Len(t, result.Roles, 1)
	r := result.Roles[0]
	assert.Equal(t, "ROLE-1", r.Code)
	assert.Equal(t, "Role One", r.Display)
	assert.True(t, r.Primary)
	assert.Equal(t, "Active", r.Status)

	require.NotNil(t, r.OperationalPeriod)
	assert.Equal(t, roleStart, r.OperationalPeriod.Start)
	require.NotNil(t, r.OperationalPeriod.End)
	assert.Equal(t, roleEnd, *r.OperationalPeriod.End)
	assert.Equal(t, "Operational", r.OperationalPeriod.DateType)
}

func TestGetOrganisationByODSCode_NotFound(t *testing.T) {
	t.Parallel()

	handler, mockODS := newHandlerWithMock(t)

	ctx := context.Background()
	odsCode := "UNKNOWN"

	var notFoundErr = fmt.Errorf("not found")

	mockODS.GetOrganisationByIDReturns(nil, notFoundErr)

	query := queries.GetOrganisationByODSCodeQuery{
		ODSCode: odsCode,
	}

	_, err := handler.Handle(ctx, query)

	require.Error(t, err)
	assert.ErrorIs(t, err, notFoundErr)
}

func TestGetOrganisationByODSCode_InvalidInput(t *testing.T) {
	t.Parallel()

	handler, _ := newHandlerWithMock(t)

	ctx := context.Background()

	// e.g. empty ODS code should be rejected by the handler
	query := queries.GetOrganisationByODSCodeQuery{
		ODSCode: "",
	}

	_, err := handler.Handle(ctx, query)

	require.Error(t, err)
	require.ErrorContains(t, err, "ODS code is required")
}
