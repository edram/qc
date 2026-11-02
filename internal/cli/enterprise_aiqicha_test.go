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

func TestExecuteAiqichaEnterpriseViewUsesProfileSession(t *testing.T) {
	for _, userAgent := range []string{"", "configured-browser"} {
		t.Run(userAgent, func(t *testing.T) {
			t.Setenv("QC_HOME", t.TempDir())
			if userAgent != "" {
				if _, err := appconfig.SetUserAgent("work", userAgent); err != nil {
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
				if r.URL.Path != "/detail/basicAllDataAjax" || r.URL.Query().Get("pid") != "125" {
					t.Errorf("request = %s", r.URL)
				}
				if userAgent != "" && r.UserAgent() != userAgent {
					t.Errorf("User-Agent = %q", r.UserAgent())
				}
				if cookie, err := r.Cookie("BDUSS"); err != nil || cookie.Value != "work-session" {
					t.Errorf("wrong profile session: %v", err)
				}
				fmt.Fprint(w, `{"status":0,"data":{"basicData":{"entName":"示例企业"}}}`)
			}))
			defer server.Close()
			target, _ := url.Parse(server.URL)
			transport := http.DefaultTransport
			http.DefaultTransport = aiqichaRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Host != "www.aiqicha.com" {
					return nil, fmt.Errorf("unexpected host %s", r.URL.Host)
				}
				forwarded := r.Clone(r.Context())
				forwarded.URL.Scheme, forwarded.URL.Host = target.Scheme, target.Host
				return transport.RoundTrip(forwarded)
			})
			t.Cleanup(func() { http.DefaultTransport = transport })
			var stdout, stderr bytes.Buffer
			if code := Execute([]string{"--profile", "work", "ent", "view", "125", "--provider", "aiqicha"}, &stdout, &stderr); code != 0 {
				t.Fatalf("Execute() = %d: %s", code, stderr.String())
			}
			var got models.EnterpriseDetail
			if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.ID != "125" || got.Name != "示例企业" {
				t.Fatalf("detail = %#v", got)
			}
		})
	}
}

func TestAiqichaEnterpriseDetailDecodesBasicData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/detail/basicAllDataAjax" || r.URL.Query().Get("pid") != "125" {
			t.Errorf("request = %s %s", r.Method, r.URL)
		}
		pageURL := "https://www.aiqicha.com/company_detail_125"
		if r.Referer() != pageURL || r.Header.Get("Zx-Open-Url") != pageURL {
			t.Errorf("page context = %q, %q", r.Referer(), r.Header.Get("Zx-Open-Url"))
		}
		fmt.Fprint(w, `{"status":0,"ddw":1,"data":{"basicData":{
			"entName":"示例 &amp; 企业","unifiedCode":"81110000902100533B","regCode":"1004",
			"legalPerson":"张三","openStatus":"开业","startDate":"2001-07-04",
			"regAddr":"测试路4号","regCapital":"124万(元)","paidinCapital":"90万(元)",
			"scope":"软件开发","entType":"有限责任公司","authority":"示例市场监督管理局",
			"orgNo":"ORG4","taxNo":"TAX4","annualDate":"**","telephone":"-","email":"-",
			"openTime":"2001-07-04 至 无固定期限","industry":"软件开发",
			"insuranceInfo":{"reportUrl":"/annualreport?pid=124&year=2024"}
		}}}`)
	}))
	defer server.Close()
	searcher := &searchAiqicha{api: aiqicha.New(aiqicha.Options{BaseURL: server.URL, CookieSource: aiqichaTestCookies{}})}
	got, err := searcher.GetEnterpriseDetail(context.Background(), "125")
	if err != nil {
		t.Fatal(err)
	}
	want := models.EnterpriseDetail{
		Enterprise: models.Enterprise{
			ID: "125", DetailURL: "https://www.aiqicha.com/company_detail_125", Name: "示例 & 企业",
			CreditCode: "91110000802100433B", RegistrationNumber: "1005", LegalRepresentative: "张三",
			Status: "开业", EstablishedDate: "2001-06-05", Address: "测试路5号", RegisteredCapital: "125万(元)",
		},
		BusinessScope: "软件开发", CompanyType: "有限责任公司", RegistrationAuthority: "示例市场监督管理局",
		OrganizationCode: "ORG5", TaxNumber: "TAX5", CheckDate: "**", ActualCapital: "80万(元)",
		TermStart: "2001-06-05", TermEnd: "无固定期限", Industry: &models.EnterpriseIndustry{Name: "软件开发"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("detail = %#v, want %#v", got, want)
	}
}

func TestAiqichaEnterpriseDetailRejectsInvalidResponses(t *testing.T) {
	for _, tt := range []struct{ name, body, want string }{
		{"missing status", `{"data":{"basicData":{"entName":"示例"}}}`, "missing status"},
		{"denied", `{"status":-1,"msg":"非法操作"}`, "非法操作"},
		{"login", `{"status":0,"data":{"isDidiwei":true}}`, "login required"},
		{"missing data", `{"status":0}`, "missing company details"},
		{"empty name", `{"status":0,"data":{"basicData":{"entName":"-"}}}`, "missing company details"},
		{"wrong company", `{"status":0,"data":{"basicData":{"entName":"其他企业","insuranceInfo":{"reportUrl":"/annualreport?pid=999&year=2025"}}}}`, "different enterprise"},
		{"html", `<html>请登录</html>`, "JSON response"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, tt.body) }))
			defer server.Close()
			searcher := &searchAiqicha{api: aiqicha.New(aiqicha.Options{BaseURL: server.URL, CookieSource: aiqichaTestCookies{}})}
			_, err := searcher.GetEnterpriseDetail(context.Background(), "125")
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestAiqichaEnterpriseDetailRejectsInvalidIDsBeforeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("request sent for invalid enterprise ID")
		fmt.Fprint(w, `{"status":0,"data":{"basicData":{"entName":"示例"}}}`)
	}))
	defer server.Close()
	searcher := &searchAiqicha{api: aiqicha.New(aiqicha.Options{BaseURL: server.URL, CookieSource: aiqichaTestCookies{}})}
	for _, id := range []string{"", " ", "qc/125", "../125", "125?pid=999", "F125", "-125"} {
		t.Run(id, func(t *testing.T) {
			_, err := searcher.GetEnterpriseDetail(context.Background(), id)
			if err == nil || !strings.Contains(err.Error(), "enterprise ID") {
				t.Fatalf("error = %v", err)
			}
		})
	}
}

func TestAiqichaEnterpriseDetailIncludesAvailableLists(t *testing.T) {
	for _, restricted := range []bool{false, true} {
		t.Run(fmt.Sprint(restricted), func(t *testing.T) {
			limit := `null`
			if restricted {
				limit = `{"limitStatus":"2","userType":"loginUser"}`
			}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, `{"status":0,"data":{"basicData":{"entName":"示例企业"},
				"shareholdersData":{"limitForward":%s,"list":[
					{"stockId":"person-1","name":"股东一","subRate":"75%%","subMoney":"75万(元)","subDate":"2020-01-01","labels":[{"value":{"text":"大股东"}}]},
					{"pid":"company-2","name":"股东企业","subRate":"25%%"}]},
				"directorsData":{"limitForward":%s,"list":[{"personId":"person-1","name":"股东一","title":"执行董事,经理"}]},
				"branchsData":{"limitForward":%s,"list":[{"pid":"12501","entName":"示例分公司","legalPerson":"李四"}]}
				}}`, limit, limit, limit)
			}))
			defer server.Close()
			searcher := &searchAiqicha{api: aiqicha.New(aiqicha.Options{BaseURL: server.URL, CookieSource: aiqichaTestCookies{}})}
			got, err := searcher.GetEnterpriseDetail(context.Background(), "125")
			if err != nil {
				t.Fatal(err)
			}
			var shareholders []models.EnterpriseShareholder
			var employees []models.EnterpriseEmployee
			var branches []models.EnterpriseBranch
			if !restricted {
				shareholders = []models.EnterpriseShareholder{
					{ID: "person-1", Name: "股东一", Percent: "75%", SubscribedCapital: "75万(元)", SubscribedCapitalDate: "2020-01-01", Tags: []string{"大股东"}},
					{ID: "company-2", Name: "股东企业", Percent: "25%"},
				}
				employees = []models.EnterpriseEmployee{{ID: "person-1", Name: "股东一", Position: "执行董事,经理"}}
				branches = []models.EnterpriseBranch{{ID: "12501", Name: "示例分公司", LegalRepresentative: "李四"}}
			}
			if !reflect.DeepEqual(got.Shareholders, shareholders) || !reflect.DeepEqual(got.Employees, employees) || !reflect.DeepEqual(got.Branches, branches) {
				t.Fatalf("related lists = %#v, %#v, %#v", got.Shareholders, got.Employees, got.Branches)
			}
		})
	}
}
