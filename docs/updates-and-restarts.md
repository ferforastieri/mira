# Updates and restarts

This page is for administrators. It covers how Mira tells you of a new version, and how you restart or update it from the admin app's **Home**.

## New versions

Mira tells administrators when a new version is out, while **Check for new versions** is on.

| Setting | Where | Default | What it does |
|---|---|---|---|
| **Check for new versions** | **Settings › General** | On | Asks GitHub once a day, and a minute after Mira starts, for Mira's latest release, and tells administrators about a newer one. Off, GitHub is never asked. |
| **Include pre-releases** | **Settings › General** | Off | The check, and Mira Desktop's updates, consider pre-releases too. Pre-releases are for trying what comes next, and may have bugs. |

When the release is newer than the running version, in semantic version order:

- **Home** and **System › Health** each show a tile under **To look at**, with the new version and **Release notes**, which opens them on GitHub;
- Jellyfin apps see `HasUpdateAvailable` true in `/System/Info`;
- the **New version** notification is sent once per version, to the server's targets and administrators' own (see [Notifications](notifications.md#new-versions)).

Mira keeps the version found and the version told in the database. After a restart, the notice shows at once, and the notification is not sent again. Once Mira runs that version, the notice goes. Mira Desktop offers the same version (see [Updates](desktop.md#updates)).

### The request to GitHub

This is the one request Mira makes to GitHub for itself:

- **Address:** `https://api.github.com/repos/ferforastieri/mira/releases/latest`, which leaves out pre-releases and drafts.
- **With Include pre-releases:** `https://api.github.com/repos/ferforastieri/mira/releases?per_page=20`, the 20 latest releases. Mira takes the newest release and the newest pre-release.
- **What it sends:** nothing but Mira's version, in its `User-Agent: Mira/<version>` header.
- **Route:** the proxy the environment sets (`HTTPS_PROXY`). It gives up after 10 seconds.

A failed request, or GitHub asking to wait (403 or 429), changes nothing. The request is tried again an hour later.

### Pre-releases

A pre-release is a release GitHub marks as one, or whose tag is one, such as `v1.4.0-rc.1`.

- **Include pre-releases off:** pre-releases are left out.
- **On:** the newest pre-release shows while it is newer than the newest release.
- **Turning it on:** looks at once, rather than at the next daily check.
- **Turning it off:** shows the newest release again at once.

### No check

- A development build, whose version is `dev`, asks nothing and never shows a new version.
- With the check off, Mira never contacts GitHub. The notice and the Health item go, and `HasUpdateAvailable` is false.

**Compared with Jellyfin**

- Jellyfin 12.4 has no check for its own new versions: `HasUpdateAvailable` of `/System/Info` is marked obsolete, "This should be handled by the package manager" ([SystemInfo.cs, lines 133 to 135](https://github.com/jellyfin/jellyfin/blob/v12.4/MediaBrowser.Model/System/SystemInfo.cs#L133-L135)), and never set, so it is always false ([SystemManager.cs, lines 56 to 83](https://github.com/jellyfin/jellyfin/blob/v12.4/Emby.Server.Implementations/SystemManager.cs#L56-L83)). Mira sets it from its own check, following releases alone unless **Include pre-releases** is on.

## Restart and updates

Administrators restart Mira, and start its update, from the foot of **Server state** on **Home**, without opening the host or Docker. **Search** finds both, as "Restart Mira" and "Update now". Each opens **Home** with the action's confirmation.

| Setting | Where | Default | What it does |
|---|---|---|---|
| **Mira is started again when it stops** | **Settings › General** | Off | Confirms that a restart policy or a service manager starts Mira again: **Restart** then warns no more. Not shown in Mira Desktop, which always does. |
| **Update tool address** | **Settings › General** | Empty | The update API **Update now** sends its request to. Empty hides **Update now**. |
| **Update tool token** | **Settings › General** | Empty | The bearer token sent with the request, when the tool asks for one. |

Each action:

- is an administrator's alone;
- is confirmed each time with the administrator's password. With two-step sign-in on, a code of their app or a recovery code confirms it. Wrong ones count as wrong passwords at sign-in do (see [Two-step sign-in](sign-in-and-security.md#two-step-sign-in));
- runs at most once a minute. Another request within the minute is refused, and so is a restart within a minute of Mira starting, so restarts cannot follow each other;
- is recorded in the activity log, with who asked, and shows in **Recent activity** under **Server and addons**.

Mira has no action that stops it for good:

- in a container with a restart policy, or in Mira Desktop, it would start again at once;
- a remote off switch is easy to misuse from a stolen session.

Mira is never given the Docker socket either: it is root on the host for whoever takes over a server.

### Restart

**Restart** stops Mira as a stop signal does:

1. Playbacks, conversions, recordings and tasks end.
2. Open connections are given up to 15 seconds.
3. Mira exits with code 75 (`EX_TEMPFAIL`), which tells that it should be started again.

Whatever runs Mira starts it again:

| Runs Mira | What starts it again |
|---|---|
| Docker | The container's restart policy: `unless-stopped` and `always`, and `on-failure` too, as the code is not 0 (see [Restart from the admin app](installation.md#restart-from-the-admin-app)) |
| systemd | `Restart=always` or `Restart=on-failure` |
| Mira Desktop | The app itself, as when the server ends by itself (see [Restart](desktop.md#restart)) |

The admin app then waits, and reloads once a new Mira answers. Apps reconnect to it as well. Sessions are kept in the database, so nobody signs in again.

Only the Docker socket tells a container's restart policy, not the container itself. So Mira cannot know whether it will be started again, unless:

- it runs in Mira Desktop;
- or an administrator turns on **Mira is started again when it stops**, in **Settings › General**.

Until then, **Restart** warns that Mira may not come back, and asks to choose **Restart anyway**. The API refuses a restart that does not say so. Without a restart policy, Mira stays stopped until it is started on the host. The admin app says so after 2 minutes.

<a id="mira-desktop-s-update"></a><a id="mira-desktops-update"></a>

### Mira Desktop update

In Mira Desktop, the tile of a new version on **Home** offers **Install Mira `<version>`**. It shows while the app's menu offers it: a release newer than the server, with a package for the system (see [Updates](desktop.md#updates)). Elsewhere than in Mira Desktop, it is never offered.

It asks the app to do what its menu item does:

1. Download the package, and check it against the release's signed manifest.
2. Stop Mira, install the new version and start it.
3. Go back to the previous version when the new one does not start (see [Going back to the previous version](desktop.md#going-back-to-the-previous-version)).

The notice shows the download's progress, then the admin app waits for the new version to answer. When the install fails, Mira runs on as it was, and the admin app says so.

### Update now

A container cannot update its own image without the Docker socket. An update tool you run beside Mira can: many offer an HTTP API that updates the containers they watch. Give its address in **Update tool address**, in **Settings › General**. **Update now** then shows at the foot of **Server state** on **Home**, and in the tile of a new version when one is out.

Mira sends this request, and nothing else:

- method `POST`, with no body;
- to the address exactly as given, path and query included, such as `http://updater:8080/v1/update`. Match it to your tool's update endpoint;
- with the header `Authorization: Bearer <token>` when a token is set, and `User-Agent: Mira/<version>`;
- giving up after 10 seconds to connect, and 50 seconds for the answer, and following no redirect.

The admin app shows what the tool answered, whatever the status: its HTTP status and the start of its text. It also says when the tool could not be reached or did not answer in time. A tool that answers only once it has updated may replace Mira's container first. The admin app then waits for Mira to answer again, and reloads.

The token is stored encrypted, as the other secrets. It is never sent back to the admin app, and is shown again only on demand (see [Showing a saved key](administration.md#showing-a-saved-key)).

<a id="compared-with-jellyfin"></a>

**Compared with Jellyfin**

- Jellyfin's dashboard has "Restart" and "Shut down" buttons, through `POST /System/Restart`, allowed to administrators and to any user on the local network ([SystemController.cs](https://github.com/jellyfin/jellyfin/blob/v12.4/Jellyfin.Api/Controllers/SystemController.cs#L114-L122)), and `POST /System/Shutdown`, to administrators ([SystemController.cs](https://github.com/jellyfin/jellyfin/blob/v12.4/Jellyfin.Api/Controllers/SystemController.cs#L130-L138)).
- Jellyfin restarts within its own process: it stops its host and starts a new one in a loop ([SystemManager.cs](https://github.com/jellyfin/jellyfin/blob/v12.4/Emby.Server.Implementations/SystemManager.cs#L126-L139), [Program.cs](https://github.com/jellyfin/jellyfin/blob/v12.4/Jellyfin.Server/Program.cs#L145-L163)).
- Jellyfin's dashboard puts **Restart**, in red, with the server's version at the top of its first widget ([ServerInfoWidget.tsx](https://github.com/jellyfin/jellyfin-web/blob/v12.4/src/apps/dashboard/components/widgets/ServerInfoWidget.tsx#L87-L96)). **Home** puts **Restart** with the version, at the foot of **Server state**.
- Mira differs on purpose:
  - it exits and lets what runs it start it again, which also recovers a server stuck in a way a restart within the process would not;
  - it has no shutdown;
  - its actions are the admin app's, for administrators only, never for a local address, and each asks for the password or a two-step code again;
  - `/System/Restart` and `/System/Shutdown` are not offered to Jellyfin apps.
