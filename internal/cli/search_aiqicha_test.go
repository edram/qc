package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/edram/qi/internal/aiqicha"
	appconfig "github.com/edram/qi/internal/config"
	"github.com/edram/qi/internal/cookies"
	"github.com/edram/qi/internal/models"
)

func TestSearchAiqichaDecodesEnterpriseResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/s/advanceFilterAjax" {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		query := r.URL.Query()
		for key, want := range map[string]string{"q": "百度", "t": "", "p": "1", "s": "10", "o": "0", "f": "{}"} {
			if !query.Has(key) || query.Get(key) != want {
				t.Errorf("query %s = %q, want %q", key, query.Get(key), want)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":0,"msg":"","ddw":1,"data":{"totalNum":15669,"resultList":[{
			"pid":"29693244029383","entName":"北京<em>百</em><em>度</em>网讯科技有限公司",
			"titleName":"北京百度网讯科技有限公司","legalPerson":"示例法人","openStatus":"开业",
			"validityFrom":"2001-07-04","regNo":"81110000902100533B","regCap":"1,352,129万(元)",
			"domicile":"测试路<em>4</em>号","telephone":"010-12354769","email":"office4@example.test",
			"newLabels":[{"key":"opening","value":{"text":"开业"}},{"key":"hightech","value":{"text":"高新技术企业(2024)"}}],
			"risk":{"selfRiskTotal":4,"unionRiskTotal":0}
		}]}}`)
	}))
	defer server.Close()
	searcher := &searchAiqicha{api: aiqicha.New(aiqicha.Options{BaseURL: server.URL, UserAgent: "test-browser", CookieSource: aiqichaTestCookies{}})}
	got, err := searcher.SearchEnterprises(context.Background(), "百度", enterpriseSearchFilter{})
	if err != nil {
		t.Fatal(err)
	}
	want := models.EnterpriseSearchResult{
		Total: 14778,
		Enterprises: []models.Enterprise{{
			ID: "28783255028393", DetailURL: "https://www.aiqicha.com/company_detail_28783255028393",
			Name: "北京百度网讯科技有限公司", LegalRepresentative: "示例法人", Status: "开业",
			EstablishedDate: "2001-06-05", CreditCode: "91110000802100433B", RegisteredCapital: "1,342,128万(元)",
			Address: "测试路5号", Phone: "010-12345678", Email: "office5@example.test",
			Tags: []string{"高新技术企业(2025)"},
			Risk: &models.EnterpriseRiskSummary{Direct: &models.EnterpriseRiskScope{Count: 5}, Associated: &models.EnterpriseRiskScope{Count: 0}},
		}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SearchEnterprises() = %#v, want %#v", got, want)
	}
}

func TestSearchAiqichaReportsRejectedResponses(t *testing.T) {
	for _, tt := range []struct {
		name       string
		httpStatus int
		body       string
		want       string
	}{
		{"login", 200, `{"status":0,"ddw":2,"data":{"limitForward":{"userType":"nologin","limitStatus":"0"},"total":0,"list":[]}}`, "login required"},
		{"limit", 200, `{"status":0,"data":{"limitForward":{"limitStatus":"1"},"totalNum":0,"resultList":[]}}`, "restricted"},
		{"denied", 200, `{"status":-1,"msg":"非法操作"}`, "非法操作"},
		{"missing status", 200, `{"data":{"totalNum":0,"resultList":[]}}`, "missing status"},
		{"missing data", 200, `{"status":0}`, "missing search results"},
		{"missing list", 200, `{"status":0,"data":{"totalNum":5}}`, "missing search results"},
		{"encoding", 200, `{"status":0,"ddw":3,"data":{}}`, "unsupported ddw"},
		{"html", 200, `<html><title>验证</title></html>`, "JSON response"},
		{"HTTP error", 403, `{"status":0,"data":{"totalNum":0,"resultList":[]}}`, "HTTP 403"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpStatus)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()
			searcher := &searchAiqicha{api: aiqicha.New(aiqicha.Options{BaseURL: server.URL, UserAgent: "test-browser", CookieSource: aiqichaTestCookies{}})}
			_, err := searcher.SearchEnterprises(context.Background(), "百度", enterpriseSearchFilter{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

type aiqichaTestCookies struct{}

func (aiqichaTestCookies) Cookies(context.Context) ([]*http.Cookie, error) {
	return []*http.Cookie{{Name: "BDUSS", Value: "test", Path: "/"}}, nil
}

func TestExecuteAiqichaSearchUsesProfileSession(t *testing.T) {
	for _, tt := range []struct {
		name      string
		userAgent string
	}{
		{"configured user agent", "configured-browser"},
		{"no user agent", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("QC_HOME", t.TempDir())
			if tt.userAgent != "" {
				if _, err := appconfig.SetUserAgent("work", tt.userAgent); err != nil {
					t.Fatal(err)
				}
			}
			path, err := cookies.Path("aiqicha", "work")
			if err != nil {
				t.Fatal(err)
			}
			if err := cookies.Write(path, []*http.Cookie{{Name: "BDUSS", Value: "work-session", Domain: ".aiqicha.com", Path: "/", Secure: true}}); err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("User-Agent"); tt.userAgent != "" && got != tt.userAgent {
					t.Errorf("User-Agent = %q", got)
				}
				if cookie, err := r.Cookie("BDUSS"); err != nil || cookie.Value != "work-session" {
					t.Errorf("wrong session: %v", err)
				}
				if r.URL.Query().Get("q") != "百度 & 技术" {
					t.Errorf("query = %q", r.URL.RawQuery)
				}
				if got, want := r.Referer(), "https://www.aiqicha.com/s?q=%E7%99%BE%E5%BA%A6+%26+%E6%8A%80%E6%9C%AF&t=0"; got != want {
					t.Errorf("Referer = %q, want %q", got, want)
				}
				fmt.Fprint(w, `{"status":0,"ddw":2,"data":{"totalNumFound":"17995","resultList":[{"pid":"25953288025363","titleName":"北京百度网讯科技有限公司"}]}}`)
			}))
			defer server.Close()
			target, _ := url.Parse(server.URL)
			transport := http.DefaultTransport
			http.DefaultTransport = aiqichaRoundTripFunc(func(request *http.Request) (*http.Response, error) {
				if request.URL.Host != "www.aiqicha.com" {
					return nil, fmt.Errorf("unexpected host %s", request.URL.Host)
				}
				forwarded := request.Clone(request.Context())
				forwarded.URL.Scheme, forwarded.URL.Host = target.Scheme, target.Host
				return transport.RoundTrip(forwarded)
			})
			t.Cleanup(func() { http.DefaultTransport = transport })
			var stdout, stderr bytes.Buffer
			code := Execute([]string{"--profile", "work", "search", "ents", "百度 & 技术", "--provider", "aiqicha"}, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Execute() = %d: %s", code, stderr.String())
			}
			var result models.EnterpriseSearchResult
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatal(err)
			}
			if result.Total != 14778 || len(result.Enterprises) != 1 || result.Enterprises[0].ID != "28783255028393" {
				t.Fatalf("unexpected output: %s", stdout.String())
			}
		})
	}
}

type aiqichaRoundTripFunc func(*http.Request) (*http.Response, error)

func (roundTrip aiqichaRoundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return roundTrip(request)
}

func TestSearchAiqichaRejectsUnsupportedFilters(t *testing.T) {
	for _, tt := range []struct {
		name   string
		query  string
		filter enterpriseSearchFilter
		want   string
	}{
		{"empty", "  ", enterpriseSearchFilter{}, "query is required"},
		{"match", "百度", enterpriseSearchFilter{Fields: []enterpriseSearchField{enterpriseSearchFieldName}}, "--match"},
		{"area", "百度", enterpriseSearchFilter{Areas: []string{"北京市"}}, "--area"},
		{"industry", "百度", enterpriseSearchFilter{Industries: []string{"软件"}}, "--industry"},
		{"status", "百度", enterpriseSearchFilter{Statuses: []enterpriseSearchStatus{enterpriseSearchStatusActive}}, "--status"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newSearchAiqicha("", "").SearchEnterprises(context.Background(), tt.query, tt.filter)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %s", err, tt.want)
			}
		})
	}
}

func TestSearchAiqichaHandlesHongKongAndEmptyResults(t *testing.T) {
	for _, tt := range []struct {
		name string
		body string
		want []models.Enterprise
	}{
		{"hong kong", `{"status":0,"ddw":2,"data":{"totalNum":1,"resultList":[{"pid":"F44900270694","hkid":"F44900270694","isHkComp":1,"entName":"<em>百</em>度 &amp; 示例","regNo":"123123123123123","newLabels":[],"risk":{}}]}}`, []models.Enterprise{{ID: "F66700240976", DetailURL: "https://www.aiqicha.com/hkcompany_detail_F66700240976", Name: "百度 & 示例", RegistrationNumber: "123123123123123"}}},
		{"empty", `{"status":0,"data":{"totalNum":0,"resultList":[]}}`, []models.Enterprise{}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, tt.body) }))
			defer server.Close()
			searcher := &searchAiqicha{api: aiqicha.New(aiqicha.Options{BaseURL: server.URL, UserAgent: "test-browser", CookieSource: aiqichaTestCookies{}})}
			result, err := searcher.SearchEnterprises(context.Background(), "百度", enterpriseSearchFilter{})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(result.Enterprises, tt.want) {
				t.Fatalf("enterprises = %#v, want %#v", result.Enterprises, tt.want)
			}
		})
	}
}
