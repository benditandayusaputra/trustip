package domain

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	MatchingRadiusMeters = 10_000
	metersPerKm          = 1000
	tenthsPerKm          = 10
)

type Notice struct {
	OrganizationID string
	ActorUserID    string
	Type           string
	Title          string
	Body           string
	Link           string
	OwnersOnly     bool
}

type Interested struct {
	OrganizationID string
	DistanceMeters int
}

func (n Need) PostedNotice(interested Interested, actorUserID string) Notice {
	query := url.Values{"sasaran": {n.StoreID}, "kebutuhan": {n.ID}}
	return Notice{
		OrganizationID: interested.OrganizationID, ActorUserID: actorUserID, Type: "NEED_POSTED",
		Title: "Kebutuhan toko baru",
		Body: fmt.Sprintf("%s butuh %s, %s dari Anda.", n.StoreName, strings.ToLower(n.CategoryName),
			FormatDistance(interested.DistanceMeters)),
		Link: "/app/ajuan/baru?" + query.Encode(), OwnersOnly: true,
	}
}

type ReminderWindow struct {
	From  time.Time
	Until time.Time
}

func ReminderWindowAt(now time.Time) ReminderWindow {
	return ReminderWindow{From: now, Until: now.Add(ExpiryReminderBefore)}
}

type Expiring struct {
	NeedID  string
	StoreID string
	Title   string
}

func (e Expiring) Notice() Notice {
	return Notice{
		OrganizationID: e.StoreID, Type: "NEED_EXPIRING", Title: "Kebutuhan segera berakhir",
		Body: fmt.Sprintf("Kebutuhan '%s' berakhir 3 hari lagi.", e.Title), Link: "/app/kebutuhan/" + e.NeedID,
		OwnersOnly: true,
	}
}

func FormatDistance(meters int) string {
	if meters < metersPerKm {
		return strconv.Itoa(meters) + " m"
	}
	tenths := (meters*tenthsPerKm + metersPerKm/2) / metersPerKm
	whole, fraction := tenths/tenthsPerKm, tenths%tenthsPerKm
	if fraction == 0 {
		return strconv.Itoa(whole) + " km"
	}
	return strconv.Itoa(whole) + "," + strconv.Itoa(fraction) + " km"
}
