package industries

import "testing"

func TestLoadAndResolveQCC(t *testing.T) {
	catalog, err := Load(ProviderQCC)
	if err != nil {
		t.Fatal(err)
	}
	industry, err := catalog.Resolve("农业")
	if err != nil {
		t.Fatal(err)
	}
	if industry.ID != "A/01" || industry.ProviderCode != "01" || industry.Depth != 2 {
		t.Fatalf("industry = %#v", industry)
	}
}

func TestSearchAiqicha(t *testing.T) {
	catalog, err := Load(ProviderAiqicha)
	if err != nil {
		t.Fatal(err)
	}
	result := catalog.Search("软件", 10)
	if len(result) == 0 {
		t.Fatal("Search() returned no results")
	}
	found := false
	for _, industry := range result {
		if industry.ProviderCode == "65" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Search() results = %#v, want code 65", result)
	}
}
