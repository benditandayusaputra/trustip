package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	authdomain "github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	filedomain "github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
	orgdomain "github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	orgservice "github.com/benditandayusaputra/trustip/api/internal/modules/org/service"
	productdomain "github.com/benditandayusaputra/trustip/api/internal/modules/product/domain"
	productservice "github.com/benditandayusaputra/trustip/api/internal/modules/product/service"
)

const (
	setupHour      = 8
	demoMailDomain = "@demo.trustip.id"
)

func (s *demoSeeder) loadCategories(ctx context.Context) error {
	categories, err := s.use.org.Categories.Execute(ctx)
	if err != nil {
		return fmt.Errorf("list categories: %w", err)
	}
	for _, category := range categories {
		s.cast.categories[category.Name] = category.ID
	}
	return nil
}

func (s *demoSeeder) category(name string) (string, error) {
	id, ok := s.cast.categories[name]
	if !ok {
		return "", fmt.Errorf("unknown demo category %q", name)
	}
	return id, nil
}

func (s *demoSeeder) foundOrganizations(ctx context.Context) error {
	s.moveTo(s.dayAt(s.world.SetupDaysAgo, setupHour))
	for index, organization := range s.world.Organizations {
		s.clock.Advance(time.Minute)
		owner, err := s.register(ctx, organization, index)
		if err != nil {
			return err
		}
		if owner, err = s.found(ctx, owner, organization, index); err != nil {
			return err
		}
		if err := s.describe(ctx, owner, organization); err != nil {
			return err
		}
		if err := s.showLogo(ctx, owner, organization, index); err != nil {
			return err
		}
		s.cast.owners[organization.Key] = owner
	}
	return nil
}

func demoEmail(organization demoOrganization) string {
	if organization.Email != "" {
		return organization.Email
	}
	return strings.ReplaceAll(organization.Key, "_", ".") + demoMailDomain
}

func demoWhatsApp(index int) string {
	return fmt.Sprintf("08123450%04d", index+1)
}

func (s *demoSeeder) register(ctx context.Context, organization demoOrganization, index int) (actor.Actor, error) {
	registration, err := authdomain.NewRegistration(authdomain.RegistrationForm{
		FullName: organization.Owner, Email: demoEmail(organization), Password: s.world.Password,
		WhatsApp: demoWhatsApp(index), PrivacyAccepted: true,
	}, authdomain.NewPasswordPolicy())
	if err != nil {
		return actor.Actor{}, fmt.Errorf("validate demo user %s: %w", organization.Key, err)
	}
	password, err := s.hasher.Hash(registration.Password)
	if err != nil {
		return actor.Actor{}, err
	}
	userID, err := newDemoKey()
	if err != nil {
		return actor.Actor{}, err
	}
	if err := s.users.CreateUser(ctx, registration.Complete(userID, password, s.clock.Now())); err != nil {
		return actor.Actor{}, fmt.Errorf("create demo user %s: %w", organization.Key, err)
	}
	return actor.Actor{UserID: userID}, nil
}

func (s *demoSeeder) found(
	ctx context.Context, newcomer actor.Actor, organization demoOrganization, index int,
) (actor.Actor, error) {
	created, err := s.use.org.Create.Execute(ctx, newcomer, orgdomain.OnboardingForm{
		Type: organization.Type, ProfileForm: s.profileForm(organization, index),
	})
	if err != nil {
		return actor.Actor{}, fmt.Errorf("found demo organization %s: %w", organization.Key, err)
	}
	err = s.queries(ctx).SetOrganizationDemo(ctx, sqlc.SetOrganizationDemoParams{ID: created.ID, IsDemo: true})
	if err != nil {
		return actor.Actor{}, fmt.Errorf("mark %s as demo: %w", organization.Key, err)
	}
	newcomer.Membership = actor.Membership{
		OrganizationID: created.ID, OrganizationType: actor.OrganizationType(organization.Type), Role: actor.Owner,
	}
	return newcomer, nil
}

func (s *demoSeeder) profileForm(organization demoOrganization, index int) orgdomain.ProfileForm {
	point := s.world.locate(organization)
	return orgdomain.ProfileForm{
		Name: organization.Name, Description: organization.Description, WhatsApp: demoWhatsApp(index),
		CityCode: firstOf(organization.CityCode, s.world.CityCode), Address: organization.Address,
		District: firstOf(organization.District, s.world.District),
		Location: orgdomain.LocationInput{Given: true, Latitude: point.Latitude, Longitude: point.Longitude},
	}
}

func (s *demoSeeder) showLogo(
	ctx context.Context, owner actor.Actor, organization demoOrganization, index int,
) error {
	logoID, err := s.photos.put(ctx, owner, filedomain.Logo, organizationPhotos+organization.Key)
	if err != nil || logoID == "" {
		return err
	}
	_, err = s.use.org.EditProfile.Execute(ctx, owner, orgservice.UpdateProfileCommand{
		OrganizationID: owner.Membership.OrganizationID,
		Form: orgdomain.ProfileUpdateForm{
			ProfileForm: s.profileForm(organization, index), LogoFileID: logoID,
		},
	})
	if err != nil {
		return fmt.Errorf("show logo of demo organization %s: %w", organization.Key, err)
	}
	return nil
}

func firstOf(preferred, fallback string) string {
	if preferred != "" {
		return preferred
	}
	return fallback
}

func (s *demoSeeder) describe(ctx context.Context, owner actor.Actor, organization demoOrganization) error {
	organizationID := owner.Membership.OrganizationID
	var err error
	if organization.Store != nil {
		var form orgdomain.StoreProfileForm
		if form, err = s.storeForm(*organization.Store); err == nil {
			form.PhotoFileIDs, err = s.storePhotos(ctx, owner, organization.Key)
		}
		if err == nil {
			_, err = s.use.org.EditStore.Execute(ctx, owner, orgservice.UpdateStoreProfileCommand{
				OrganizationID: organizationID, Form: form,
			})
		}
	}
	if organization.Consignor != nil {
		err = s.describeConsignor(ctx, owner, *organization.Consignor)
	}
	if err != nil {
		return fmt.Errorf("describe demo organization %s: %w", organization.Key, err)
	}
	return nil
}

func (s *demoSeeder) describeConsignor(ctx context.Context, owner actor.Actor, profile demoConsignor) error {
	organizationID := owner.Membership.OrganizationID
	_, err := s.use.org.EditConsignor.Execute(ctx, owner, orgservice.UpdateConsignorProfileCommand{
		OrganizationID: organizationID, Form: orgdomain.ConsignorProfileForm{
			DeliveryRadius:      orgdomain.DeliveryRadius{Given: true, Kilometres: profile.DeliveryRadiusKm},
			DeliveryDays:        profile.DeliveryDays,
			PreferredStoreTypes: profile.PreferredStoreTypes, CanProvideSample: profile.CanProvideSample,
		},
	})
	if err != nil {
		return fmt.Errorf("edit consignor profile: %w", err)
	}
	if profile.PaymentInfo == "" {
		return nil
	}
	_, err = s.use.org.EditPayment.Execute(ctx, owner, orgservice.UpdatePaymentInfoCommand{
		OrganizationID: organizationID, Form: orgdomain.PaymentInfoForm{Info: profile.PaymentInfo},
	})
	if err != nil {
		return fmt.Errorf("edit payment info: %w", err)
	}
	return nil
}

func (s *demoSeeder) storePhotos(ctx context.Context, owner actor.Actor, key string) ([]string, error) {
	photoID, err := s.photos.put(ctx, owner, filedomain.StorePhoto, organizationPhotos+key)
	if err != nil || photoID == "" {
		return nil, err
	}
	return []string{photoID}, nil
}

func (s *demoSeeder) storeForm(store demoStore) (orgdomain.StoreProfileForm, error) {
	categoryIDs := make([]string, 0, len(store.Categories))
	for _, name := range store.Categories {
		id, err := s.category(name)
		if err != nil {
			return orgdomain.StoreProfileForm{}, err
		}
		categoryIDs = append(categoryIDs, id)
	}
	return orgdomain.StoreProfileForm{
		StoreType: store.StoreType, CategoryIDs: categoryIDs, OpeningHours: store.OpeningHours,
		Acceptance: orgdomain.Acceptance{AcceptsFixedDeposit: store.FixedDeposit, AcceptsCommission: store.Commission},
		Commission: orgdomain.CommissionRange{
			Given: store.Commission, MinBPS: store.CommissionMinBPS, MaxBPS: store.CommissionMaxBPS,
		},
	}, nil
}

func (s *demoSeeder) createProducts(ctx context.Context) error {
	for _, product := range s.world.Products {
		if err := s.createProduct(ctx, product); err != nil {
			return fmt.Errorf("create demo product %s: %w", product.Key, err)
		}
	}
	for key, owner := range s.cast.owners {
		_, err := s.use.org.EditListing.Execute(ctx, owner, orgservice.SetListingCommand{
			OrganizationID: owner.Membership.OrganizationID, IsListed: true,
		})
		if err != nil {
			return fmt.Errorf("list demo organization %s: %w", key, err)
		}
	}
	return nil
}

func (s *demoSeeder) createProduct(ctx context.Context, product demoProduct) error {
	categoryID, err := s.category(product.Category)
	if err != nil {
		return err
	}
	owner := s.cast.owners[product.Owner]
	photoID, err := s.photos.put(ctx, owner, filedomain.ProductPhoto, productPhotos+product.Key)
	if err != nil {
		return err
	}
	created, err := s.use.product.Create.Execute(ctx, owner, productdomain.ProductForm{
		Name: product.Name, CategoryID: categoryID, Unit: product.Unit, NetContent: product.NetContent,
		PhotoFileID: photoID,
		Prices:      productdomain.SuggestedPrices{Retail: product.Retail, Deposit: product.Deposit},
		ShelfLife:   productdomain.ShelfLife{Given: product.ShelfLifeDays > 0, Days: product.ShelfLifeDays},
	})
	if err != nil {
		return fmt.Errorf("create: %w", err)
	}
	s.cast.products[product.Key] = created.ID
	if !product.Seeking {
		return nil
	}
	_, err = s.use.product.SetSeeking.Execute(ctx, owner, productservice.SetSeekingCommand{
		ProductID: created.ID, IsSeeking: true,
	})
	if err != nil {
		return fmt.Errorf("seek stores: %w", err)
	}
	return nil
}
