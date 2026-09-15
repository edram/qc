package qcc

import "testing"

func TestGenerateSign(t *testing.T) {
	const path = "/api/batchSearch/getBatchSearchSuccessList"
	const payload = `{"cacheData":[],"pageIndex":1,"pageSize":20,"type":"company"}`

	sign := GenerateSign(path, payload, "tid-123")

	if sign.HeaderName != "9e9d01ada82ea2f1850b" {
		t.Fatalf("HeaderName = %q", sign.HeaderName)
	}
	if sign.HeaderValue != "25a5974275e3cbf1c3bcf172a9f374fc742ebc9d4863bfc60e55a7d6f5bc8e2621f9955500b80e0c5d64246ce98ae42d1816b73160a87196e03ceda613268172" {
		t.Fatalf("HeaderValue = %q", sign.HeaderValue)
	}
}
