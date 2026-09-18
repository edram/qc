package cli

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/edram/qi/internal/models"
	"github.com/edram/qi/internal/qcc"
)

const qccTestUserAgent = "Mozilla/5.0 test browser"

type emptyCookieSource struct{}

func (emptyCookieSource) Cookies(context.Context) ([]*http.Cookie, error) {
	return nil, nil
}

func TestSearchQCCEnterprisesRejectsProviderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"Status":401,"Message":"login required"}`))
	}))
	defer server.Close()

	searcher := &searchQCC{api: qcc.New(qcc.Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    qccTestUserAgent,
		CookieSource: emptyCookieSource{},
	})}

	_, err := searcher.SearchEnterprises(context.Background(), "百度", enterpriseSearchFilter{})
	if err == nil || !strings.Contains(err.Error(), "login required") {
		t.Fatalf("error = %v, want provider error", err)
	}
}

func TestSearchQCCEnterprisesRejectsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}))
	defer server.Close()

	searcher := &searchQCC{api: qcc.New(qcc.Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    qccTestUserAgent,
		CookieSource: emptyCookieSource{},
	})}

	_, err := searcher.SearchEnterprises(context.Background(), "百度", enterpriseSearchFilter{})
	if err == nil || !strings.Contains(err.Error(), "HTTP 502") {
		t.Fatalf("error = %v, want HTTP status", err)
	}
}

func TestSearchQCCEnterprises(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/search/searchMulti" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}

		var request struct {
			SearchKey   string `json:"searchKey"`
			PageIndex   int    `json:"pageIndex"`
			PageSize    int    `json:"pageSize"`
			SearchIndex string `json:"searchIndex"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.SearchKey != `{"onlyname":"百度"}` || request.PageIndex != 1 || request.PageSize != 20 || request.SearchIndex != "multicondition" {
			t.Fatalf("request = %#v", request)
		}

		_, _ = w.Write([]byte(`{
			"Status": 200,
			"Result": [{
				"KeyNo": "3f603703d59a04cb",
				"Name": "<em>百度</em>在线网络技术（北京）有限公司",
				"No": "110000410144104",
				"CreditCode": "91110108717743469K",
				"OperName": "何俊杰",
				"Status": "存续",
				"StartDate": 948124800000,
				"Address": "北京市海淀区上地十街10号百度大厦三层",
				"RegistCapi": "4520万元",
				"ContactNumber": "010-59928888",
				"Email": "jiangyao@baidu.com",
				"ImageUrl": "https://image.qcc.com/logo/baidu.jpg"
			}]
		}`))
	}))
	defer server.Close()

	searcher := &searchQCC{api: qcc.New(qcc.Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    qccTestUserAgent,
		CookieSource: emptyCookieSource{},
	})}

	got, err := searcher.SearchEnterprises(context.Background(), "百度", enterpriseSearchFilter{})
	if err != nil {
		t.Fatal(err)
	}
	want := []models.Enterprise{{
		ID:                  "3f603703d59a04cb",
		DetailURL:           "https://www.qcc.com/firm/3f603703d59a04cb.html",
		Name:                "百度在线网络技术（北京）有限公司",
		RegistrationNumber:  "110000410144104",
		CreditCode:          "91110108717743469K",
		LegalRepresentative: "何俊杰",
		Status:              "存续",
		EstablishedDate:     "2000-01-18",
		Address:             "北京市海淀区上地十街10号百度大厦三层",
		RegisteredCapital:   "4520万元",
		Phone:               "010-59928888",
		Email:               "jiangyao@baidu.com",
		LogoURL:             "https://image.qcc.com/logo/baidu.jpg",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SearchEnterprises() = %#v, want %#v", got, want)
	}
}

func TestSearchQCCEnterpriseFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request qccEnterpriseSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.SearchKey != `{"scope":"建筑"}` {
			t.Fatalf("searchKey = %q", request.SearchKey)
		}
		if request.Filter != `{"r":[{"pr":"ZJ","cc":[330782]}],"i":["E"],"s":["20","10","50"]}` {
			t.Fatalf("filter = %q", request.Filter)
		}
		_, _ = w.Write([]byte(`{"Status":200,"Result":[]}`))
	}))
	defer server.Close()

	searcher := &searchQCC{api: qcc.New(qcc.Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    qccTestUserAgent,
		CookieSource: emptyCookieSource{},
	})}
	_, err := searcher.SearchEnterprises(context.Background(), "建筑", enterpriseSearchFilter{
		Fields:     []enterpriseSearchField{enterpriseSearchFieldScope},
		Areas:      []string{"义乌市"},
		Industries: []string{"建筑业"},
		Statuses:   []enterpriseSearchStatus{enterpriseSearchStatusActive},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestQCCEnterpriseSearchFilterSuggestsSimilarArea(t *testing.T) {
	for _, tt := range []struct {
		input string
		exact string
	}{
		{input: "义乌", exact: `unsupported QCC area "义乌"; did you mean "义乌市"?`},
		{input: "义务市"},
	} {
		t.Run(tt.input, func(t *testing.T) {
			_, err := qccEnterpriseSearchFilter(enterpriseSearchFilter{Areas: []string{tt.input}})
			if tt.exact != "" && (err == nil || err.Error() != tt.exact) {
				t.Fatalf("qccEnterpriseSearchFilter() error = %v, want %q", err, tt.exact)
			}
			if tt.exact == "" && (err == nil || !strings.Contains(err.Error(), "did you mean") || !strings.Contains(err.Error(), `"义乌市"`)) {
				t.Fatalf("qccEnterpriseSearchFilter() error = %v, want area suggestion", err)
			}
		})
	}
}

func TestSearchQCCPeople(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/bigsearch/searchPerson" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}

		var request struct {
			Key       string `json:"key"`
			AreaInfo  string `json:"areaInfo"`
			PageIndex int    `json:"pageIndex"`
			Status    []int  `json:"status"`
			Name      string `json:"name"`
			PageSize  int    `json:"pageSize"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		wantStatus := []int{0, 1}
		if request.Key != "李彦宏" || request.AreaInfo != "" || request.PageIndex != 1 || !reflect.DeepEqual(request.Status, wantStatus) || request.Name != "李彦宏" || request.PageSize != 18 {
			t.Fatalf("request = %#v", request)
		}

		_, _ = w.Write([]byte(`{
			"Status": 200,
			"Result": [{
				"Id": "p089070bd9e914e774ef3e95d00ec5b0",
				"Name": "李彦宏",
				"MainCompany": {
					"KeyNo": "576c21e3468a6b178bbf291e4820e896",
					"Company": "北京<em>百度</em>网讯科技有限公司",
					"Role": "3"
				},
				"Intro": "百度创始人、董事长兼首席执行官（CEO）。",
				"ImageUrl": "https://image.qcc.com/person/liyanhong.jpg",
				"Count": 11,
				"CountInfo": {"PartnerCount": 20},
				"RoleDesc": "执行董事"
			}]
		}`))
	}))
	defer server.Close()

	searcher := &searchQCC{api: qcc.New(qcc.Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    qccTestUserAgent,
		CookieSource: emptyCookieSource{},
	})}

	got, err := searcher.SearchPeople(context.Background(), "李彦宏", personSearchFilter{})
	if err != nil {
		t.Fatal(err)
	}
	want := []models.Person{{
		ID:                  "p089070bd9e914e774ef3e95d00ec5b0",
		DetailURL:           "https://www.qcc.com/pl/p089070bd9e914e774ef3e95d00ec5b0.html",
		Name:                "李彦宏",
		MainCompanyID:       "576c21e3468a6b178bbf291e4820e896",
		MainCompanyName:     "北京百度网讯科技有限公司",
		Role:                "执行董事",
		RelatedCompanyCount: 11,
		PartnerCount:        20,
		Introduction:        "百度创始人、董事长兼首席执行官（CEO）。",
		AvatarURL:           "https://image.qcc.com/person/liyanhong.jpg",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SearchPeople() = %#v, want %#v", got, want)
	}
}

func TestSearchQCCPeopleFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request qccPersonSearchRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request.AreaInfo != "440300" || request.IndustryInfo != "信息传输、软件和信息技术服务业 软件和信息技术服务业" {
			t.Fatalf("request = %#v", request)
		}
		_, _ = w.Write([]byte(`{"Status":200,"Result":[]}`))
	}))
	defer server.Close()

	searcher := &searchQCC{api: qcc.New(qcc.Options{
		BaseURL:      server.URL,
		PID:          "pid",
		TID:          "tid",
		UserAgent:    qccTestUserAgent,
		CookieSource: emptyCookieSource{},
	})}
	_, err := searcher.SearchPeople(context.Background(), "李彦宏", personSearchFilter{
		Area:     "广东省 深圳市",
		Industry: "软件和信息技术服务业",
	})
	if err != nil {
		t.Fatal(err)
	}
}
