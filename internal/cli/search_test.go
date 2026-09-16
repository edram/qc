package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/edram/qi/internal/models"
)

type searchStub struct {
	enterprises []models.Enterprise
	people      []models.Person
}

func (s searchStub) SearchEnterprises(context.Context, string) ([]models.Enterprise, error) {
	return s.enterprises, nil
}

func (s searchStub) SearchPeople(context.Context, string) ([]models.Person, error) {
	return s.people, nil
}

func TestSearchPeopleWritesModels(t *testing.T) {
	var output bytes.Buffer
	searchCmd := SearchCmd{
		Sources: []searchSourceName{searchSourceQCC},
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
		Sources: []searchSourceName{searchSourceQCC},
		qcc: searchStub{enterprises: []models.Enterprise{{
			ID:   "3f603703d59a04cb",
			Name: "百度在线网络技术（北京）有限公司",
		}}},
		output: &output,
	}
	command := SearchEntsCmd{SearchArgs: SearchArgs{Query: "百度"}}

	if err := command.Run(&searchCmd); err != nil {
		t.Fatal(err)
	}
	const want = `[{"id":"3f603703d59a04cb","name":"百度在线网络技术（北京）有限公司"}]` + "\n"
	if got := output.String(); got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
