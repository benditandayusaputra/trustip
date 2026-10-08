package domain

import "testing"

func TestFormatDistance_ReadsLikeTheApp(t *testing.T) {
	cases := map[int]string{800: "800 m", 1200: "1,2 km", 1249: "1,2 km", 10000: "10 km", 9960: "10 km"}
	for meters, want := range cases {
		if got := FormatDistance(meters); got != want {
			t.Errorf("FormatDistance(%d) = %q, want %q", meters, got, want)
		}
	}
}

func TestPostedNotice_FRNED07_TellsTheConsignorWhatTheStoreNeedsAndHowFar(t *testing.T) {
	need := openNeed()
	need.StoreName, need.CategoryName = "Kantin Kampus", "Kue kering dan camilan"

	notice := need.PostedNotice(Interested{OrganizationID: "c1", DistanceMeters: 1200}, "u1")

	if notice.Type != "NEED_POSTED" || notice.OrganizationID != "c1" || !notice.OwnersOnly ||
		notice.Body != "Kantin Kampus butuh kue kering dan camilan, 1,2 km dari Anda." ||
		notice.Link != "/app/ajuan/baru?kebutuhan=n1&sasaran=s1" {
		t.Fatalf("notice = %+v", notice)
	}
}

func TestExpiringNotice_BRNED03_RemindsTheStoreThreeDaysBefore(t *testing.T) {
	notice := Expiring{NeedID: "n1", StoreID: "s1", Title: "Butuh roti"}.Notice()

	if notice.Type != "NEED_EXPIRING" || notice.OrganizationID != "s1" ||
		notice.Body != "Kebutuhan 'Butuh roti' berakhir 3 hari lagi." || notice.Link != "/app/kebutuhan/n1" {
		t.Fatalf("notice = %+v", notice)
	}
}
