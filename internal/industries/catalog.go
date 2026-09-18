package industries

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/edram/qi/internal/data"
)

type Provider string

const (
	ProviderQCC     Provider = "qcc"
	ProviderAiqicha Provider = "aiqicha"
)

var ErrProviderRequired = errors.New("industry provider is required")

type Industry struct {
	ID           string
	Name         string
	Path         string
	Depth        int
	ParentID     string
	ProviderCode string
}

type Catalog struct {
	provider Provider
	byPath   map[string]Industry
	byName   map[string][]Industry
	byCode   map[string]Industry
	all      []Industry
}

type rawIndustry struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Depth    int               `json:"depth"`
	Codes    map[string]string `json:"codes"`
	Children []rawIndustry     `json:"children"`
}

type rawProviderIndustry struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Depth int    `json:"depth"`
	Code  string `json:"code"`
}

type rawCatalog struct {
	Industries   []rawIndustry                    `json:"industries"`
	ProviderOnly map[string][]rawProviderIndustry `json:"providerOnly"`
}

var (
	rawOnce sync.Once
	rawData rawCatalog
	rawErr  error
	caches  sync.Map
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
	value, _ := caches.LoadOrStore(provider, &catalogState{})
	state := value.(*catalogState)
	state.once.Do(func() {
		rawOnce.Do(func() {
			rawErr = json.Unmarshal(data.Industries, &rawData)
		})
		if rawErr != nil {
			state.err = fmt.Errorf("load industry catalog: %w", rawErr)
			return
		}
		state.catalog = buildCatalog(provider, rawData)
	})
	return state.catalog, state.err
}

func (catalog *Catalog) Resolve(value string) (Industry, error) {
	value = normalize(value)
	if industry, ok := catalog.byPath[value]; ok {
		return industry, nil
	}
	if industry, ok := catalog.byCode[strings.ToUpper(value)]; ok {
		return industry, nil
	}
	if candidates := catalog.byName[value]; len(candidates) == 1 {
		return candidates[0], nil
	} else if len(candidates) > 1 {
		return Industry{}, fmt.Errorf("ambiguous %s industry %q; use a full industry path", catalog.provider, value)
	}
	return Industry{}, fmt.Errorf("unsupported %s industry %q", catalog.provider, value)
}

// Search returns industries whose name contains query.
func (catalog *Catalog) Search(query string, limit int) []Industry {
	query = strings.ToLower(normalize(query))
	if query == "" || limit == 0 {
		return nil
	}
	result := make([]Industry, 0)
	for _, industry := range catalog.all {
		if strings.Contains(strings.ToLower(industry.Name), query) {
			result = append(result, industry)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Depth != result[j].Depth {
			return result[i].Depth < result[j].Depth
		}
		return result[i].Path < result[j].Path
	})
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

// TopLevel returns the provider-supported first-level industries.
func (catalog *Catalog) TopLevel(limit int) []Industry {
	result := make([]Industry, 0)
	for _, industry := range catalog.all {
		if industry.Depth == 1 {
			result = append(result, industry)
		}
	}
	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}
	return result
}

func buildCatalog(provider Provider, raw rawCatalog) *Catalog {
	catalog := &Catalog{
		provider: provider,
		byPath:   make(map[string]Industry),
		byName:   make(map[string][]Industry),
		byCode:   make(map[string]Industry),
	}
	var add func(rawIndustry, string, Industry)
	add = func(node rawIndustry, parentPath string, parent Industry) {
		code := node.Codes[string(provider)]
		path := parentPath
		current := parent
		if code != "" {
			if path == "" {
				path = node.Name
			} else {
				path += " " + node.Name
			}
			current = Industry{
				ID:           node.ID,
				Name:         node.Name,
				Path:         path,
				Depth:        node.Depth,
				ParentID:     parent.ID,
				ProviderCode: code,
			}
			catalog.add(current)
		}
		for _, child := range node.Children {
			add(child, path, current)
		}
	}
	for _, node := range raw.Industries {
		add(node, "", Industry{})
	}
	for _, node := range raw.ProviderOnly[string(provider)] {
		catalog.add(Industry{ID: node.ID, Name: node.Name, Path: node.Name, Depth: node.Depth, ProviderCode: node.Code})
	}
	return catalog
}

func (catalog *Catalog) add(industry Industry) {
	catalog.byPath[industry.Path] = industry
	catalog.byCode[strings.ToUpper(industry.ProviderCode)] = industry
	catalog.byName[industry.Name] = append(catalog.byName[industry.Name], industry)
	catalog.all = append(catalog.all, industry)
}

func normalize(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
