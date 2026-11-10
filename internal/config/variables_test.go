package config

import (
	"strings"
	"testing"
)

func TestVariablesShowTheConfigurationWithoutSecrets(t *testing.T) {
	t.Parallel()
	environ := []string{
		"MIRA_DATABASE_URL=postgresql://mira:s3cret-pw@db.internal:5432/media?sslmode=disable",
		"MIRA_CACHE_SIZE=20GB",
		"MIRA_API_TOKEN=tok-123",
		"MIRA_ADMIN_PASSWORD=hunter2",
		"MIRA_SECRET_KEY=MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY=",
		"MIRA_EXTRA_FEED=https://user:pw@feeds.example/list?key=abc",
		"MIRA_NVENC13_TOOLS=/tmp/nvenc13.json",
		"HOME=/root",
	}
	getenv := func(name string) string {
		for _, entry := range environ {
			if key, value, _ := strings.Cut(entry, "="); key == name {
				return value
			}
		}
		return ""
	}
	cfg, err := Load(getenv)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Variable{}
	for _, v := range Variables(environ, cfg) {
		byName[v.Name] = v
		for _, secret := range []string{"s3cret-pw", "mira:", "tok-123", "hunter2", "pw@", "abc", "sslmode", "MDEy"} {
			if strings.Contains(v.Value, secret) {
				t.Errorf("%s shows %q: %q", v.Name, secret, v.Value)
			}
		}
	}
	if _, listed := byName["HOME"]; listed {
		t.Error("a variable of another program is listed")
	}
	for name, want := range map[string]Variable{
		"MIRA_DATABASE_URL": {Value: "db.internal:5432/media", Set: true, Known: true},
		// Copied into the settings once, it is no longer read.
		"MIRA_CACHE_SIZE":     {Value: "20GB", Set: true},
		"MIRA_LISTEN":         {Value: ":8096", Known: true},
		"MIRA_DATA_DIR":       {Value: cfg.DataDir, Known: true},
		"MIRA_API_TOKEN":      {Set: true, Hidden: true},
		"MIRA_ADMIN_PASSWORD": {Set: true, Hidden: true},
		"MIRA_SECRET_KEY":     {Set: true, Hidden: true, Known: true},
		"MIRA_EXTRA_FEED":     {Value: "https://feeds.example/…", Set: true},
		// The FFmpeg for older NVIDIA drivers of a test: read when set.
		"MIRA_NVENC13_TOOLS": {Value: "/tmp/nvenc13.json", Set: true, Known: true},
	} {
		got := byName[name]
		want.Name = name
		if got != want {
			t.Errorf("%s: got %+v, want %+v", name, got, want)
		}
	}
	for _, retired := range []string{"MIRA_HWACCEL", "MIRA_SEGMENTS", "MIRA_LOG_LEVEL", "MIRA_CACHE_DIR"} {
		if _, listed := byName[retired]; listed {
			t.Errorf("%s is listed while not set", retired)
		}
	}
	if location := databaseLocation("host=db.internal port=5433 user=mira password=s3cret dbname=media"); location != "db.internal:5433/media" {
		t.Errorf("keyword form: %q", location)
	}
}
