package service_test

import (
	"testing"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
)

func consignorMember(role actor.Role) actor.Actor {
	member := storeMember(role)
	member.Membership.OrganizationType = actor.Consignor
	return member
}

func TestUpdateConsignorProfile_FRORG05_OwnerSavesTheProfile(t *testing.T) {
	sections := &fakeStoreProfiles{}
	form := domain.ConsignorProfileForm{
		DeliveryRadius: domain.DeliveryRadius{Given: true, Kilometres: 15}, DeliveryDays: []string{"FRI"},
	}
	command := service.UpdateConsignorProfileCommand{OrganizationID: ownOrganizationID, Form: form}

	_, err := service.NewUpdateConsignorProfile(sectionDeps(sections)).Execute(t.Context(),
		consignorMember(actor.Owner), command)

	if err != nil || len(sections.consignors) != 1 || sections.consignors[0].Profile.DeliveryRadius.Kilometres != 15 {
		t.Fatalf("err = %v, changes = %+v", err, sections.consignors)
	}
}

func TestUpdateConsignorProfile_StoreHasNoConsignorProfile(t *testing.T) {
	_, err := service.NewUpdateConsignorProfile(sectionDeps(&fakeStoreProfiles{})).Execute(t.Context(),
		storeMember(actor.Owner), service.UpdateConsignorProfileCommand{OrganizationID: ownOrganizationID})

	if apperror.From(err).Code() != apperror.CodeNotFound {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdatePaymentInfo_FRORG07_StaffIsForbidden(t *testing.T) {
	_, err := service.NewUpdatePaymentInfo(sectionDeps(&fakeStoreProfiles{})).Execute(t.Context(),
		consignorMember(actor.Staff), service.UpdatePaymentInfoCommand{OrganizationID: ownOrganizationID})

	if apperror.From(err).Code() != apperror.CodeForbidden {
		t.Fatalf("err = %v", err)
	}
}

func TestUpdatePaymentInfo_UnknownQRIsAFieldError(t *testing.T) {
	sections := &fakeStoreProfiles{failure: domain.ErrUnknownQR}

	_, err := service.NewUpdatePaymentInfo(sectionDeps(sections)).Execute(t.Context(), consignorMember(actor.Owner),
		service.UpdatePaymentInfoCommand{OrganizationID: ownOrganizationID, Form: domain.PaymentInfoForm{Info: "BCA"}})

	fields := apperror.From(err).Fields()
	if len(fields) != 1 || fields[0].Field != "payment_qr_file_id" {
		t.Fatalf("fields = %+v", fields)
	}
}

func TestUpdatePaymentInfo_FRORG07_OwnerSavesTheInfo(t *testing.T) {
	sections := &fakeStoreProfiles{}

	_, err := service.NewUpdatePaymentInfo(sectionDeps(sections)).Execute(t.Context(), consignorMember(actor.Owner),
		service.UpdatePaymentInfoCommand{OrganizationID: ownOrganizationID, Form: domain.PaymentInfoForm{Info: "BCA"}})

	if err != nil || len(sections.payments) != 1 || sections.payments[0].Payment.Info != "BCA" {
		t.Fatalf("err = %v, payments = %+v", err, sections.payments)
	}
}
