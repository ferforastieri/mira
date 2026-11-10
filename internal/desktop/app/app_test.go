package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/ferforastieri/mira/internal/desktop"
	"github.com/ferforastieri/mira/internal/desktop/keystore"
	"github.com/ferforastieri/mira/internal/desktop/tools"
)

// The test binary is also the fake server the app runs, when
// FAKE_MIRA_SERVER is set: it answers the admin app's status on
// MIRA_LISTEN, logs a setup code as the server does, records its
// environment, and stops on SIGTERM or Ctrl-Break. FAKE_MIRA_EXIT makes
// its first run end by itself. FAKE_MIRA_BROKEN makes it a broken
// build's: it marks the database as migrated, then ends before it
// answers. With FAKE_DESKTOP_APP set, it is an app holding a data folder
// instead (see quit_test.go).
func TestMain(m *testing.M) {
	if os.Getenv("FAKE_MIRA_SERVER") != "" {
		fakeServer()
		return
	}
	if os.Getenv("FAKE_DESKTOP_APP") != "" {
		fakeApp()
		return
	}
	os.Exit(m.Run())
}

func fakeServer() {
	data := os.Getenv("MIRA_DATA_DIR")
	env, _ := json.Marshal(os.Environ())
	_ = os.WriteFile(filepath.Join(data, "fake-env.json"), env, 0o600)
	runs := filepath.Join(data, "fake-runs")
	previous, _ := os.ReadFile(runs)
	_ = os.WriteFile(runs, append(previous, 'x'), 0o600)
	if os.Getenv("FAKE_MIRA_EXIT") != "" && len(previous) == 0 {
		fmt.Fprintln(os.Stderr, "fake: ending by itself")
		os.Exit(3)
	}
	if os.Getenv("FAKE_MIRA_BROKEN") != "" {
		_ = os.WriteFile(filepath.Join(data, "postgresql", "migrated-by-the-broken-build"), nil, 0o600)
		fmt.Fprintln(os.Stderr, "fake: a broken build ends before it answers")
		os.Exit(2)
	}
	fmt.Fprintln(os.Stderr, `time=2026-10-18T10:00:00.000Z level=WARN msg="No administrator yet: open /admin/ and create one with this setup code" setup_code=AB3D-7KQ9`)
	listener, err := net.Listen("tcp", os.Getenv("MIRA_LISTEN"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "fake:", err)
		os.Exit(1)
	}
	go func() {
		_ = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"database":"ready","setupRequired":true}`))
		}))
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	fmt.Fprintln(os.Stderr, "fake: stopping")
}

type fakeKeys struct{ secrets map[string]string }

func (f *fakeKeys) Get(service, account string) (string, error) {
	if s, ok := f.secrets[service+account]; ok {
		return s, nil
	}
	return "", keystore.ErrNotFound
}

func (f *fakeKeys) Set(service, account, secret string) error {
	f.secrets[service+account] = secret
	return nil
}

type fakeLogin struct{ on bool }

func (f *fakeLogin) Enabled() (bool, error) { return f.on, nil }
func (f *fakeLogin) Enable() error          { f.on = true; return nil }
func (f *fakeLogin) Disable() error         { f.on = false; return nil }

// install lays out an install folder: the test binary as the server, and
// the PostgreSQL programs of MIRA_TEST_POSTGRES_BIN.
func install(t *testing.T) desktop.Install {
	t.Helper()
	bin := os.Getenv("MIRA_TEST_POSTGRES_BIN")
	if bin == "" {
		t.Skip("MIRA_TEST_POSTGRES_BIN is not set")
	}
	dir := t.TempDir()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(exe, filepath.Join(dir, desktop.Executable("mira"))); err != nil {
		t.Skipf("no symbolic links: %v", err)
	}
	if err := os.Symlink(filepath.Dir(bin), filepath.Join(dir, "postgresql")); err != nil {
		t.Fatal(err)
	}
	return desktop.Install{Dir: dir}
}

func listenPort(t *testing.T) string {
	if port := os.Getenv("MIRA_TEST_DESKTOP_PORT"); port != "" {
		return port
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	return strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)
}

func databasePort(t *testing.T) int {
	port, _ := strconv.Atoi(os.Getenv("MIRA_TEST_POSTGRES_PORT"))
	return port
}

// waitFor waits until the app's status satisfies done, for up to a
// minute: CI runners can be slow.
func waitFor(t *testing.T, a *App, what string, done func(Status) bool) Status {
	t.Helper()
	deadline := time.Now().Add(time.Minute)
	for time.Now().Before(deadline) {
		if s := a.Status(); done(s) {
			return s
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("%s: status stayed %+v", what, a.Status())
	return Status{}
}

// readyTools lays out FFmpeg and ffprobe in the data folder, as after a
// first start, and returns their manifest and folder.
func readyTools(t *testing.T, data string) (tools.Manifest, string) {
	t.Helper()
	toolsDir := filepath.Join(data, "tools", "ffmpeg-9.0.2")
	for _, p := range []string{"bin/ffmpeg", "bin/ffprobe"} {
		if err := os.MkdirAll(filepath.Join(toolsDir, "bin"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(toolsDir, p), []byte("x"), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	manifest := tools.Manifest{Version: "9.0.2", FFmpeg: "bin/ffmpeg", FFprobe: "bin/ffprobe",
		Files: []tools.File{{URL: "http://127.0.0.1:1/never", SHA256: strings.Repeat("a", 64), Size: 1, Format: "zip"}}}
	if err := os.WriteFile(filepath.Join(toolsDir, ".mira-tools"), []byte(strings.Repeat("a", 64)+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return manifest, toolsDir
}

// The app starts the cluster, then the server with its environment; shows
// the setup code; opens the setup once; starts the server again when it
// ends by itself; and stops both, leaving no process.
func TestAppRunsAndStopsEverything(t *testing.T) {
	inst := install(t)
	data := t.TempDir()
	port := listenPort(t)
	t.Setenv("MIRA_LISTEN", "127.0.0.1:"+port)
	t.Setenv("FAKE_MIRA_SERVER", "1")
	t.Setenv("FAKE_MIRA_EXIT", "1")
	t.Setenv("MIRA_SECRET_KEY", "from the environment, left out")
	manifest, toolsDir := readyTools(t, data)
	keys := &fakeKeys{secrets: map[string]string{}}
	// The app opens the setup from its own goroutine.
	var mu sync.Mutex
	var opened []string
	openedNow := func() []string {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(opened)
	}
	login := &fakeLogin{}
	a := New(Options{Version: "0.42.0", Install: inst, Folders: desktop.Folders{Data: data}, Keys: keys, Tools: manifest,
		DatabasePort: databasePort(t), Login: login, OpenSetup: true, Open: func(u string) error {
			mu.Lock()
			defer mu.Unlock()
			opened = append(opened, u)
			return nil
		}})
	a.Start()
	t.Cleanup(func() { a.Stop(context.Background()) })
	status := waitFor(t, a, "running", func(s Status) bool { return s.Phase == Running && s.SetupCode != "" })
	if status.SetupCode != "AB3D-7KQ9" || !status.SetupRequired || len(status.Problems) != 0 {
		t.Errorf("status = %+v", status)
	}
	if runs, _ := os.ReadFile(filepath.Join(data, "fake-runs")); len(runs) != 2 {
		t.Errorf("the server ran %d times, want 2: once ending by itself, then again", len(runs))
	}
	waitFor(t, a, "setup opened", func(Status) bool { return len(openedNow()) > 0 })
	if opened := openedNow(); opened[0] != "http://127.0.0.1:"+port+"/admin/" {
		t.Errorf("opened %v", opened)
	}
	if login.on {
		t.Error("Start with the computer was turned on before the setup")
	}

	var env []string
	raw, err := os.ReadFile(filepath.Join(data, "fake-env.json"))
	if err != nil || json.Unmarshal(raw, &env) != nil {
		t.Fatalf("the server's environment: %v", err)
	}
	key := keys.secrets[keystore.Service+keystore.Account(data)]
	for _, want := range []string{
		"MIRA_DATA_DIR=" + data,
		"MIRA_SECRET_KEY=" + key,
		"MIRA_FFMPEG=" + filepath.Join(toolsDir, "bin", "ffmpeg"),
		"MIRA_WEB_DIR=" + inst.WebDir(),
		"MIRA_FONTS_DIR=" + inst.FontsDir(),
	} {
		if !slices.Contains(env, want) {
			t.Errorf("the server's environment lacks %s", want)
		}
	}
	var databaseURL, path string
	for _, entry := range env {
		if v, ok := strings.CutPrefix(entry, "MIRA_DATABASE_URL="); ok {
			databaseURL = v
		}
		if v, ok := strings.CutPrefix(entry, "PATH="); ok {
			path = v
		}
		if strings.Contains(entry, "from the environment") {
			t.Error("the environment's secret key reached the server")
		}
	}
	if !strings.HasPrefix(databaseURL, "postgresql://mira:") || !strings.HasPrefix(path, inst.PostgresBin()+string(os.PathListSeparator)) {
		t.Errorf("database URL set %v, PATH %q", databaseURL != "", path)
	}

	pidData, err := os.ReadFile(filepath.Join(data, "mira.pid"))
	if err != nil && runtime.GOOS != "windows" {
		t.Fatalf("no server PID file: %v", err)
	}
	serverPID, _ := strconv.Atoi(strings.Fields(string(pidData) + " 0")[0])
	clusterPIDData, _ := os.ReadFile(filepath.Join(data, "postgresql", "postmaster.pid"))
	clusterPID, _ := strconv.Atoi(strings.SplitN(string(clusterPIDData), "\n", 2)[0])

	a.Stop(context.Background())
	if s := a.Status(); s.Phase != Stopped {
		t.Errorf("phase after Stop = %s", s.Phase)
	}
	// pg_ctl stop returns once the postmaster removed its lock file, just
	// before it ends: each process gets a few seconds to be gone.
	for name, pid := range map[string]int{"server": serverPID, "cluster": clusterPID} {
		deadline := time.Now().Add(10 * time.Second)
		for pid != 0 && alive(pid) && time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
		}
		if pid != 0 && alive(pid) {
			t.Errorf("the %s, process %d, still runs", name, pid)
		}
	}
	if _, err := os.Stat(filepath.Join(data, "postgresql", "postmaster.pid")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("the cluster's lock file is left: %v", err)
	}
	if log, _ := os.ReadFile(filepath.Join(data, "log", "mira.log")); !strings.Contains(string(log), "fake: stopping") {
		t.Errorf("the server's log lacks its stop:\n%s", log)
	}

	if copied, err := a.SecretKey(); err != nil || copied != key {
		t.Errorf("SecretKey = %v, the stored key: %v", err, copied == key)
	}

	// A second start reuses the key, and opens the setup again.
	a.Start()
	waitFor(t, a, "running again", func(s Status) bool { return s.Phase == Running })
	if len(keys.secrets) != 1 || keys.secrets[keystore.Service+keystore.Account(data)] != key {
		t.Error("the second start made another key")
	}
	waitFor(t, a, "setup opened again", func(Status) bool { return len(openedNow()) >= 1 })
}

func TestServerEnvironment(t *testing.T) {
	t.Parallel()
	o := Options{Install: desktop.Install{Dir: "/app"}, Folders: desktop.Folders{Data: "/data"}}
	env := serverEnvironment([]string{"PATH=/usr/bin", "HOME=/home/a", "MIRA_DATABASE_URL=postgresql://elsewhere", "MIRA_FFMPEG=/opt/ffmpeg",
		"PGHOST=elsewhere", "MIRA_LISTEN=:19001"}, o, "postgresql://mira:pw@127.0.0.1:5/mira", "", tools.Paths{})
	joined := strings.Join(env, "\n")
	for _, want := range []string{"HOME=/home/a", "MIRA_LISTEN=:19001", "MIRA_DATABASE_URL=postgresql://mira:pw@127.0.0.1:5/mira",
		"PATH=" + filepath.Join("/app", "postgresql", "bin") + string(os.PathListSeparator) + "/usr/bin", "MIRA_DESKTOP=1"} {
		if !slices.Contains(env, want) {
			t.Errorf("environment lacks %s:\n%s", want, joined)
		}
	}
	for _, unwanted := range []string{"elsewhere", "/opt/ffmpeg", "MIRA_SECRET_KEY"} {
		if strings.Contains(joined, unwanted) {
			t.Errorf("environment holds %s:\n%s", unwanted, joined)
		}
	}
}

func TestListen(t *testing.T) {
	for _, c := range []struct{ listen, ask, open string }{
		{"", "127.0.0.1:8096", "localhost:8096"},
		{":19001", "127.0.0.1:19001", "localhost:19001"},
		{"0.0.0.0:8100", "127.0.0.1:8100", "localhost:8100"},
		{"192.168.1.5:8096", "192.168.1.5:8096", "192.168.1.5:8096"},
	} {
		t.Setenv("MIRA_LISTEN", c.listen)
		if ask, open := Listen(); ask != c.ask || open != c.open {
			t.Errorf("Listen(%q) = %s, %s", c.listen, ask, open)
		}
	}
	t.Setenv("MIRA_LISTEN", ":19001")
	if AdminURL() != "http://localhost:19001/admin/" || WebURL() != "http://localhost:19001/web/" {
		t.Errorf("AdminURL %s, WebURL %s", AdminURL(), WebURL())
	}
}

func TestSetupCode(t *testing.T) {
	t.Parallel()
	line := `time=2026-10-18T10:00:00Z level=WARN msg="No administrator yet: open /admin/ and create one with this setup code" setup_code=7K2M-QX4P`
	if m := setupCode.FindStringSubmatch(line); m == nil || m[1] != "7K2M-QX4P" {
		t.Errorf("match = %v", m)
	}
	if setupCode.MatchString(`msg="a user logged setup_code=hello"`) {
		t.Error("matched a line without a code")
	}
}

// The rebuild of the database's indexes shows while Mira starts, and
// changes no other phase.
func TestRebuildingPhase(t *testing.T) {
	t.Parallel()
	a := New(Options{Folders: desktop.Folders{Data: t.TempDir()}})
	a.update(func(s *Status) { s.Phase = Starting })
	a.rebuilding(true)
	if got := a.Status().Phase; got != Rebuilding {
		t.Fatalf("phase while rebuilding = %s", got)
	}
	a.rebuilding(false)
	if got := a.Status().Phase; got != Starting {
		t.Fatalf("phase once rebuilt = %s", got)
	}
	a.update(func(s *Status) { s.Phase = Stopping })
	a.rebuilding(true)
	a.rebuilding(false)
	if got := a.Status().Phase; got != Stopping {
		t.Errorf("phase while stopping = %s", got)
	}
}

// Start with the computer turns on once set up, unless chosen before, and
// the choice is kept.
func TestStartWithComputer(t *testing.T) {
	t.Parallel()
	data := t.TempDir()
	login := &fakeLogin{}
	a := New(Options{Folders: desktop.Folders{Data: data}, Login: login})
	a.setUp()
	if !login.on || !a.Status().StartWithComputer {
		t.Fatal("not turned on once set up")
	}
	if err := a.SetStartWithComputer(false); err != nil {
		t.Fatal(err)
	}
	a.setUp()
	if login.on {
		t.Error("turned on again after it was turned off")
	}
	again := New(Options{Folders: desktop.Folders{Data: data}, Login: login})
	if again.Status().StartWithComputer {
		t.Error("the choice was not kept")
	}
	if err := New(Options{Folders: desktop.Folders{Data: t.TempDir()}}).SetStartWithComputer(true); err == nil {
		t.Error("a run without login item changed it")
	}
}

// A second app on the same data folder is refused while the first runs.
//
//nolint:paralleltest // a child another test forks holds the lock file until it execs, keeping its lock past the release
func TestLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "desktop.lock")
	release, err := Lock(path)
	if err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		// flock locks belong to the open file: a second open is a second
		// holder, as another app's would be.
		if _, err := Lock(path); !errors.Is(err, ErrRunning) {
			t.Errorf("second Lock = %v", err)
		}
	}
	release()
	again, err := Lock(path)
	if err != nil {
		t.Fatalf("Lock after release: %v", err)
	}
	again()
}

func TestRotateLogs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	big := filepath.Join(dir, "mira.log")
	if err := os.WriteFile(big, make([]byte, 10<<20+1), 0o600); err != nil {
		t.Fatal(err)
	}
	small := filepath.Join(dir, "desktop.log")
	if err := os.WriteFile(small, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	RotateLogs(dir)
	if _, err := os.Stat(big + ".1"); err != nil {
		t.Error("the large log was not rotated")
	}
	if _, err := os.Stat(small); err != nil {
		t.Error("the small log was rotated")
	}
}
