package domain

import "testing"

func TestNormalizeCode_BRSIG05_AcceptsLowercaseAndMissingDashes(t *testing.T) {
	for raw, want := range map[string]string{
		"7VAK-4V4R-2EXS-P3N6":   "7VAK-4V4R-2EXS-P3N6",
		"7vak4v4r2exsp3n6":      "7VAK-4V4R-2EXS-P3N6",
		" 7vak 4v4r-2exs p3n6 ": "7VAK-4V4R-2EXS-P3N6",
	} {
		if got, ok := NormalizeCode(raw); !ok || got != want {
			t.Errorf("NormalizeCode(%q) = %q, %t", raw, got, ok)
		}
	}
}

func TestNormalizeCode_WrongLengthOrCharacters_IsRejected(t *testing.T) {
	for _, raw := range []string{
		"", "7VAK-4V4R-2EXS-P3N", "7VAK-4V4R-2EXS-P3N61", "7VAK-4V4R-2EXS-P3N1", "7VAK-4V4R-2EXS-P3N!",
	} {
		if _, ok := NormalizeCode(raw); ok {
			t.Errorf("NormalizeCode(%q) accepted", raw)
		}
	}
}

func signedByPlatform(Seal) bool { return true }

func TestJudgePublic_FRSIG03_IntactLinkedSignedSealIsValid(t *testing.T) {
	chain := chainOf(t, 2)

	verdict := JudgePublic(PublicSeal{Seal: chain[1], PreviousEntryHash: chain[0].EntryHash}, signedByPlatform)

	if !verdict.IsValid {
		t.Fatal("intact seal judged invalid")
	}
}

func TestJudgePublic_RSEC33_ChangedEntryIsInvalid(t *testing.T) {
	chain := chainOf(t, 1)
	changed := chain[0]
	changed.Entry.Document.Number = "TM-202610-9999"

	if JudgePublic(PublicSeal{Seal: changed}, signedByPlatform).IsValid {
		t.Fatal("changed entry judged valid")
	}
}

func TestJudgePublic_BRSIG02_BrokenLinkIsInvalid(t *testing.T) {
	chain := chainOf(t, 2)

	verdict := JudgePublic(PublicSeal{Seal: chain[1], PreviousEntryHash: chain[1].EntryHash}, signedByPlatform)

	if verdict.IsValid {
		t.Fatal("seal with a broken link judged valid")
	}
}

func TestJudgePublic_UnsignedSealIsInvalid(t *testing.T) {
	chain := chainOf(t, 1)

	if JudgePublic(PublicSeal{Seal: chain[0]}, func(Seal) bool { return false }).IsValid {
		t.Fatal("unsigned seal judged valid")
	}
}

func TestIsIntactChain_FRSIG05_IntactSignedChainPasses(t *testing.T) {
	if !IsIntactChain(chainOf(t, 3), signedByPlatform) {
		t.Fatal("intact chain judged broken")
	}
}

func TestIsIntactChain_FRSIG05_MissingLinkOrSignatureFails(t *testing.T) {
	chain := chainOf(t, 3)

	if IsIntactChain([]Seal{chain[0], chain[2]}, signedByPlatform) {
		t.Fatal("chain with a missing seal judged intact")
	}
	if IsIntactChain(chain, func(Seal) bool { return false }) {
		t.Fatal("unsigned chain judged intact")
	}
}
