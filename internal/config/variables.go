package config

import (
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/ferforastieri/mira/internal/desktop"
	"github.com/ferforastieri/mira/internal/desktop/tools"
)

// prefix starts the names of Mira's environment variables.
const prefix = "MIRA_"

// Variable is one of the MIRA_ environment variables, as the admin app
// shows them: never a secret.
type Variable struct {
	Name string
	// Value is the value in effect, redacted: the one the environment
	// sets, or the default. Hidden values are empty.
	Value string
	// Set tells whether the environment sets the variable; Value is the
	// default otherwise.
	Set bool
	// Hidden marks a value not shown, as its name tells it holds a
	// password, a token or a key.
	Hidden bool
	// Known is false for a variable Mira does not read.
	Known bool
}

// Variables describes the MIRA_ variables environ sets, in os.Environ's
// form, and those cfg, loaded from them, gives a default, by name. The
// database URL shows its host and database only; the values of variables
// named after a password, a token, a key or a secret are hidden, and the
// addresses of the others keep their host only. The variables copied into
// the settings once (see Config.Environment) are not known: Mira no
// longer reads them.
func Variables(environ []string, cfg Config) []Variable {
	set := map[string]string{}
	for _, entry := range environ {
		name, value, _ := strings.Cut(entry, "=")
		if strings.HasPrefix(name, prefix) {
			set[name] = value
		}
	}
	effective := map[string]string{
		"MIRA_DATABASE_URL":        databaseLocation(cfg.DatabaseURL),
		"MIRA_LISTEN":              cfg.Listen,
		"MIRA_DISCOVERY_PORT":      strconv.Itoa(cfg.DiscoveryPort),
		"MIRA_DLNA_PORT":           strconv.Itoa(cfg.DLNAPort),
		"MIRA_HDHOMERUN_PORT":      strconv.Itoa(cfg.HDHomeRunPort),
		"MIRA_HDHOMERUN_ADDRESSES": strings.Join(cfg.HDHomeRunAddresses, ","),
		"MIRA_FFPROBE":             cfg.FFprobe,
		"MIRA_FFMPEG":              cfg.FFmpeg,
		"MIRA_TESSERACT":           cfg.Tesseract,
		"MIRA_TESSDATA_DIR":        cfg.TessdataDir,
		"MIRA_TESSDATA_URL":        cfg.TessdataURL,
		"MIRA_DATA_DIR":            cfg.DataDir,
		"MIRA_FONTS_DIR":           cfg.FontsDir,
		"MIRA_WEB_DIR":             cfg.WebDir,
		"MIRA_ACME_DIRECTORY":      cfg.ACMEDirectory,
		"MIRA_ACME_CA_FILE":        cfg.ACMECAFile,
		// A secret, never shown: secretName hides it.
		"MIRA_SECRET_KEY": "",
	}
	// Mira Desktop's own, listed when it runs Mira only.
	if cfg.Desktop {
		effective[desktop.ServerVariable] = "1"
	}
	// The FFmpeg for older NVIDIA drivers of a test, read when set only.
	if value, ok := set[tools.OlderDriverVariable]; ok {
		effective[tools.OlderDriverVariable] = value
	}
	names := make([]string, 0, len(effective)+len(set))
	for name := range effective {
		names = append(names, name)
	}
	for name := range set {
		if _, known := effective[name]; !known {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	variables := make([]Variable, 0, len(names))
	for _, name := range names {
		raw, isSet := set[name]
		value, known := effective[name]
		v := Variable{Name: name, Set: isSet, Known: known}
		switch {
		case secretName(name):
			v.Hidden = true
		case name == "MIRA_DATABASE_URL":
			v.Value = value
		case known:
			v.Value = redactAddress(value)
		default:
			v.Value = redactAddress(raw)
		}
		variables = append(variables, v)
	}
	return variables
}

// secretName reports whether a variable's name tells its value is secret.
func secretName(name string) bool {
	name = strings.ToUpper(name)
	for _, word := range []string{"PASSWORD", "PASSWD", "TOKEN", "KEY", "SECRET", "CREDENTIAL"} {
		if strings.Contains(name, word) {
			return true
		}
	}
	return false
}

// databaseLocation is the host and database of a PostgreSQL connection
// string, a URL or keyword=value pairs, without its user and password.
func databaseLocation(connection string) string {
	if strings.Contains(connection, "://") {
		parsed, err := url.Parse(connection)
		if err != nil {
			return ""
		}
		return parsed.Host + "/" + strings.TrimPrefix(parsed.Path, "/")
	}
	var host, port, database string
	for field := range strings.FieldsSeq(connection) {
		key, value, _ := strings.Cut(field, "=")
		value = strings.Trim(value, "'")
		switch key {
		case "host":
			host = value
		case "port":
			port = value
		case "dbname":
			database = value
		}
	}
	if port != "" {
		host += ":" + port
	}
	return host + "/" + database
}

// redactAddress keeps the scheme and host of a value that is a URL, which
// may carry credentials in its other parts; other values are kept.
func redactAddress(value string) string {
	if !strings.Contains(value, "://") {
		return value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return "…"
	}
	if parsed.User == nil && (parsed.Path == "" || parsed.Path == "/") && parsed.RawQuery == "" {
		return parsed.Scheme + "://" + parsed.Host
	}
	return parsed.Scheme + "://" + parsed.Host + "/…"
}
