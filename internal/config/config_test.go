package config

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ferforastieri/mira/internal/accounts"
)

func env(values map[string]string) func(string) string {
	return func(key string) string { return values[key] }
}

// database is the one variable every configuration needs.
const database = "postgresql://mira@db/mira"

// with is the variables of values, with the database URL.
func with(values map[string]string) func(string) string {
	all := map[string]string{"MIRA_DATABASE_URL": database}
	for name, value := range values {
		all[name] = value
	}
	return env(all)
}

func TestLoadAcceptsExplicitSettings(t *testing.T) {
	t.Parallel()
	// The first GPU method of this system, nvenc on Linux and Windows,
	// videotoolbox on macOS.
	method := accounts.HardwareAccelerations[1]
	cfg, err := Load(with(map[string]string{
		"MIRA_LISTEN":       "[::1]:9000",
		"MIRA_LOG_LEVEL":    "DEBUG",
		"MIRA_HWACCEL":      " " + strings.ToUpper(method) + " ",
		"MIRA_VAAPI_DEVICE": "/dev/dri/renderD129",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Listen != "[::1]:9000" || !cfg.Environment.DetailedLog || cfg.Environment.Hardware != method ||
		cfg.Environment.VAAPIDevice != "/dev/dri/renderD129" || len(cfg.Warnings) != 0 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
	// Without a choice, the settings keep theirs.
	if cfg, err := Load(with(nil)); err != nil || cfg.Environment.Hardware != "" || cfg.Environment.DetailedLog {
		t.Errorf("without variables: %+v, %v", cfg.Environment, err)
	}
	// Another level than debug leaves the detailed log as it is.
	for _, level := range []string{"info", "warn", "error"} {
		if cfg, err := Load(with(map[string]string{"MIRA_LOG_LEVEL": level})); err != nil || cfg.Environment.DetailedLog || len(cfg.Warnings) != 0 {
			t.Errorf("%s: %+v, %v", level, cfg.Environment, err)
		}
	}
}

func TestFontsComeFromTheSystemFolderUnlessSet(t *testing.T) {
	t.Parallel()
	for value, want := range map[string]string{"": "/usr/share/fonts", " /fonts ": "/fonts"} {
		cfg, err := Load(with(map[string]string{"MIRA_FONTS_DIR": value}))
		if err != nil || cfg.FontsDir != want {
			t.Errorf("MIRA_FONTS_DIR %q: %q, %v", value, cfg.FontsDir, err)
		}
	}
}

func TestWebClientComesFromTheImageFolderUnlessSet(t *testing.T) {
	t.Parallel()
	for value, want := range map[string]string{"": DefaultWebDir, " /srv/jellyfin-web ": "/srv/jellyfin-web"} {
		cfg, err := Load(with(map[string]string{"MIRA_WEB_DIR": value}))
		if err != nil || cfg.WebDir != want {
			t.Errorf("MIRA_WEB_DIR %q: %q, %v", value, cfg.WebDir, err)
		}
	}
}

// Discovery listens on Jellyfin's port unless another is set; a value that
// is not a port stops the start.
func TestDiscoveryPortIsJellyfinsUnlessSet(t *testing.T) {
	t.Parallel()
	for value, want := range map[string]int{"": 7359, " 18420 ": 18420, "65535": 65535} {
		cfg, err := Load(with(map[string]string{"MIRA_DISCOVERY_PORT": value}))
		if err != nil || cfg.DiscoveryPort != want {
			t.Errorf("MIRA_DISCOVERY_PORT %q: %d, %v", value, cfg.DiscoveryPort, err)
		}
	}
	for _, value := range []string{"0", "65536", "-1", "port"} {
		if _, err := Load(with(map[string]string{"MIRA_DISCOVERY_PORT": value})); err == nil ||
			!strings.Contains(err.Error(), "MIRA_DISCOVERY_PORT") {
			t.Errorf("MIRA_DISCOVERY_PORT %q: %v", value, err)
		}
	}
}

// Find tuners asks HDHomeRun's port unless another is set, by broadcast
// unless addresses are given: a host alone is asked on that port.
func TestTunerDiscoveryFollowsItsVariables(t *testing.T) {
	t.Parallel()
	cfg, err := Load(with(map[string]string{}))
	if err != nil || cfg.HDHomeRunPort != 65001 || cfg.HDHomeRunAddresses != nil {
		t.Errorf("defaults: %d %q %v", cfg.HDHomeRunPort, cfg.HDHomeRunAddresses, err)
	}
	cfg, err = Load(with(map[string]string{"MIRA_HDHOMERUN_PORT": "18920", "MIRA_HDHOMERUN_ADDRESSES": " 192.168.1.20, tuner.lan:18921,,[::1]:5 "}))
	if err != nil || cfg.HDHomeRunPort != 18920 || !slices.Equal(cfg.HDHomeRunAddresses, []string{"192.168.1.20:18920", "tuner.lan:18921", "[::1]:5"}) {
		t.Errorf("set: %d %q %v", cfg.HDHomeRunPort, cfg.HDHomeRunAddresses, err)
	}
	for variable, value := range map[string]string{"MIRA_HDHOMERUN_PORT": "0", "MIRA_HDHOMERUN_ADDRESSES": "http://192.168.1.20"} {
		if _, err := Load(with(map[string]string{variable: value})); err == nil || !strings.Contains(err.Error(), variable) {
			t.Errorf("%s %q: %v", variable, value, err)
		}
	}
}

func TestLoadReportsEveryInvalidSetting(t *testing.T) {
	t.Parallel()
	_, err := Load(env(map[string]string{
		"MIRA_LISTEN":   "8096",
		"MIRA_DATA_DIR": "data",
	}))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"MIRA_DATABASE_URL", "MIRA_LISTEN", "MIRA_DATA_DIR"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %s", err, want)
		}
	}
}

// The variables copied into the settings never stop the start: a value
// that is not valid is left out, with a warning.
func TestCopiedVariablesOnlyWarn(t *testing.T) {
	t.Parallel()
	invalid := map[string]string{
		"MIRA_LOG_LEVEL":      "verbose",
		"MIRA_CACHE_SIZE":     "lots",
		"MIRA_HWACCEL":        "opencl",
		"MIRA_SEGMENTS":       "theintrodb,other",
		"MIRA_VAAPI_DEVICE":   "/dev/dri/card0",
		"MIRA_RECORDINGS_DIR": "recordings",
		"MIRA_BACKUP_DIR":     "/backups\x01",
	}
	cfg, err := Load(with(invalid))
	if err != nil {
		t.Fatal(err)
	}
	if e := cfg.Environment; e.Hardware != "" || e.Segments != nil || e.CacheSizeGB != 0 || e.DetailedLog || e.VAAPIDevice != "" ||
		e.RecordingsFolder != "" || e.BackupFolder != "" {
		t.Errorf("invalid values copied: %+v", e)
	}
	for name := range invalid {
		if !slices.ContainsFunc(cfg.Warnings, func(err error) bool { return strings.HasPrefix(err.Error(), name+":") }) {
			t.Errorf("no warning for %s: %v", name, cfg.Warnings)
		}
	}
	if len(cfg.Warnings) != len(invalid) {
		t.Errorf("warnings: %v", cfg.Warnings)
	}
}

// The variables Mira no longer reads are listed when they are set, with
// the section of the settings that holds them now.
func TestRetiredVariablesAreListed(t *testing.T) {
	t.Parallel()
	cfg, err := Load(with(map[string]string{"MIRA_CACHE_SIZE": "20GB", "MIRA_BACKUP_DIR": "/backups", "MIRA_HWACCEL": "nvenc"}))
	if err != nil {
		t.Fatal(err)
	}
	want := []Retired{{"MIRA_CACHE_SIZE", "Settings › Playback"}, {"MIRA_BACKUP_DIR", "Settings › Backups"}}
	if !slices.Equal(cfg.Retired, want) {
		t.Errorf("retired: %v, want %v", cfg.Retired, want)
	}
	if cfg, _ := Load(with(nil)); len(cfg.Retired) != 0 {
		t.Errorf("retired without variables: %v", cfg.Retired)
	}
}

func TestSegmentDatabasesKeepTheOrderOfPreference(t *testing.T) {
	t.Parallel()
	for text, want := range map[string][]string{
		" IntroDB , theintrodb ":            {"introdb", "theintrodb"},
		"PublicMetaDB,introdb,publicmetadb": {"publicmetadb", "introdb"},
		"introdb,introdb":                   {"introdb"},
		"SkipDB,theintrodb":                 {"skipdb", "theintrodb"},
	} {
		cfg, err := Load(with(map[string]string{"MIRA_SEGMENTS": text}))
		if err != nil || !slices.Equal(cfg.Environment.Segments, want) {
			t.Errorf("%q: %v, %v; want %v", text, cfg.Environment.Segments, err, want)
		}
	}
	// Not set is nil, none an empty list.
	if cfg, _ := Load(with(nil)); cfg.Environment.Segments != nil {
		t.Errorf("not set: %#v", cfg.Environment.Segments)
	}
	if cfg, _ := Load(with(map[string]string{"MIRA_SEGMENTS": "none"})); cfg.Environment.Segments == nil || len(cfg.Environment.Segments) != 0 {
		t.Errorf("none: %#v", cfg.Environment.Segments)
	}
	// Turning the databases off cannot be mixed with naming one.
	if cfg, _ := Load(with(map[string]string{"MIRA_SEGMENTS": "none,introdb"})); cfg.Environment.Segments != nil || len(cfg.Warnings) != 1 {
		t.Error("none with a database was accepted")
	}
}

func TestCacheSizesReadAsPeopleWriteThem(t *testing.T) {
	t.Parallel()
	for text, want := range map[string]int{
		"":            0,
		"20GB":        20,
		"500 mb":      1,
		"1.5TB":       1500,
		"8GiB":        9,
		"10g":         11,
		"21474836480": 21,
	} {
		cfg, err := Load(with(map[string]string{"MIRA_CACHE_SIZE": text}))
		if err != nil || cfg.Environment.CacheSizeGB != want || len(cfg.Warnings) != 0 {
			t.Errorf("%q: %d %v %v, want %d", text, cfg.Environment.CacheSizeGB, err, cfg.Warnings, want)
		}
	}
	for _, text := range []string{"10XB", "-1GB", "100MB", "GB", "3TB"} {
		if cfg, err := Load(with(map[string]string{"MIRA_CACHE_SIZE": text})); err != nil || cfg.Environment.CacheSizeGB != 0 || len(cfg.Warnings) != 1 {
			t.Errorf("%q: %d %v %v", text, cfg.Environment.CacheSizeGB, err, cfg.Warnings)
		}
	}
}

func TestLoadRejectsUnusablePorts(t *testing.T) {
	t.Parallel()
	for _, listen := range []string{":0", ":65536", ":http", "host:"} {
		if _, err := Load(with(map[string]string{"MIRA_LISTEN": listen})); err == nil {
			t.Errorf("MIRA_LISTEN=%q was accepted", listen)
		}
	}
}

// Mira keeps its files in MIRA_DATA_DIR, by default a folder of the
// system's temporary folder.
func TestDataFolder(t *testing.T) {
	t.Parallel()
	cfg, err := Load(with(nil))
	if want := filepath.Join(os.TempDir(), "mira"); err != nil || cfg.DataDir != want {
		t.Errorf("default: %q %v", cfg.DataDir, err)
	}
	cfg, err = Load(with(map[string]string{"MIRA_DATA_DIR": " " + systemPath("/data/") + " "}))
	if err != nil || cfg.DataDir != systemPath("/data") {
		t.Errorf("set: %q %v", cfg.DataDir, err)
	}
	for _, relative := range []string{"data", "./data"} {
		if _, err := Load(with(map[string]string{"MIRA_DATA_DIR": relative})); err == nil || !strings.Contains(err.Error(), "MIRA_DATA_DIR") {
			t.Errorf("%q: %v", relative, err)
		}
	}
}
