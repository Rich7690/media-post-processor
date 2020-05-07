package utils

import (
	"testing"
)

func TestWebClientImpl_MakeGetRequest(t *testing.T) {
	/*t.SkipNow()
	r := mux.NewRouter()
	r.HandleFunc("/foo", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(r)
	defer srv.Close()

	impl := WebClientImpl{client:*srv.Client()}

	base, err := url.Parse(srv.URL)
	assert.NoError(t, err)
	resp, _, err := impl.MakeGetRequest(*base, "/foo", url.Values{})

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)*/
}
