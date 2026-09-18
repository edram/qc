package qcc

import "github.com/edram/qi/internal/regions"

// Area identifies a QCC region and the province filter that contains it.
type Area struct {
	// ProvinceCode is QCC's pr filter value.
	ProvinceCode string
	// Code is the selected terminal code and equals ProvinceCode for province-level areas.
	Code string
}

// ResolveArea resolves a full path, unique name, or code from QCC's region catalog.
func ResolveArea(value string) (Area, error) {
	region, err := qccRegions().Resolve(value)
	if err != nil {
		return Area{}, err
	}
	return Area{ProvinceCode: region.ProvinceCode, Code: region.ProviderCode}, nil
}

func qccRegions() *regions.Catalog {
	catalog, err := regions.Load(regions.ProviderQCC)
	if err != nil {
		panic(err)
	}
	return catalog
}
