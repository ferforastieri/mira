package loginitem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func noEnv(string) string { return "" }

// macOS: a launch agent that opens the bundle at login, once.
func TestLaunchAgent(t *testing.T) {
	t.Parallel()
	args := command("darwin", "/Applications/Mira.app/Contents/MacOS/mira-desktop", noEnv)
	if strings.Join(args, " ") != "/usr/bin/open -a /Applications/Mira.app --args --background" {
		t.Errorf("bundle command = %q", args)
	}
	if got := command("darwin", "/Users/a/dev/mira-desktop", noEnv); strings.Join(got, " ") != "/Users/a/dev/mira-desktop --background" {
		t.Errorf("command outside a bundle = %q", got)
	}
	plist := string(LaunchAgent([]string{"/usr/bin/open", "-a", "/Users/a & b/Mira.app", "--args", Arg}))
	for _, want := range []string{
		"<key>Label</key>\n\t<string>io.github.ferforastieri.mira-desktop</string>",
		"<string>/Users/a &amp; b/Mira.app</string>",
		"<string>--background</string>",
		"<key>RunAtLoad</key>\n\t<true/>",
	} {
		if !strings.Contains(plist, want) {
			t.Errorf("the launch agent lacks %q:\n%s", want, plist)
		}
	}
	if strings.Contains(plist, "KeepAlive") {
		t.Error("the launch agent would start the app again after it quits")
	}
	if got := LaunchAgentPath("/Users/a"); got != "/Users/a/Library/LaunchAgents/io.github.ferforastieri.mira-desktop.plist" {
		t.Errorf("LaunchAgentPath = %q", got)
	}
}

// Windows: a value of the Run key, its executable quoted.
func TestRunValue(t *testing.T) {
	t.Parallel()
	args := command("windows", `C:\Program Files\Mira\mira-desktop.exe`, noEnv)
	if got := RunValue(args); got != `"C:\Program Files\Mira\mira-desktop.exe" --background` {
		t.Errorf("RunValue = %q", got)
	}
}

// Linux: an autostart entry, of the AppImage's own file when it runs from
// one.
func TestDesktopEntry(t *testing.T) {
	t.Parallel()
	env := func(name string) string {
		if name == "APPIMAGE" {
			return "/home/a/Apps/Mira Desktop.AppImage"
		}
		return ""
	}
	entry := string(DesktopEntry(command("linux", "/tmp/.mount_Polyf/usr/bin/mira-desktop", env)))
	for _, want := range []string{"[Desktop Entry]\n", "Type=Application\n", `Exec="/home/a/Apps/Mira Desktop.AppImage" --background` + "\n", "Terminal=false\n"} {
		if !strings.Contains(entry, want) {
			t.Errorf("the entry lacks %q:\n%s", want, entry)
		}
	}
	if got := quoteExec(`/opt/a"b$c 100%`); got != `"/opt/a\\"b\\$c 100%%"` {
		t.Errorf("quoteExec = %s", got)
	}
	if got := command("linux", "/usr/bin/mira-desktop", noEnv); strings.Join(got, " ") != "/usr/bin/mira-desktop --background" {
		t.Errorf("command = %q", got)
	}
	if got := AutostartPath("/home/a", noEnv); got != "/home/a/.config/autostart/mira-desktop.desktop" {
		t.Errorf("AutostartPath = %q", got)
	}
	xdg := func(string) string { return "/home/a/conf" }
	if got := AutostartPath("/home/a", xdg); got != "/home/a/conf/autostart/mira-desktop.desktop" {
		t.Errorf("AutostartPath with XDG_CONFIG_HOME = %q", got)
	}
}

// The file of macOS and Linux is written, found and removed.
func TestFileItem(t *testing.T) {
	t.Parallel()
	item := fileItem{path: filepath.Join(t.TempDir(), "autostart", "mira-desktop.desktop"), content: []byte("entry")}
	if on, err := item.Enabled(); err != nil || on {
		t.Fatalf("Enabled before: %v, %v", on, err)
	}
	if err := item.Enable(); err != nil {
		t.Fatal(err)
	}
	if data, err := os.ReadFile(item.path); err != nil || string(data) != "entry" {
		t.Fatalf("written: %q, %v", data, err)
	}
	if on, err := item.Enabled(); err != nil || !on {
		t.Fatalf("Enabled after: %v, %v", on, err)
	}
	if err := item.Disable(); err != nil {
		t.Fatal(err)
	}
	if err := item.Disable(); err != nil {
		t.Fatalf("a second Disable: %v", err)
	}
	if on, _ := item.Enabled(); on {
		t.Error("still enabled")
	}
}
