package queries

import (
	"github.com/Cleo-Systems/ods-fhir-gateway/internal/service/common/utils"
	"github.com/Cleo-Systems/ods-fhir-gateway/internal/service/ods-gateway/app/domain"
	fhirHTTP "github.com/Cleo-Systems/ods-fhir-gateway/pkg/ods-fhir-api/client"
)

const (
	ActivePeriodURL = "https://fhir.nhs.uk/STU3/StructureDefinition/Extension-ODSAPI-ActivePeriod-1"
	DateTypeURL     = "https://fhir.nhs.uk/STU3/StructureDefinition/Extension-ODSAPI-DateType-1"
	OrgRoleURL      = "https://fhir.nhs.uk/STU3/StructureDefinition/Extension-ODSAPI-OrganizationRole-1"
	ODSCodeURL      = "https://fhir.nhs.uk/Id/ods-organization-code"
)

const (
	ExtensionRole         = "role"
	ExtensionPrimaryRole  = "primaryRole"
	ExtensionStatus       = "status"
	ExtensionActivePeriod = "activePeriod"
)

func mapOrganisationToDomain(org fhirHTTP.OrganizationResource) domain.Organisation {
	odsCode := org.Id
	if *org.Identifier.System == ODSCodeURL {
		odsCode = *org.Identifier.Value
	}

	orgAddress := utils.Deref(org.Address)
	op := getOperationalPeriod(org)

	return domain.Organisation{
		ID:       org.Id,
		ODSCode:  odsCode,
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
		// we only care about OrganizationRole extensions
		if utils.Deref(ext.Url) != OrgRoleURL {
			continue
		}

		var r domain.OrganisationRole

		for _, inner := range utils.Deref(ext.Extension) {
			switch utils.Deref(inner.Url) {
			case ExtensionRole:
				if inner.ValueCoding != nil {
					r.Code = utils.Deref(inner.ValueCoding.Code)
					r.Display = utils.Deref(inner.ValueCoding.Display)
				}

			case ExtensionPrimaryRole:
				if inner.ValueBoolean != nil {
					r.Primary = *inner.ValueBoolean
				}

			case ExtensionStatus:
				r.Status = utils.Deref(inner.ValueString)

			case ExtensionActivePeriod:
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
						if utils.Deref(pe.Url) == DateTypeURL && pe.ValueString != nil {
							op.DateType = *pe.ValueString
							break
						}
					}
					r.OperationalPeriod = &op
				}
			}
		}

		// only include roles that at least have a code
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
		if utils.Deref(ext.Url) != ActivePeriodURL || ext.ValuePeriod == nil {
			continue
		}

		p := ext.ValuePeriod
		op := &domain.OperationalPeriod{
			Start: utils.Deref(p.Start).Time,
		}
		if p.End != nil && !p.End.IsZero() {
			op.End = utils.Ref(p.End.Time)
		}

		// find the DateType extension inside valuePeriod.extension
		for _, inner := range utils.Deref(p.Extension) {
			valueString := utils.Deref(inner.ValueString)
			if utils.Deref(inner.Url) == DateTypeURL && valueString != "" {
				op.DateType = valueString
				break
			}
		}

		// we only want the Operational one
		if op.DateType == "Operational" {
			return op
		}
	}

	// nothing found
	return nil
}
