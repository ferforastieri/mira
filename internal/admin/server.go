package admin

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/ferforastieri/mira/internal/desktop"
	"github.com/ferforastieri/mira/internal/updates"
	"github.com/ferforastieri/mira/internal/updatetool"
)

// The server's actions in the admin app: Restart, the install of the
// update Mira Desktop offers, and Update now, through an update tool
// the administrator runs beside Mira. Each is an administrator's
// alone, confirmed with their password or a two-step code (see
// checkIdentity), allowed once a minute, and recorded in the activity log.
// None stops Mira for good: a container's restart policy, or Mira
// Desktop, would start it again at once.

// What starts Mira again once it exits, as the server's answer tells
// it: Mira Desktop, a restart policy or service manager an
// administrator confirmed in the settings, or nothing Mira knows of. A
// container tells nobody inside it its restart policy: only the Docker
// socket does, which Mira is never given.
const (
	restartDesktop   = "desktop"
	restartConfirmed = "confirmed"
	restartUnknown   = "unknown"
)

// serverJSON is how the server's actions stand. StartedAt changes once a
// restart has happened. System is where Mira runs, as the status names
// it. DesktopUpdate is the update Mira Desktop offers, null for none or
// elsewhere; UpdateTool tells whether the settings give an update tool.
type serverJSON struct {
	Version       string             `json:"version"`
	StartedAt     time.Time          `json:"startedAt"`
	Restart       string             `json:"restart"`
	System        string             `json:"system"`
	DesktopUpdate *desktopUpdateJSON `json:"desktopUpdate"`
	UpdateTool    bool               `json:"updateTool"`
}

// desktopUpdateJSON is the update Mira Desktop offers: its version and
// release notes, whether Mira Desktop is downloading it, how much, and
// whether its last install failed.
type desktopUpdateJSON struct {
	Version    string `json:"version"`
	Notes      string `json:"notes"`
	Installing bool   `json:"installing"`
	Percent    int    `json:"percent"`
	Failed     bool   `json:"failed"`
}

// serverActions answers how the server's actions stand.
func (h *handler) serverActions(w http.ResponseWriter, _ *http.Request) {
	state := serverJSON{Version: h.Version, StartedAt: h.Health.Started, Restart: h.restartKind(), System: system(h.InContainer),
		UpdateTool: h.UpdateTool != nil && h.Accounts.Settings().UpdateToolURL != ""}
	if offered := h.desktopOffer(); offered != nil {
		state.DesktopUpdate = &desktopUpdateJSON{Version: offered.Version, Notes: offered.Notes, Installing: offered.Installing,
			Percent: offered.Percent, Failed: offered.Failed}
	}
	writeJSON(w, http.StatusOK, state)
}

// restartKind tells what starts Mira again once it exits.
func (h *handler) restartKind() string {
	switch {
	case h.Desktop:
		return restartDesktop
	case h.Accounts.Settings().RestartConfirmed:
		return restartConfirmed
	}
	return restartUnknown
}

// desktopOffer is the update Mira Desktop offers, newer than this
// server; nil when it offers none, or Mira Desktop does not run
// Mira.
func (h *handler) desktopOffer() *desktop.Offered {
	if !h.Desktop {
		return nil
	}
	offered, err := desktop.ReadOffer(desktop.Folders{Data: h.DataDir})
	if err != nil {
		h.Logger.Warn("The update Mira Desktop offers could not be read", "error", err)
		return nil
	}
	if offered == nil || !updates.Newer(offered.Version, h.Version) {
		return nil
	}
	return offered
}

// desktopRolledBack is Mira Desktop's return to the version that runs,
// from a newer one that did not start; nil when there was none, or
// Mira Desktop does not run Mira.
func (h *handler) desktopRolledBack() *desktop.RolledBack {
	if !h.Desktop {
		return nil
	}
	back, err := desktop.ReadRolledBack(desktop.Folders{Data: h.DataDir})
	if err != nil {
		h.Logger.Warn("Mira Desktop's return to its previous version could not be read", "error", err)
		return nil
	}
	if back == nil || !updates.Newer(back.Version, h.Version) {
		return nil
	}
	return back
}

// serverAction checks what each of the server's actions needs once its
// own conditions hold: the administrator's password or two-step code
// (see checkIdentity), then that the action was not asked within the last
// minute, 429 too_soon with Retry-After otherwise.
func (h *handler) serverAction(w http.ResponseWriter, r *http.Request, action string, given identity) bool {
	if !h.checkIdentity(w, r, sessionFrom(r.Context()).User, given) {
		return false
	}
	if allowed, wait := h.actions.allow(action, h.now()); !allowed {
		w.Header().Set("Retry-After", strconv.Itoa(int(wait.Seconds())+1))
		writeError(w, http.StatusTooManyRequests, "too_soon")
		return false
	}
	return true
}

// restartServer stops Mira, as a signal does, for what runs it to
// start it again: 202 with when this Mira started, which changes once
// the new one answers. While Mira knows nothing that starts it again,
// the request must say anyway: 409 restart_unconfirmed otherwise.
func (h *handler) restartServer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		identity
		Anyway bool `json:"anyway"`
	}
	if !decode(w, r, &body) {
		return
	}
	if h.Restart == nil {
		writeError(w, http.StatusNotFound, "not_found")
		return
	}
	if h.restartKind() == restartUnknown && !body.Anyway {
		writeError(w, http.StatusConflict, "restart_unconfirmed")
		return
	}
	if !h.serverAction(w, r, "restart", body.identity) {
		return
	}
	name := sessionFrom(r.Context()).User.Name
	h.Logger.Info("Mira restarts, as an administrator asked", "administrator", name, "started_again_by", h.restartKind())
	h.Activity.ServerRestarted(r.Context(), name)
	writeJSON(w, http.StatusAccepted, map[string]time.Time{"startedAt": h.Health.Started})
	// Mira waits for this answer to be sent before it stops.
	h.Restart()
}

// installDesktopUpdate asks Mira Desktop to install the update it
// offers, version, as its tray's Install Mira <version> does: 202, then
// Mira Desktop downloads it, stops Mira, installs the new version
// and starts it. 404 not_desktop: Mira Desktop does not run Mira;
// 409 no_update: it offers no such update, or installs it already.
func (h *handler) installDesktopUpdate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		identity
		Version string `json:"version"`
	}
	if !decode(w, r, &body) {
		return
	}
	if !h.Desktop {
		writeError(w, http.StatusNotFound, "not_desktop")
		return
	}
	offered := h.desktopOffer()
	if offered == nil || offered.Version != body.Version || offered.Installing {
		writeError(w, http.StatusConflict, "no_update")
		return
	}
	if !h.serverAction(w, r, "desktop-update", body.identity) {
		return
	}
	err := desktop.RequestInstall(desktop.Folders{Data: h.DataDir}, offered.Version)
	if errors.Is(err, desktop.ErrNoApp) {
		writeError(w, http.StatusConflict, "no_update")
		return
	}
	if err != nil {
		h.internalError(w, r, err)
		return
	}
	name := sessionFrom(r.Context()).User.Name
	h.Logger.Info("Mira Desktop installs a new version, as an administrator asked", "administrator", name, "version", offered.Version)
	h.Activity.DesktopUpdateStarted(r.Context(), name, offered.Version)
	writeJSON(w, http.StatusAccepted, map[string]string{"version": offered.Version})
}

// askUpdateTool sends Update now's request to the update tool the
// settings give (see package updatetool), and answers what it said: 200
// with its status and the start of its body, whatever its status. 409
// no_update_tool: the settings give no tool. 502 update_tool_timeout: it
// did not answer in time; update_tool_unreachable: it could not be
// reached, the message telling why.
func (h *handler) askUpdateTool(w http.ResponseWriter, r *http.Request) {
	var body identity
	if !decode(w, r, &body) {
		return
	}
	settings := h.Accounts.Settings()
	if h.UpdateTool == nil || settings.UpdateToolURL == "" {
		writeError(w, http.StatusConflict, "no_update_tool")
		return
	}
	if !h.serverAction(w, r, "update-tool", body) {
		return
	}
	name := sessionFrom(r.Context()).User.Name
	h.Logger.Info("Asking the update tool to update now, as an administrator asked", "administrator", name)
	h.Activity.UpdateToolAsked(r.Context(), name)
	// The request goes on if the admin app leaves: the tool may be
	// updating already.
	answer, err := h.UpdateTool.Ask(context.WithoutCancel(r.Context()), settings.UpdateToolURL, settings.UpdateToolToken)
	switch {
	case errors.Is(err, updatetool.ErrTimeout):
		h.Logger.Warn("The update tool did not answer in time")
		writeError(w, http.StatusBadGateway, "update_tool_timeout")
	case err != nil:
		h.Logger.Warn("The update tool could not be reached", "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "update_tool_unreachable", "message": err.Error()})
	default:
		h.Logger.Info("The update tool answered", "status", answer.Status)
		writeJSON(w, http.StatusOK, map[string]any{"status": answer.Status, "body": answer.Body})
	}
}
