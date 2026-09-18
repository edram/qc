package cookies

import (
	"fmt"
	"strings"

	"github.com/edram/qi/internal/app"
)

// Path returns a per-source, per-profile cookie cache under the application directory.
func Path(source, profile string) (string, error) {
	if profile == "" {
		profile = "default"
	}
	if profile == "." || profile == ".." || strings.ContainsAny(profile, `/\`) {
		return "", fmt.Errorf("invalid cookie profile %q", profile)
	}
	return app.Path("cookies", source+"."+profile+".json")
}
