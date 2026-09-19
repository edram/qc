package regions

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/edram/qi/internal/data"
)

type Provider string

const (
	ProviderQCC     Provider = "qcc"
	ProviderAiqicha Provider = "aiqicha"
)

var ErrProviderRequired = errors.New("region provider is required")

type Region struct {
	ID           string
	Name         string
	Depth        int
	ParentID     string
	ProviderCode string
	ParentCode   string
	ProvinceCode string
}

type Catalog struct {
	provider Provider
	byPath   map[string]Region
	byName   map[string][]Region
	byCode   map[string]Region
	all      []Region
}

type rawRegion struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Depth    int               `json:"depth"`
	Codes    map[string]string `json:"codes"`
	Children []rawRegion       `json:"children"`
}

type rawProviderRegion struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	Depth      int    `json:"depth"`
	ParentCode string `json:"parentCode"`
}

type rawCatalog struct {
	Regions      []rawRegion                    `json:"regions"`
	ProviderOnly map[string][]rawProviderRegion `json:"providerOnly"`
}

var (
	rawCatalogOnce sync.Once
	rawCatalogData rawCatalog
	rawCatalogErr  error
	catalogs       sync.Map
)

type catalogState struct {
	once    sync.Once
	catalog *Catalog
	err     error
}

func Load(provider Provider) (*Catalog, error) {
	if provider == "" {
		return nil, ErrProviderRequired
	}
	stateValue, _ := catalogs.LoadOrStore(provider, &catalogState{})
	state := stateValue.(*catalogState)
	state.once.Do(func() {
		rawCatalogOnce.Do(func() {
			rawCatalogErr = json.Unmarshal(data.Regions, &rawCatalogData)
		})
		if rawCatalogErr != nil {
			state.err = fmt.Errorf("load region catalog: %w", rawCatalogErr)
			return
		}
		state.catalog = buildCatalog(provider, rawCatalogData)
	})
	return state.catalog, state.err
}

func (catalog *Catalog) Resolve(value string) (Region, error) {
	providerName := catalog.providerName()
	value = strings.Join(strings.Fields(value), " ")
	if region, ok := catalog.byPath[value]; ok {
		return region, nil
	}
	if region, ok := catalog.byCode[strings.ToUpper(value)]; ok {
		return region, nil
	}
	candidates := catalog.byName[value]
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) > 1 {
		return Region{}, fmt.Errorf("ambiguous %s area %q; use a full area path", providerName, value)
	}
	if suggestions := areaSuggestions(catalog.byName, value, 3); len(suggestions) > 0 {
		quoted := make([]string, len(suggestions))
		for i, suggestion := range suggestions {
			quoted[i] = strconv.Quote(suggestion)
		}
		return Region{}, fmt.Errorf("unsupported %s area %q; did you mean %s?", providerName, value, strings.Join(quoted, ", "))
	}
	return Region{}, fmt.Errorf("unsupported %s area %q; use an area name, full path, or code", providerName, value)
}

// Search returns regions whose name contains query.
func (catalog *Catalog) Search(query string, limit int) []Region {
	query = strings.ToLower(strings.Join(strings.Fields(query), " "))
	if query == "" || limit == 0 {
		return nil
	}
	result := make([]Region, 0)
	for _, region := range catalog.all {
		if strings.Contains(strings.ToLower(region.Name), query) {
			result = append(result, region)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Depth != result[j].Depth {
			return result[i].Depth < result[j].Depth
		}
		if result[i].Name != result[j].Name {
			return result[i].Name < result[j].Name
		}
		return result[i].ProviderCode < result[j].ProviderCode
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

// TopLevel returns the provider-supported first-level regions.
func (catalog *Catalog) TopLevel(limit int) []Region {
	result := make([]Region, 0)
	for _, region := range catalog.all {
		if region.Depth == 1 {
			result = append(result, region)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func (catalog *Catalog) providerName() string {
	if catalog.provider == ProviderQCC {
		return "QCC"
	}
	return string(catalog.provider)
}

func buildCatalog(provider Provider, raw rawCatalog) *Catalog {
	catalog := &Catalog{
		provider: provider,
		byPath:   make(map[string]Region),
		byName:   make(map[string][]Region),
		byCode:   make(map[string]Region),
	}
	var add func(rawRegion, string, Region)
	add = func(node rawRegion, parentPath string, parent Region) {
		code := node.Codes[string(provider)]
		path := parentPath
		region := parent
		if code != "" {
			provinceCode := parent.ProvinceCode
			if provinceCode == "" {
				provinceCode = code
			}
			path = node.Name
			if parentPath != "" {
				path = parentPath + " " + node.Name
			}
			region = Region{
				ID:           node.ID,
				Name:         node.Name,
				Depth:        node.Depth,
				ParentID:     parent.ID,
				ProviderCode: code,
				ParentCode:   parent.ProviderCode,
				ProvinceCode: provinceCode,
			}
			catalog.add(path, region)
		}
		for _, child := range node.Children {
			add(child, path, region)
		}
	}
	for _, node := range raw.Regions {
		add(node, "", Region{})
	}
	for _, node := range raw.ProviderOnly[string(provider)] {
		parent := catalog.byCode[strings.ToUpper(node.ParentCode)]
		provinceCode := parent.ProvinceCode
		if provinceCode == "" {
			provinceCode = node.ParentCode
			if provinceCode == "" {
				provinceCode = node.Code
			}
		}
		catalog.add(node.Name, Region{
			ID:           node.ID,
			Name:         node.Name,
			Depth:        node.Depth,
			ParentID:     parent.ID,
			ProviderCode: node.Code,
			ParentCode:   node.ParentCode,
			ProvinceCode: provinceCode,
		})
	}
	return catalog
}

func (catalog *Catalog) add(path string, region Region) {
	catalog.byPath[path] = region
	catalog.byCode[strings.ToUpper(region.ProviderCode)] = region
	for _, known := range catalog.byName[region.Name] {
		if known == region {
			return
		}
	}
	catalog.byName[region.Name] = append(catalog.byName[region.Name], region)
	catalog.all = append(catalog.all, region)
}

func areaSuggestions(byName map[string][]Region, value string, limit int) []string {
	if value == "" || limit <= 0 {
		return nil
	}
	type candidate struct {
		name     string
		distance int
	}
	query := []rune(value)
	maxDistance := len(query) / 3
	if maxDistance < 1 {
		maxDistance = 1
	}
	candidates := make([]candidate, 0)
	for name, areas := range byName {
		if len(areas) != 1 {
			continue
		}
		distance := editDistance(query, []rune(name))
		if shortName := areaNameWithoutSuffix(name); shortName != name {
			distance = min(distance, editDistance(query, []rune(shortName)))
		}
		if distance <= maxDistance {
			candidates = append(candidates, candidate{name: name, distance: distance})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].distance != candidates[j].distance {
			return candidates[i].distance < candidates[j].distance
		}
		return candidates[i].name < candidates[j].name
	})
	if len(candidates) > 0 {
		bestDistance := candidates[0].distance
		end := 1
		for end < len(candidates) && end < limit && candidates[end].distance == bestDistance {
			end++
		}
		candidates = candidates[:end]
	}
	suggestions := make([]string, len(candidates))
	for i, candidate := range candidates {
		suggestions[i] = candidate.name
	}
	return suggestions
}

func areaNameWithoutSuffix(name string) string {
	for _, suffix := range []string{"特别行政区", "自治区", "自治州", "地区", "省", "市", "区", "县", "旗", "盟"} {
		if strings.HasSuffix(name, suffix) {
			return strings.TrimSuffix(name, suffix)
		}
	}
	return name
}

func editDistance(left, right []rune) int {
	previous := make([]int, len(right)+1)
	for i := range previous {
		previous[i] = i
	}
	for i, leftRune := range left {
		current := make([]int, len(right)+1)
		current[0] = i + 1
		for j, rightRune := range right {
			cost := 1
			if leftRune == rightRune {
				cost = 0
			}
			current[j+1] = min(current[j]+1, previous[j+1]+1, previous[j]+cost)
		}
		previous = current
	}
	return previous[len(right)]
}
