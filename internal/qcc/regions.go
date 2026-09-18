package qcc

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Area identifies a QCC region and the province filter that contains it.
type Area struct {
	// ProvinceCode is QCC's pr filter value.
	ProvinceCode string
	// Code is the selected terminal code and equals ProvinceCode for province-level areas.
	Code string
}

type regionCode string

func (code *regionCode) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*code = regionCode(value)
		return nil
	}

	var value json.Number
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*code = regionCode(value.String())
	return nil
}

type region struct {
	Code regionCode `json:"code"`
	Name string     `json:"name"`
	List []region   `json:"list"`
}

type regionIndex struct {
	areasByPath map[string]Area
	areasByName map[string][]Area
	areasByCode map[string]Area
}

// The current QCC snapshot omits these previously supported province-level regions.
var regionCompatibilityCodes = map[string]string{
	"香港特别行政区": "HK",
	"澳门特别行政区": "MO",
	"台湾省":     "TW",
}

// regionData is a snapshot of QCC's region response.
//
//go:embed data/qcc_regions.json
var regionData []byte

// regionCatalog indexes the embedded snapshot by full path, unique name, and code.
var regionCatalog = mustLoadRegions(regionData)

func mustLoadRegions(data []byte) regionIndex {
	index, err := loadRegions(data)
	if err != nil {
		panic(fmt.Sprintf("load embedded QCC regions: %v", err))
	}
	return index
}

func loadRegions(data []byte) (regionIndex, error) {
	var response struct {
		Status int      `json:"status"`
		Result []region `json:"result"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return regionIndex{}, err
	}
	if response.Status != 200 {
		return regionIndex{}, fmt.Errorf("QCC region status %d", response.Status)
	}

	index := regionIndex{
		areasByPath: make(map[string]Area),
		areasByName: make(map[string][]Area),
		areasByCode: make(map[string]Area),
	}
	for _, province := range response.Result {
		code := string(province.Code)
		if code == "" {
			continue
		}
		area := Area{ProvinceCode: code, Code: code}
		index.addArea(province.Name, province.Name, area)
		index.addChildren(province.Name, province.List, area)
	}
	for name, code := range regionCompatibilityCodes {
		index.addArea(name, name, Area{ProvinceCode: code, Code: code})
	}
	return index, nil
}

func (index *regionIndex) addChildren(parentPath string, regions []region, parent Area) {
	for _, child := range regions {
		area := parent
		if code := string(child.Code); code != "" && code != parent.Code {
			area.Code = code
		}
		path := parentPath + " " + child.Name
		index.addArea(path, child.Name, area)
		index.addChildren(path, child.List, area)
	}
}

func (index *regionIndex) addArea(path, name string, area Area) {
	if area.Code == "" {
		return
	}
	index.areasByPath[path] = area
	index.areasByCode[strings.ToUpper(area.Code)] = area
	for _, known := range index.areasByName[name] {
		if known == area {
			return
		}
	}
	index.areasByName[name] = append(index.areasByName[name], area)
}

// ResolveArea resolves a full path, unique name, or code from QCC's region catalog.
func ResolveArea(value string) (Area, error) {
	value = strings.Join(strings.Fields(value), " ")
	if area, ok := regionCatalog.areasByPath[value]; ok {
		return area, nil
	}
	if area, ok := regionCatalog.areasByCode[strings.ToUpper(value)]; ok {
		return area, nil
	}
	candidates := regionCatalog.areasByName[value]
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) > 1 {
		return Area{}, fmt.Errorf("ambiguous QCC area %q; use a full area path", value)
	}
	if suggestions := areaSuggestions(value, 3); len(suggestions) > 0 {
		quoted := make([]string, len(suggestions))
		for i, suggestion := range suggestions {
			quoted[i] = strconv.Quote(suggestion)
		}
		return Area{}, fmt.Errorf("unsupported QCC area %q; did you mean %s?", value, strings.Join(quoted, ", "))
	}
	return Area{}, fmt.Errorf("unsupported QCC area %q; use an area name, full path, or code", value)
}

func areaSuggestions(value string, limit int) []string {
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
	for name, areas := range regionCatalog.areasByName {
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
			current[j+1] = min(
				current[j]+1,
				previous[j+1]+1,
				previous[j]+cost,
			)
		}
		previous = current
	}
	return previous[len(right)]
}
