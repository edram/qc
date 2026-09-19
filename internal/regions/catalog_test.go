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

func TestSearchQCCRegions(t *testing.T) {
	catalog, err := Load(ProviderQCC)
	if err != nil {
		t.Fatal(err)
	}
	result := catalog.Search("深圳", 5)
	if len(result) != 1 {
		t.Fatalf("Search() results = %#v, want one result", result)
	}
	if result[0].Name != "深圳市" || result[0].ProviderCode != "440300" {
		t.Fatalf("Search() result = %#v", result[0])
	}
}

func TestTopLevelQCCRegions(t *testing.T) {
	catalog, err := Load(ProviderQCC)
	if err != nil {
		t.Fatal(err)
	}
	result := catalog.TopLevel(2)
	if len(result) != 2 || result[0].Name != "北京市" || result[1].Name != "天津市" {
		t.Fatalf("TopLevel() results = %#v", result)
	}
}
