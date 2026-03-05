package service

import "testing"

func TestMergeDistributorOffer_ChoosesLowerCostAndKeepsExistingText(t *testing.T) {
	existing := DistributorOffer{
		Name:         "Unlimited",
		CostCNYCents: 8000,
		Enabled:      true,
		Notes:        "existing notes",
	}

	name, cost, enabled, notes, updated := mergeDistributorOffer(existing, "Unlimited", 6000, false, "import notes")
	if !updated {
		t.Fatalf("expected update")
	}
	if name != "Unlimited" {
		t.Fatalf("expected name unchanged, got %q", name)
	}
	if cost != 6000 {
		t.Fatalf("expected lower cost selected, got %d", cost)
	}
	if !enabled {
		t.Fatalf("expected enabled to remain true")
	}
	if notes != "existing notes" {
		t.Fatalf("expected existing notes kept, got %q", notes)
	}
}

func TestMergeDistributorOffer_FillsMissingFieldsFromIncoming(t *testing.T) {
	existing := DistributorOffer{
		Name:         "",
		CostCNYCents: 0,
		Enabled:      false,
		Notes:        "",
	}

	name, cost, enabled, notes, updated := mergeDistributorOffer(existing, "Starter", 1200, true, "new notes")
	if !updated {
		t.Fatalf("expected update")
	}
	if name != "Starter" {
		t.Fatalf("expected incoming name used, got %q", name)
	}
	if cost != 1200 {
		t.Fatalf("expected incoming cost used, got %d", cost)
	}
	if !enabled {
		t.Fatalf("expected enabled true")
	}
	if notes != "new notes" {
		t.Fatalf("expected incoming notes used, got %q", notes)
	}
}

func TestIsForeignKeyConstraintError(t *testing.T) {
	if !isForeignKeyConstraintError(assertErr("pq: update or delete on table \"distributor_offers\" violates foreign key constraint")) {
		t.Fatalf("expected foreign key error detected")
	}
	if isForeignKeyConstraintError(assertErr("some unrelated error")) {
		t.Fatalf("expected unrelated error not detected")
	}
}

type assertErr string

func (e assertErr) Error() string { return string(e) }
