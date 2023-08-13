package pantry_path_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SergioFloresG/pantry_path"
)

type CaseConfig struct {
	Header     string `json:"header"`
	Id         string `json:"id"`
	Basket     string `json:"basket"`
	PrefixPath string `json:"prefixPath"`
}

func (cfg CaseConfig) Path() string {
	prefix := ""
	if cfg.PrefixPath == "" {
		prefix = strings.TrimPrefix(cfg.PrefixPath, "/")
		prefix = strings.TrimSuffix(prefix, "/")
		prefix = fmt.Sprintf("%s%s", prefix, "/")
	}

	return fmt.Sprintf("%s%s", prefix, cfg.Basket)
}

func createCaseConfig() *CaseConfig {
	return &CaseConfig{
		Header: "X-Pantry-Key",
		Id:     "715489ae-0dfd-44a4-b12b-bd7b9f69a473",
	}
}

func TestDefaultPantry(t *testing.T) {
	cfg := pantry_path.CreateConfig()
	cfgCase := createCaseConfig()

	cfgCase.Basket = "test-basket-name"
	req := caseText(t, cfgCase, cfg)

	assertPath(t, req, pantry_path.BuildPantryPathWithBasket(cfgCase.Id, cfgCase.Basket))
}

func TestPrefixPath(t *testing.T) {
	cfg := pantry_path.CreateConfig()
	cfgCase := createCaseConfig()

	cfgCase.Basket = "test-prefix-basket"
	cfgCase.PrefixPath = "this-is/the-prefix"

	req := caseText(t, cfgCase, cfg)
	assertPath(t, req, pantry_path.BuildPantryPathWithBasket(cfgCase.Id, cfgCase.Basket))
}

func TestKeyNotFound(t *testing.T) {
	cfg := pantry_path.CreateConfig()
	cfgCase := createCaseConfig()

	cfgCase.Basket = "test-nokey-basket"
	cfgCase.Header = "X-NoMatch-Key"

	req := caseText(t, cfgCase, cfg)
	assertPath(t, req, pantry_path.BuildPantryPathWithBasket("unknown", cfgCase.Basket))
}

func TestWithOutBasket(t *testing.T) {
	cfg := pantry_path.CreateConfig()
	cfgCase := createCaseConfig()

	cfgCase.Basket = ""

	req := caseText(t, cfgCase, cfg)
	assertPath(t, req, pantry_path.BuildPantryPath(cfgCase.Id))
}

func caseText(t *testing.T, cfgCase *CaseConfig, cfg *pantry_path.Config) *http.Request {
	ctx := context.Background()
	recorder := httptest.NewRecorder()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("http://localhost/%s", cfgCase.Path()), nil)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add(cfgCase.Header, cfgCase.Id)

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	handler, err := pantry_path.New(ctx, next, cfg, "pantry_path")
	if err != nil {
		t.Fatal(err)
	}

	handler.ServeHTTP(recorder, req)

	return req
}

func assertPath(t *testing.T, req *http.Request, expected string) {
	t.Helper()
	if req.URL.Path != expected {
		t.Errorf("invalid url path value: %s, expect:%s", req.URL.Path, expected)
	}
}
