// Package pantrypath Pantry Path Plugin package.
package pantrypath

import (
	"context"
	"fmt"
	"net/http"
	url2 "net/url"
	"os"
	"regexp"
)

// Config the plugin configuration.
type Config struct {
	KeyHeader   string `json:"keyHeader,omitempty"`   // Pantry key source header
	BasketRegex string `json:"basketRegex,omitempty"` // Regex, from a single group, to get the target basket
}

const defaultKeyHeaderValue = "X-Pantry-Key"

const defaultBasketRegexValue = `([^/]+)/?$`

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		KeyHeader:   defaultKeyHeaderValue,
		BasketRegex: defaultBasketRegexValue,
	}
}

// New created a new Demo plugin.
func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	if len(config.KeyHeader) == 0 {
		return nil, fmt.Errorf("KeyHeader cannot be empty")
	}

	if len(config.BasketRegex) == 0 {
		return nil, fmt.Errorf("BasketRegex cannot be empty")
	}

	re, err := regexp.Compile(config.BasketRegex)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex %q: %w", config.BasketRegex, err)
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		currentPath := req.URL.RawPath
		if currentPath == "" {
			currentPath = req.URL.EscapedPath()
		}

		req.Header.Add("X-Replaced-Path", currentPath)

		var pantryID string
		var basketGroups []string
		var pantryPath string

		pantryID = req.Header.Get(config.KeyHeader)
		req.Header.Del(config.KeyHeader)
		if pantryID == "" {
			pantryID = "unknown"
			_, _ = os.Stderr.WriteString("Pantry ID not found")
			http.Error(rw, "Pantry ID not found", http.StatusBadRequest)
			return
		}

		basketGroups = re.FindStringSubmatch(req.URL.Path)
		if basketGroups == nil {
			pantryPath = BuildPantryPath(pantryID)
		} else {
			pantryPath = BuildPantryPathWithBasket(pantryID, basketGroups[1])
		}

		var err error
		req.URL.Path, err = url2.PathUnescape(pantryPath)
		if err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("[ERRO] unable to parse the new URL; newUrl=%s\n", pantryPath))
			_, _ = os.Stderr.WriteString(err.Error())
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}
		req.RequestURI = req.URL.RequestURI()

		next.ServeHTTP(rw, req)
	}), nil
}

// BuildPantryPath , build path to ROOT pantry storage.
func BuildPantryPath(key string) string {
	return fmt.Sprintf("/apiv1/pantry/%s", key)
}

// BuildPantryPathWithBasket , build path to BASKET on pantry storage.
func BuildPantryPathWithBasket(key, basket string) string {
	return fmt.Sprintf("/apiv1/pantry/%s/basket/%s", key, basket)
}
