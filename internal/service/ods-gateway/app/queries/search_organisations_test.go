package queries_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/common/mocks"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/utils"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/queries"
	http "github.com/Cleo-Systems/ods-directory-gateway/pkg/ods-fhir-api/client"
)

func newSearchHandlerWithMock(t *testing.T) (queries.SearchOrganisationsQueryHandler, *mocks.FakeOdsFHIRClient) {
	t.Helper()

	mockODS := &mocks.FakeOdsFHIRClient{}

	h := queries.NewSearchOrganisationsQueryHandler(
		mockODS,
	)

	return h, mockODS
}

func TestSearchOrganisations_Success(t *testing.T) {
	t.Parallel()

	handler, mockODS := newSearchHandlerWithMock(t)

	ctx := context.Background()

	// input query
	q := queries.SearchOrganisationsQuery{
		Name:            utils.Ref("Acme"),
		City:            utils.Ref("Warsaw"),
		Postcode:        utils.Ref("02-659"),
		RoleCode:        utils.Ref("ROLE-1"),
		Active:          utils.Ref(true),
		PrimaryRoleOnly: utils.Ref(true),
		PageSize:        25,
		Page:            2,
	}

	// helper for FHIR date wrapper (matches your previous test)
	dt := func(t time.Time) *types.Date { return &types.Date{Time: t} }

	operStart := time.Date(2025, 1, 10, 9, 0, 0, 0, time.UTC)
	operEnd := time.Date(2025, 12, 31, 18, 0, 0, 0, time.UTC)

	// org #1 (identifier matches -> ODSCode taken from Identifier.Value)
	org1ODS := "ABC123"
	org1 := http.OrganizationResource{
		Id:     "ID-1",
		Name:   "Org One",
		Active: utils.Ref(true),
		Identifier: &http.Identifier{
			System: utils.Ref("https://fhir.nhs.uk/Id/ods-organization-code"),
			Value:  utils.Ref(org1ODS),
		},
		Address: &http.Address{
			City:       utils.Ref("Warsaw"),
			Country:    utils.Ref("Poland"),
			Line:       utils.Ref([]string{"Line 1"}),
			PostalCode: utils.Ref("02-659"),
		},
		Extension: utils.Ref([]http.Extension{
			{
				Url: utils.Ref(queries.ActivePeriodURL),
				ValuePeriod: &http.Period{
					Start: dt(operStart),
					End:   dt(operEnd),
					Extension: utils.Ref([]http.Extension{
						{Url: utils.Ref(queries.DateTypeURL), ValueString: utils.Ref("Operational")},
					}),
				},
			},
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
				}),
			},
		}),
	}

	// org #2 (identifier does NOT match -> ODSCode should default to org.Id)
	org2 := http.OrganizationResource{
		Id:     "ID-2",
		Name:   "Org Two",
		Active: utils.Ref(false),
		Identifier: &http.Identifier{
			System: utils.Ref("some-other-system"),
			Value:  utils.Ref("SHOULD_NOT_BE_USED"),
		},
	}

	// bundle with 3 entries: org1, nil-resource (skipped), org2
	bundle := &http.OrganizationBundle{
		Total: utils.Ref("3"),
		Entry: utils.Ref([]http.OrganizationEntry{
			{Resource: &org1},
			{Resource: nil}, // should be ignored
			{Resource: &org2},
		}),
	}

	mockODS.SearchOrganisationsReturns(bundle, nil)

	// act
	resp, err := handler.Handle(ctx, q)
	require.NoError(t, err)

	// assert client called with expected request payload
	require.Equal(t, 1, mockODS.SearchOrganisationsCallCount())
	_, req := mockODS.SearchOrganisationsArgsForCall(0)

	assert.Equal(t, q.Name, req.Name)
	assert.Equal(t, q.City, req.City)
	assert.Equal(t, q.Postcode, req.Postcode)
	assert.Equal(t, q.RoleCode, req.RoleCode)
	assert.Equal(t, q.Active, req.Active)
	assert.Equal(t, q.PrimaryRoleOnly, req.PrimaryRoleOnly)
	assert.Equal(t, q.PageSize, req.PageSize)
	assert.Equal(t, q.Page, req.Page)

	// assert response
	assert.Equal(t, 3, resp.TotalCount)

	require.Len(t, resp.Organisations, 2) // nil entry must be skipped

	// org1 mapped
	got1 := resp.Organisations[0]
	assert.Equal(t, "ID-1", got1.ID)
	assert.Equal(t, "Org One", got1.Name)
	assert.Equal(t, true, got1.IsActive)
	assert.Equal(t, org1ODS, got1.ODSCode)
	require.NotNil(t, got1.OperationalPeriod)
	assert.Equal(t, operStart, got1.OperationalPeriod.Start)
	require.NotNil(t, got1.OperationalPeriod.End)
	assert.Equal(t, operEnd, *got1.OperationalPeriod.End)

	// org2 mapped (ODSCode falls back to org.Id)
	got2 := resp.Organisations[1]
	assert.Equal(t, "ID-2", got2.ID)
	assert.Equal(t, "Org Two", got2.Name)
	assert.Equal(t, false, got2.IsActive)
	assert.Equal(t, "ID-2", got2.ODSCode)
}

func TestSearchOrganisations_FHIRClientError_Wrapped(t *testing.T) {
	t.Parallel()

	handler, mockODS := newSearchHandlerWithMock(t)

	ctx := context.Background()
	q := queries.SearchOrganisationsQuery{Name: utils.Ref("Acme")}

	mockODS.SearchOrganisationsReturns(nil, errors.New("boom"))

	_, err := handler.Handle(ctx, q)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error getting organisation from ODS API")
}

func TestSearchOrganisations_InvalidTotal_ReturnsError(t *testing.T) {
	t.Parallel()

	handler, mockODS := newSearchHandlerWithMock(t)

	ctx := context.Background()
	q := queries.SearchOrganisationsQuery{Name: utils.Ref("Acme")}

	bundle := &http.OrganizationBundle{
		Total: utils.Ref("not-a-number"),
		Entry: utils.Ref([]http.OrganizationEntry{}),
	}
	mockODS.SearchOrganisationsReturns(bundle, nil)

	_, err := handler.Handle(ctx, q)
	require.Error(t, err)
}
