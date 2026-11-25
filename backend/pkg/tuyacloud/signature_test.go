package tuyacloud

import "testing"

func TestSignatureMatchesIndependentlyComputedVectors(t *testing.T) {
	tests := []struct {
		name    string
		request signedRequest
		want    string
	}{
		{
			name: "token call without access token",
			request: signedRequest{
				method: "GET", pathWithQuery: "/v1.0/token?grant_type=1", accessID: "testid",
				accessSecret: "testsecret", timestampMilli: "1700000000000",
			},
			want: "18A53EA2AB6725F8B19D5A658B2453789B9DF916B9C4E6BA4D64B6C5A5564653",
		},
		{
			name: "business call with access token",
			request: signedRequest{
				method: "GET", pathWithQuery: "/v1.0/users/sweepers/file/dev1/realtime-map", accessID: "testid",
				accessSecret: "testsecret", accessToken: "tok", timestampMilli: "1700000000123",
			},
			want: "888A61A2E842A1BCBEA36667F054E838AE51C7C8A0BAB7F377763059FD2C1060",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.request.signature(); got != test.want {
				t.Fatalf("got %s, want %s", got, test.want)
			}
		})
	}
}

func TestBaseURLForRegion(t *testing.T) {
	wanted := map[string]string{
		"us": "https://openapi.tuyaus.com",
		"eu": "https://openapi.tuyaeu.com",
		"cn": "https://openapi.tuyacn.com",
		"in": "https://openapi.tuyain.com",
	}
	for region, wantURL := range wanted {
		got, err := baseURLForRegion(region)
		if err != nil || got != wantURL {
			t.Errorf("%s: got %q, %v", region, got, err)
		}
	}
	if _, err := baseURLForRegion("xx"); err == nil {
		t.Fatal("unknown region must fail")
	}
}
