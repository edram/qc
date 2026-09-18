package regions

import "testing"

func TestLoadQCCRegion(t *testing.T) {
	catalog, err := Load(ProviderQCC)
	if err != nil {
		t.Fatal(err)
	}
	region, err := catalog.Resolve("义乌市")
	if err != nil {
		t.Fatal(err)
	}
	if region.ID != "330782" || region.ProviderCode != "330782" || region.ProvinceCode != "ZJ" || region.Depth != 3 {
		t.Fatalf("region = %#v", region)
	}
}

func TestLoadQCCProviderOnlyRegion(t *testing.T) {
	catalog, err := Load(ProviderQCC)
	if err != nil {
		t.Fatal(err)
	}
	region, err := catalog.Resolve("雄安新区")
	if err != nil {
		t.Fatal(err)
	}
	if region.ProviderCode != "Z027" || region.ProvinceCode != "HB" {
		t.Fatalf("region = %#v", region)
	}
}

func TestLoadQCCCompatibilityRegion(t *testing.T) {
	catalog, err := Load(ProviderQCC)
	if err != nil {
		t.Fatal(err)
	}
	region, err := catalog.Resolve("香港特别行政区")
	if err != nil {
		t.Fatal(err)
	}
	if region.ProviderCode != "HK" || region.ProvinceCode != "HK" {
		t.Fatalf("region = %#v", region)
	}
}
