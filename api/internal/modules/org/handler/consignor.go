package handler

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
	"github.com/benditandayusaputra/trustip/api/internal/platform/httpserver"
)

type ConsignorProfileEditor interface {
	Execute(
		ctx context.Context, signedIn actor.Actor, cmd service.UpdateConsignorProfileCommand,
	) (domain.Settings, error)
}

type PaymentInfoEditor interface {
	Execute(ctx context.Context, signedIn actor.Actor, cmd service.UpdatePaymentInfoCommand) (domain.Settings, error)
}

type consignorProfileBody struct {
	DeliveryRadiusKM     *int     `json:"delivery_radius_km"`
	DeliveryDays         []string `json:"delivery_days"`
	PreferredStoreTypes  []string `json:"preferred_store_types"`
	BusinessIDNo         string   `json:"business_id_no"`
	DistributionPermitNo string   `json:"distribution_permit_no"`
	HalalCertNo          string   `json:"halal_cert_no"`
	CanProvideSample     bool     `json:"can_provide_sample"`
}

type paymentInfoRequest struct {
	PaymentInfo     string  `json:"payment_info"`
	PaymentQRFileID *string `json:"payment_qr_file_id"`
}

type paymentResponse struct {
	Info string    `json:"info"`
	QR   *imageRef `json:"qr"`
}

func updateConsignorProfile(editor ConsignorProfileEditor, filesBaseURL string) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, organizationID, err := actorAndOrganization(c)
		if err != nil {
			return err
		}
		var request consignorProfileBody
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read consignor profile: %w", err)
		}
		form, err := request.toForm()
		if err != nil {
			return err
		}
		settings, err := editor.Execute(c.Context(), signedIn,
			service.UpdateConsignorProfileCommand{OrganizationID: organizationID, Form: form})
		if err != nil {
			return fmt.Errorf("update consignor profile: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toSettingsResponse(settings, filesBaseURL)})
	}
}

func (r consignorProfileBody) toForm() (domain.ConsignorProfileForm, error) {
	inputs := []httpserver.TextInput{
		httpserver.LineInput("business_id_no", r.BusinessIDNo),
		httpserver.LineInput("distribution_permit_no", r.DistributionPermitNo),
		httpserver.LineInput("halal_cert_no", r.HalalCertNo),
	}
	if err := httpserver.RejectUncleanInputs(inputs...); err != nil {
		return domain.ConsignorProfileForm{}, fmt.Errorf("check consignor profile text: %w", err)
	}
	radius := domain.DeliveryRadius{}
	if r.DeliveryRadiusKM != nil {
		radius = domain.DeliveryRadius{Given: true, Kilometres: *r.DeliveryRadiusKM}
	}
	return domain.ConsignorProfileForm{
		DeliveryRadius: radius, DeliveryDays: orEmpty(r.DeliveryDays),
		PreferredStoreTypes: orEmpty(r.PreferredStoreTypes),
		Legal: domain.LegalNumbers{
			BusinessIDNo: inputs[0].Value, DistributionPermitNo: inputs[1].Value, HalalCertNo: inputs[2].Value,
		},
		CanProvideSample: r.CanProvideSample,
	}, nil
}

func updatePaymentInfo(editor PaymentInfoEditor, filesBaseURL string) fiber.Handler {
	return func(c fiber.Ctx) error {
		signedIn, organizationID, err := actorAndOrganization(c)
		if err != nil {
			return err
		}
		var request paymentInfoRequest
		if err := httpserver.DecodeJSON(c, &request); err != nil {
			return fmt.Errorf("read payment info: %w", err)
		}
		info := httpserver.MultilineInput("payment_info", request.PaymentInfo)
		if err := httpserver.RejectUncleanInputs(info); err != nil {
			return fmt.Errorf("check payment info text: %w", err)
		}
		form := domain.PaymentInfoForm{Info: info.Value, QRFileID: valueOrEmpty(request.PaymentQRFileID)}
		settings, err := editor.Execute(c.Context(), signedIn,
			service.UpdatePaymentInfoCommand{OrganizationID: organizationID, Form: form})
		if err != nil {
			return fmt.Errorf("update payment info: %w", err)
		}
		return httpserver.Respond(c, httpserver.Reply{Data: toSettingsResponse(settings, filesBaseURL)})
	}
}

func toConsignorProfileResponse(profile domain.ConsignorProfile) *consignorProfileBody {
	response := &consignorProfileBody{
		DeliveryDays: orEmpty(profile.DeliveryDays), PreferredStoreTypes: orEmpty(profile.PreferredStoreTypes),
		BusinessIDNo: profile.Legal.BusinessIDNo, DistributionPermitNo: profile.Legal.DistributionPermitNo,
		HalalCertNo: profile.Legal.HalalCertNo, CanProvideSample: profile.CanProvideSample,
	}
	if profile.DeliveryRadius.Given {
		response.DeliveryRadiusKM = &profile.DeliveryRadius.Kilometres
	}
	return response
}

func toPaymentResponse(payment domain.PaymentSettings) *paymentResponse {
	response := &paymentResponse{Info: payment.Info}
	if payment.QR.ID != "" {
		response.QR = &imageRef{ID: payment.QR.ID, URL: httpserver.APIPrefix + "/files/" + payment.QR.ID}
	}
	return response
}
