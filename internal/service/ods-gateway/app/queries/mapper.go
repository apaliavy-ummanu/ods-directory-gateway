package queries

import (
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/common/utils"
	"github.com/Cleo-Systems/ods-directory-gateway/internal/service/ods-gateway/app/domain"
	fhirHTTP "github.com/Cleo-Systems/ods-directory-gateway/pkg/ods-fhir-api/client"
)

const (
	activePeriodURL = "https://fhir.nhs.uk/STU3/StructureDefinition/Extension-ODSAPI-ActivePeriod-1"
	dateTypeURL     = "https://fhir.nhs.uk/STU3/StructureDefinition/Extension-ODSAPI-DateType-1"
	orgRoleURL      = "https://fhir.nhs.uk/STU3/StructureDefinition/Extension-ODSAPI-OrganizationRole-1"
	odsCodeURL      = "https://fhir.nhs.uk/Id/ods-organization-code"
)

func mapOrganisationToDomain(org fhirHTTP.OrganizationResource) domain.Organisation {
	odsCode := org.Id
	if *org.Identifier.System == odsCodeURL {
		odsCode = *org.Identifier.Value
	}

	orgAddress := utils.Deref(org.Address)
	op := getOperationalPeriod(org)

	return domain.Organisation{
		ID:       org.Id,
		OdsCode:  odsCode,
		Name:     org.Name,
		IsActive: utils.Deref(org.Active),
		Metadata: getMetadata(org),
		Address: domain.Address{
			City:       orgAddress.City,
			Country:    orgAddress.Country,
			Lines:      orgAddress.Line,
			PostalCode: orgAddress.PostalCode,
		},
		OperationalPeriod: op,
		RecordClass:       getRecordClassCode(org),
		Roles:             getRoles(org),
	}
}

func getRoles(org fhirHTTP.OrganizationResource) []domain.OrganisationRole {
	var roles []domain.OrganisationRole

	for _, ext := range utils.Deref(org.Extension) {
		// We only care about OrganizationRole extensions
		if utils.Deref(ext.Url) != orgRoleURL {
			continue
		}

		var r domain.OrganisationRole

		for _, inner := range utils.Deref(ext.Extension) {
			switch utils.Deref(inner.Url) {
			case "role":
				if inner.ValueCoding != nil {
					r.Code = utils.Deref(inner.ValueCoding.Code)
					r.Display = utils.Deref(inner.ValueCoding.Display)
				}

			case "primaryRole":
				if inner.ValueBoolean != nil {
					r.Primary = *inner.ValueBoolean
				}

			case "status":
				r.Status = utils.Deref(inner.ValueString)

			case "activePeriod":
				if inner.ValuePeriod != nil {
					p := inner.ValuePeriod
					op := domain.OperationalPeriod{
						Start: utils.Deref(p.Start).Time,
					}
					if p.End != nil && !p.End.IsZero() {
						op.End = utils.Ref(p.End.Time)
					}

					// pick DateType (Operational / Legal etc.)
					for _, pe := range utils.Deref(p.Extension) {
						if utils.Deref(pe.Url) == dateTypeURL && pe.ValueString != nil {
							op.DateType = *pe.ValueString
							break
						}
					}
					r.OperationalPeriod = &op
				}
			}
		}

		// Only include roles that at least have a code
		if r.Code != "" {
			roles = append(roles, r)
		}
	}

	return roles
}

func getMetadata(org fhirHTTP.OrganizationResource) domain.OrganisationMetadata {
	if org.Meta == nil {
		return domain.OrganisationMetadata{}
	}

	return domain.OrganisationMetadata{
		LastUpdated: utils.Deref(org.Meta.LastUpdated),
	}
}

func getRecordClassCode(org fhirHTTP.OrganizationResource) string {
	if org.Type == nil || org.Type.Coding == nil {
		return ""
	}
	return utils.Deref(org.Type.Coding.Code)
}

func getOperationalPeriod(org fhirHTTP.OrganizationResource) *domain.OperationalPeriod {
	for _, ext := range utils.Deref(org.Extension) {
		if utils.Deref(ext.Url) != activePeriodURL || ext.ValuePeriod == nil {
			continue
		}

		p := ext.ValuePeriod
		op := &domain.OperationalPeriod{
			Start: utils.Deref(p.Start).Time,
		}
		if p.End != nil && !p.End.IsZero() {
			op.End = utils.Ref(p.End.Time)
		}

		// Find the DateType extension inside valuePeriod.extension
		for _, inner := range utils.Deref(p.Extension) {
			valueString := utils.Deref(inner.ValueString)
			if utils.Deref(inner.Url) == dateTypeURL && valueString != "" {
				op.DateType = valueString
				break
			}
		}

		// We only want the Operational one
		if op.DateType == "Operational" {
			return op
		}
	}

	// nothing found
	return nil
}
