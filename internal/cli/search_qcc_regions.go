package cli

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	qccdata "github.com/edram/qi/internal/qcc/data"
)

type qccRegionCode string

func (code *qccRegionCode) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*code = qccRegionCode(value)
		return nil
	}

	var value json.Number
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*code = qccRegionCode(value.String())
	return nil
}

type qccRegion struct {
	Code qccRegionCode `json:"code"`
	Name string        `json:"name"`
	List []qccRegion   `json:"list"`
}

type qccRegionIndex struct {
	areasByPath map[string]qccAreaSelection
	areasByName map[string][]qccAreaSelection
	areasByCode map[string]qccAreaSelection
}

// qccAreaSelection preserves the province and terminal code required by QCC's pr/cc filter.
type qccAreaSelection struct {
	Province string
	AreaCode string
}

func (selection qccAreaSelection) code() string {
	if selection.AreaCode != "" {
		return selection.AreaCode
	}
	return selection.Province
}

// The current QCC snapshot omits these previously supported province-level regions.
var qccRegionCompatibilityCodes = map[string]string{
	"香港特别行政区": "HK",
	"澳门特别行政区": "MO",
	"台湾省":     "TW",
}

var qccRegions = mustLoadQCCRegions(qccdata.RegionsJSON)

func mustLoadQCCRegions(data []byte) qccRegionIndex {
	index, err := loadQCCRegions(data)
	if err != nil {
		panic(fmt.Sprintf("load embedded QCC regions: %v", err))
	}
	return index
}

func loadQCCRegions(data []byte) (qccRegionIndex, error) {
	var response struct {
		Status int         `json:"status"`
		Result []qccRegion `json:"result"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return qccRegionIndex{}, err
	}
	if response.Status != 200 {
		return qccRegionIndex{}, fmt.Errorf("QCC region status %d", response.Status)
	}

	index := qccRegionIndex{
		areasByPath: make(map[string]qccAreaSelection),
		areasByName: make(map[string][]qccAreaSelection),
		areasByCode: make(map[string]qccAreaSelection),
	}
	for _, province := range response.Result {
		code := string(province.Code)
		if code == "" {
			continue
		}
		selection := qccAreaSelection{Province: code}
		index.addArea(province.Name, province.Name, selection)
		index.addChildren(province.Name, province.List, selection)
	}
	for name, code := range qccRegionCompatibilityCodes {
		index.addArea(name, name, qccAreaSelection{Province: code})
	}
	return index, nil
}

func (index *qccRegionIndex) addChildren(parentPath string, regions []qccRegion, parent qccAreaSelection) {
	for _, region := range regions {
		selection := parent
		code := string(region.Code)
		if code != "" && code != parent.code() {
			selection.AreaCode = code
		}
		path := parentPath + " " + region.Name
		index.addArea(path, region.Name, selection)
		index.addChildren(path, region.List, selection)
	}
}

func (index *qccRegionIndex) addArea(path, name string, selection qccAreaSelection) {
	code := selection.code()
	if code == "" {
		return
	}
	index.areasByPath[path] = selection
	index.areasByCode[strings.ToUpper(code)] = selection
	for _, known := range index.areasByName[name] {
		if known == selection {
			return
		}
	}
	index.areasByName[name] = append(index.areasByName[name], selection)
}

func qccAreaCode(value string) (string, error) {
	value = strings.Join(strings.Fields(value), " ")
	if selection, ok := qccRegions.areasByPath[value]; ok {
		return selection.code(), nil
	}
	return "", fmt.Errorf("unsupported QCC area %q; use a full area path", value)
}

func qccEnterpriseArea(value string) (qccAreaSelection, error) {
	value = strings.Join(strings.Fields(value), " ")
	if selection, ok := qccRegions.areasByPath[value]; ok {
		return selection, nil
	}
	if selection, ok := qccRegions.areasByCode[strings.ToUpper(value)]; ok {
		return selection, nil
	}
	candidates := qccRegions.areasByName[value]
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	if len(candidates) > 1 {
		return qccAreaSelection{}, fmt.Errorf("ambiguous QCC area %q; use a full area path", value)
	}
	if suggestions := qccAreaSuggestions(value, 3); len(suggestions) > 0 {
		quoted := make([]string, len(suggestions))
		for i, suggestion := range suggestions {
			quoted[i] = strconv.Quote(suggestion)
		}
		return qccAreaSelection{}, fmt.Errorf("unsupported QCC area %q; did you mean %s?", value, strings.Join(quoted, ", "))
	}
	return qccAreaSelection{}, fmt.Errorf("unsupported QCC area %q; use an area name, full path, or code", value)
}

func qccAreaSuggestions(value string, limit int) []string {
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
	for name, selections := range qccRegions.areasByName {
		if len(selections) != 1 {
			continue
		}
		distance := qccEditDistance(query, []rune(name))
		if shortName := qccAreaNameWithoutSuffix(name); shortName != name {
			distance = min(distance, qccEditDistance(query, []rune(shortName)))
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

func qccAreaNameWithoutSuffix(name string) string {
	for _, suffix := range []string{"特别行政区", "自治区", "自治州", "地区", "省", "市", "区", "县", "旗", "盟"} {
		if strings.HasSuffix(name, suffix) {
			return strings.TrimSuffix(name, suffix)
		}
	}
	return name
}

func qccEditDistance(left, right []rune) int {
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
