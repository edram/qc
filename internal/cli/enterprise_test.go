package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
	"github.com/edram/qi/internal/models"
	"github.com/edram/qi/internal/qcc"
)

func TestExtractQCCInitialState(t *testing.T) {
	page := []byte(`<html><script>
const label = "ignored { brace }";
window.__INITIAL_STATE__ = {"company":{"companyDetail":{"Name":"示例企业","Scope":"经营范围含 } 字符","Value":"escaped \\\" quote"}},"other":{"ok":true}};
</script></html>`)
	want := `{"company":{"companyDetail":{"Name":"示例企业","Scope":"经营范围含 } 字符","Value":"escaped \\\" quote"}},"other":{"ok":true}}`

	got, err := qccEnterpriseInitialState(page)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("initial state = %s, want %s", got, want)
	}
}

func TestExtractQCCInitialStateErrors(t *testing.T) {
	for _, page := range [][]byte{
		[]byte(`<html>no state</html>`),
		[]byte(`<script>window.__INITIAL_STATE__ = {"company":</script>`),
		[]byte(`<script>window.__INITIAL_STATE__ = {"other":{}};</script>`),
	} {
		if _, err := qccEnterpriseInitialState(page); err == nil {
			t.Fatalf("qccEnterpriseInitialState(%q) error = nil", page)
		}
	}
}

type enterprisePageCookieSource struct{}

func (enterprisePageCookieSource) Cookies(context.Context) ([]*http.Cookie, error) { return nil, nil }

func TestSearchQCCEnterpriseDetailReadsFirmPage(t *testing.T) {
	client := &http.Client{Transport: enterpriseRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodGet || r.URL.Path != "/firm/company-1.html" {
			return nil, fmt.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body: io.NopCloser(strings.NewReader(`<script>window.__INITIAL_STATE__ = {"company":{"companyDetail":{
  "KeyNo":"company-1","Name":"示例企业","CreditCode":"9132","No":"1001",
  "Oper":{"Name":"法定代表人"},"Status":"存续","StartDate":"2020-01-02",
  "Address":"南京市","RegistCapi":"100万元",
  "Scope":"软件开发","EconKind":"有限责任公司","BelongOrg":"南京市市场监管局",
  "OrgNo":"ORG1","TaxNo":"TAX1","TermStart":"2020-01-02","TermEnd":"2040-01-01",
  "CheckDate":"2024-01-01","RecCap":"80万元","EnglishName":"Example Co.",
  "Area":{"Province":"江苏省","City":"南京市","County":"鼓楼区"},
  "ContactInfo":{"PhoneNumber":"025-123456","Email":"hello@example.com","WebSite":[{"Name":"官网","Url":"https://example.com"}]},
  "QccIndustry":{"Cn":"I","Dn":"信息传输、软件和信息技术服务业"},
  "IndustryV3":{"SmallCategory":"软件开发"},
  "Partners":[{"KeyNo":"p1","StockName":"股东","StockPercent":"50%"}],
  "Employees":[{"KeyNo":"e1","Name":"经理","Job":"经理"}],
  "Branches":[{"CompanyId":"b1","Name":"分公司"}],
  "ChangeRecords":[{"ProjectName":"名称","BeforeContent":"旧名","AfterContent":"新名","ChangeDate":"2024-01-01"}],
  "OriginalName":[{"Name":"旧公司","ChangeDate":"2023-01-01"}]
}}};</script>`)),
			Request: r,
		}, nil
	})}

	searcher := &searchQCC{api: qcc.New(qcc.Options{
		BaseURL:      "https://www.qcc.com",
		HTTPClient:   client,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    "test",
		CookieSource: enterprisePageCookieSource{},
	})}
	got, err := searcher.GetEnterpriseDetail(context.Background(), "company-1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Enterprise.ID != "company-1" || got.Enterprise.Name != "示例企业" || got.Enterprise.CreditCode != "9132" ||
		got.RegistrationNumber != "1001" || got.LegalRepresentative != "法定代表人" ||
		got.Status != "存续" || got.EstablishedDate != "2020-01-02" || got.Phone != "025-123456" ||
		got.Email != "hello@example.com" || got.AdministrativeDivisions == nil ||
		got.AdministrativeDivisions.Province != "江苏省" || got.BusinessScope != "软件开发" ||
		got.CompanyType != "有限责任公司" || got.RegistrationAuthority != "南京市市场监管局" ||
		got.OrganizationCode != "ORG1" || got.TaxNumber != "TAX1" || got.TermEnd != "2040-01-01" ||
		got.ActualCapital != "80万元" || got.EnglishName != "Example Co." || got.Industry == nil ||
		got.Industry.Name != "信息传输、软件和信息技术服务业" || len(got.Websites) != 1 ||
		len(got.Shareholders) != 1 || len(got.Employees) != 1 || len(got.Branches) != 1 ||
		len(got.Changes) != 1 || len(got.PreviousNames) != 1 {
		t.Fatalf("enterprise = %#v", got)
	}
}

type enterpriseRoundTripFunc func(*http.Request) (*http.Response, error)

func (f enterpriseRoundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestQCCEnterpriseReference(t *testing.T) {
	tests := []struct {
		input string
		want  string
		err   bool
	}{
		{input: "company-1", want: "company-1"},
		{input: "qc/company-1", want: "company-1"},
		{input: "qc/", err: true},
		{input: "other/company-1", err: true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := qccEnterpriseReference(tt.input)
			if (err != nil) != tt.err || got != tt.want {
				t.Fatalf("qccEnterpriseReference() = %q, %v; want %q, error=%v", got, err, tt.want, tt.err)
			}
		})
	}
}

type enterpriseViewStub struct {
	detail models.EnterpriseDetail
}

func (s enterpriseViewStub) GetEnterpriseDetail(context.Context, string) (models.EnterpriseDetail, error) {
	return s.detail, nil
}

func TestEnterpriseViewWritesEnterprise(t *testing.T) {
	var output bytes.Buffer
	command := EnterpriseCmd{
		View:   EnterpriseViewCmd{Reference: "qc/company-1"},
		qcc:    enterpriseViewStub{detail: models.EnterpriseDetail{Enterprise: models.Enterprise{ID: "company-1", Name: "示例企业"}}},
		output: &output,
	}
	if err := command.View.Run(&command); err != nil {
		t.Fatal(err)
	}
	var got models.EnterpriseDetail
	if err := json.Unmarshal(output.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := models.EnterpriseDetail{Enterprise: models.Enterprise{ID: "company-1", Name: "示例企业"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("output = %#v, want %#v", got, want)
	}
}

func TestEnterpriseCommandPath(t *testing.T) {
	parser, err := kong.New(New(), kong.Name("qc"), kong.Exit(func(int) {}))
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := parser.Parse([]string{"ent", "view", "qc/company-1"})
	if err != nil {
		t.Fatal(err)
	}
	if got := ctx.Command(); got != "ent view <qc/id>" {
		t.Fatalf("Command() = %q, want %q", got, "ent view <qc/id>")
	}
}
