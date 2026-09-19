package cli

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/edram/qi/internal/models"
)

type searchStub struct {
	enterpriseResult models.EnterpriseSearchResult
	people           []models.Person
}

type searchFilterStub struct {
	enterprise enterpriseSearchFilter
	person     personSearchFilter
}

func (s *searchFilterStub) SearchEnterprises(_ context.Context, _ string, filter enterpriseSearchFilter) (models.EnterpriseSearchResult, error) {
	s.enterprise = filter
	return models.EnterpriseSearchResult{}, nil
}

func (s *searchFilterStub) SearchPeople(_ context.Context, _ string, filter personSearchFilter) ([]models.Person, error) {
	s.person = filter
	return nil, nil
}

func (s searchStub) SearchEnterprises(context.Context, string, enterpriseSearchFilter) (models.EnterpriseSearchResult, error) {
	return s.enterpriseResult, nil
}

func (s searchStub) SearchPeople(context.Context, string, personSearchFilter) ([]models.Person, error) {
	return s.people, nil
}

func TestSearchPeopleWritesModels(t *testing.T) {
	var output bytes.Buffer
	searchCmd := SearchCmd{
		Providers: []searchProviderName{searchProviderQCC},
		qcc: searchStub{people: []models.Person{{
			ID:   "p089070bd9e914e774ef3e95d00ec5b0",
			Name: "李彦宏",
		}}},
		output: &output,
	}
	command := SearchPersCmd{SearchArgs: SearchArgs{Query: "李彦宏"}}

	if err := command.Run(&searchCmd); err != nil {
		t.Fatal(err)
	}
	const want = `[{"id":"p089070bd9e914e774ef3e95d00ec5b0","name":"李彦宏"}]` + "\n"
	if got := output.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSearchEnterprisesWritesModels(t *testing.T) {
	var output bytes.Buffer
	searchCmd := SearchCmd{
		Providers: []searchProviderName{searchProviderQCC},
		qcc: searchStub{enterpriseResult: models.EnterpriseSearchResult{
			Total: 64,
			Enterprises: []models.Enterprise{{
				ID:   "3f603703d59a04cb",
				Name: "百度在线网络技术（北京）有限公司",
				Tags: []string{"高新技术企业"},
			}},
			Aggregations: models.EnterpriseSearchAggregations{
				Provinces:  []models.EnterpriseSearchAggregation{{Code: "BJ", Name: "北京市", Count: 7}},
				Industries: []models.EnterpriseSearchAggregation{{Code: "M", Name: "科学研究和技术服务业", Count: 23}},
			},
		}},
		output: &output,
	}
	command := SearchEntsCmd{SearchArgs: SearchArgs{Query: "百度"}}

	if err := command.Run(&searchCmd); err != nil {
		t.Fatal(err)
	}
	const want = `{"total":64,"enterprises":[{"id":"3f603703d59a04cb","name":"百度在线网络技术（北京）有限公司","tags":["高新技术企业"]}],"aggregations":{"provinces":[{"code":"BJ","name":"北京市","count":7}],"industries":[{"code":"M","name":"科学研究和技术服务业","count":23}]}}` + "\n"
	if got := output.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSearchEnterprisesCombinesProviders(t *testing.T) {
	var output bytes.Buffer
	searchCmd := SearchCmd{
		Providers: []searchProviderName{searchProviderQCC, searchProviderAiqicha},
		qcc: searchStub{enterpriseResult: models.EnterpriseSearchResult{
			Total:       2,
			Enterprises: []models.Enterprise{{ID: "qcc-1", Name: "甲公司"}},
			Aggregations: models.EnterpriseSearchAggregations{
				Provinces:  []models.EnterpriseSearchAggregation{{Code: "BJ", Name: "北京市", Count: 2}},
				Industries: []models.EnterpriseSearchAggregation{{Code: "M", Name: "科学研究和技术服务业", Count: 2}},
			},
		}},
		aiqicha: searchStub{enterpriseResult: models.EnterpriseSearchResult{
			Total:       2,
			Enterprises: []models.Enterprise{{ID: "aiqicha-1", Name: "乙公司"}},
			Aggregations: models.EnterpriseSearchAggregations{
				Provinces: []models.EnterpriseSearchAggregation{
					{Code: "BJ", Name: "北京市", Count: 1},
					{Code: "GD", Name: "广东省", Count: 1},
				},
				Industries: []models.EnterpriseSearchAggregation{{Code: "M", Name: "科学研究和技术服务业", Count: 2}},
			},
		}},
		output: &output,
	}
	command := SearchEntsCmd{SearchArgs: SearchArgs{Query: "科技"}}

	if err := command.Run(&searchCmd); err != nil {
		t.Fatal(err)
	}
	const want = `{"total":4,"enterprises":[{"id":"qcc-1","name":"甲公司"},{"id":"aiqicha-1","name":"乙公司"}],"aggregations":{"provinces":[{"code":"BJ","name":"北京市","count":3},{"code":"GD","name":"广东省","count":1}],"industries":[{"code":"M","name":"科学研究和技术服务业","count":4}]}}` + "\n"
	if got := output.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSearchEnterpriseFiltersReachProvider(t *testing.T) {
	command := New()
	parser, err := kong.New(command, kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := parser.Parse([]string{
		"search", "ents", "建筑", "--provider", "qcc",
		"--match", "scope", "--area", "北京市", "--industry", "建筑业", "--status", "active",
	})
	if err != nil {
		t.Fatal(err)
	}
	searcher := &searchFilterStub{}
	command.Search.qcc = searcher
	command.Search.output = &bytes.Buffer{}
	if err := ctx.Run(); err != nil {
		t.Fatal(err)
	}
	want := enterpriseSearchFilter{
		Fields:     []enterpriseSearchField{enterpriseSearchFieldScope},
		Areas:      []string{"北京市"},
		Industries: []string{"建筑业"},
		Statuses:   []enterpriseSearchStatus{enterpriseSearchStatusActive},
	}
	if !reflect.DeepEqual(searcher.enterprise, want) {
		t.Fatalf("enterprise filter = %#v, want %#v", searcher.enterprise, want)
	}
}

func TestSearchPeopleFiltersReachProvider(t *testing.T) {
	command := New()
	parser, err := kong.New(command, kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := parser.Parse([]string{
		"search", "pers", "李彦宏", "--provider", "qcc",
		"--area", "广东省 深圳市", "--industry", "软件和信息技术服务业",
	})
	if err != nil {
		t.Fatal(err)
	}
	searcher := &searchFilterStub{}
	command.Search.qcc = searcher
	command.Search.output = &bytes.Buffer{}
	if err := ctx.Run(); err != nil {
		t.Fatal(err)
	}
	want := personSearchFilter{
		Area:     "广东省 深圳市",
		Industry: "软件和信息技术服务业",
	}
	if searcher.person != want {
		t.Fatalf("person filter = %#v, want %#v", searcher.person, want)
	}
}
