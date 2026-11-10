# Reverse proxy

This page is for administrators who serve Mira over HTTPS through a reverse proxy, such as Traefik, Caddy or nginx, at an address like `https://media.example.org`. It covers the headers to pass, WebSockets, HTTPS, **Known proxies** and the public address.

Mira listens on plain HTTP, on port 8096 by default (see [Configuration](configuration.md)).

- The proxy terminates HTTPS and forwards each request to that port.
- Apps and browsers on the local network can keep connecting to port 8096 directly.

Mira can also serve HTTPS itself, on port 8920 by default, with no proxy (see [HTTPS](https-and-networking.md#https)). It takes certificate files or a certificate from Let's Encrypt, and needs one port forwarded on the router. A proxy is still the way when:

- Mira shares port 443 with other services under several names;
- the certificate is a wildcard, or for names that resolve only on the local network, and another tool should not write its files where Mira reads them;
- the proxy also adds authentication, rate limits or a CDN in front.

## Headers to pass

The proxy passes these headers on to Mira:

| Header | What Mira does with it |
|---|---|
| `Host`, or `X-Forwarded-Host` | Starts the stream addresses Mira gives apps, so that they use the address the app reached. Keep the original `Host`, or pass it as `X-Forwarded-Host`. |
| `X-Forwarded-Proto` | `https` makes those stream addresses start with `https://`, and marks the admin app's session cookie secure. |
| `X-Forwarded-For` | Names the client, when the request comes from a known proxy (see [Known proxies](#known-proxies)). |
| `X-Real-IP` | Names the client when `X-Forwarded-For` is missing, from a known proxy too. |

By default, Traefik and Caddy keep `Host`, and send `X-Forwarded-For`, `X-Forwarded-Proto` and `X-Forwarded-Host`.

- Traefik also sends `X-Real-Ip`.
- With nginx, set them with `proxy_set_header`.

## WebSockets

Jellyfin apps, the web player among them, keep a live connection open at `/socket`. It tells them of changes, and carries remote control and Watch together (SyncPlay in Jellyfin).

The proxy must pass the WebSocket upgrade, with the `Upgrade` and `Connection` headers:

- Traefik and Caddy do so by default.
- nginx needs `proxy_http_version 1.1` and the two headers set.
- nginx also closes a connection idle for longer than its `proxy_read_timeout`.

The admin app uses no WebSocket: it asks the server again every few seconds.

## Slow answers

A torrent version's first play over a slow swarm can keep the app's PlaybackInfo request waiting for up to 3 minutes (see [Torrents](torrents.md#deadlines)). The request sends nothing until it answers.

That wait starts once the versions are listed, and the request can last longer from its start. With the default settings, it lasts about 5 minutes at most:

- the addons' answers listing the versions, 30 seconds at most each, and 5 seconds at most after one gave a version to play (see [How Play waits for addons](versions.md#how-play-waits-for-addons));
- the versions of the list kept, analyzed, 20 seconds (**Maximum time to analyze a version**) for each 3 of them;
- when none of those plays, the addons asked again, then the torrents' 3 minutes, and a version analyzed meanwhile.

A larger **Maximum time to analyze a version**, or more **Versions tried when one does not work**, makes it longer.

| Proxy | How long it waits for an answer |
|---|---|
| Traefik, Caddy | As long as it takes, by default |
| nginx | 60 seconds without a byte (`proxy_read_timeout`), then it closes the response |

With nginx, set `proxy_read_timeout` to `360s` or more for Mira, or such a play fails at the proxy. Even without torrents, a play whose addons answer late can take more than its 60 seconds. Raise it further with those two settings.

## HTTPS and the public address

Get the certificate on the proxy, and have it redirect HTTP to HTTPS. Leave **Serve HTTPS**, in **Settings › Security**, off: the proxy forwards to Mira's HTTP port.

If Mira's own HTTPS is on behind the proxy anyway, with **Redirect HTTP to HTTPS**, requests are redirected as follows:

| Request | Redirected |
|---|---|
| From a known proxy that sends `X-Forwarded-Proto: https` | Never: the proxy received it over HTTPS. |
| From a known proxy without that header, for a client on the internet | Yes, which may send the client where it cannot go (see below). |
| From a proxy on the local network missing from **Known proxies** | Never: every client looks local. |

A redirected request goes to `https://` and the host it asked for, at one of these ports:

- the port of the `https://` **Public address**, or 443 when it names none;
- without an `https://` **Public address**, Mira's HTTPS port, which clients on the internet may not reach.

Have the proxy pass `X-Forwarded-Proto` to avoid this.

Then set **Public address**, in **Settings › Notifications**, to the address people open Mira at, such as `https://media.example.org`. Mira drops a trailing slash when it saves the address.

- Links in notifications and invite links start with it.
- Apps do not need it: stream addresses come from the headers above.

## Known proxies

Without a known proxy, Mira takes each request to come from its connection's address. Behind a proxy, that is the proxy's address for everyone, so:

- failed sign-ins count for everyone together. Ten failures in 15 minutes, from anyone, refuse every user's sign-in for up to 15 minutes, in apps, the admin app, invite links and password reset PINs;
- **Now playing**, on **Home**, and each user's devices show the proxy's address;
- when the proxy is on the local network, every client counts as on it. `/System/Endpoint` tells apps so, and a forgotten password asked from the internet makes a PIN.

List the proxy under **Known proxies**, in **Settings › Security**, one entry per line, up to 64 entries:

| Entry | Example |
|---|---|
| An IPv4 or IPv6 address | `172.30.0.2` |
| A network | `172.30.0.0/24` |
| A host name | `traefik` (see [Proxies by name](#proxies-by-name)) |

A request from a known proxy is then taken to come from the client the proxy names:

- in `X-Forwarded-For`, the rightmost address that is not a known proxy, read across every line of the header;
- without that header, the address in `X-Real-IP`.

Requests from any other address keep their connection's address. Their `X-Forwarded-For` and `X-Real-IP` are ignored, since anyone can send them.

List only proxies: a client connecting from a known address could name any address it likes.

DLNA's addresses under `/dlna/` serve the local network only (see [DLNA](dlna.md)). They refuse:

- any request that carries `X-Forwarded-For`, `X-Real-IP` or `Forwarded` from an address that is not a known proxy;
- any request that names its client in `Forwarded` alone, even from a known proxy, since Mira reads only `X-Forwarded-For` and `X-Real-IP`.

**System › Health** warns when an address on the local network sends `X-Forwarded-For` or `X-Real-IP` without being a known proxy. That is how a proxy missing from the list shows up.

- The warning names that address, for an hour after its last such request.
- Addresses on the internet are never named, since a client can send the headers itself.

### Proxies by name

A host name, such as a container's name on a Docker network, is looked up with the system's resolver:

- when the settings are saved;
- when Mira starts;
- then every minute.

Every address the name resolves to is a known proxy, until a lookup no longer returns it.

- When the proxy gets another address, the old one stops being trusted at the next lookup, within a minute.
- A name removed from the list stops being trusted at once.
- A name that resolves to nothing, or whose lookup fails, trusts nothing.

While names are listed, an address on the local network that is not a known proxy and sends `X-Forwarded-For` or `X-Real-IP` sets off a new lookup at once. This happens at most once every 10 seconds.

- A proxy that starts after Mira, as after a reboot, or comes back with another address, is thus trusted within seconds rather than a minute.
- The lookup runs in the background: the request that sets it off still counts as coming from its connection's address.
- The next requests count as coming from the client the proxy names, once the proxy's address is found.

**System › Health** shows what each name resolves to, under **Known proxies**, and warns of a name that resolves to nothing.

A name is only as trusted as the DNS that answers for it: whoever controls that answer chooses the addresses Mira trusts.

- List only names your own network resolves, such as the containers' names Docker's own DNS answers on a network you made.
- Never list a name a public DNS server answers.

Names are written in letters, digits, hyphens, underscores and dots. Mira keeps them in lower case.

API details: [Known proxies in the admin API](jellyfin-api.md#known-proxies-in-the-admin-api).

### A proxy in a Docker network

When the proxy and Mira are containers on the same Docker network, Mira sees the proxy's address on that network, such as `172.30.0.2`, not the host's.

- A container's address can change when the containers start in another order. List the proxy's container name instead (see [the example](#example-with-traefik-and-docker-labels)).
- Docker's own DNS answers that name on a network the containers share, one you declare rather than the default `bridge` network.
- Listing the network's subnet, fixed in Compose, works too, but then every container on that network is trusted.

The proxy may run on the host, or use the host's network, and reach Mira through its published port. Mira then usually sees the address of the Docker network's gateway, such as `172.18.0.1`. **System › Health** names the address it sees.

### A CDN or a second proxy in front

With a CDN or a second proxy in front, the proxy next to Mira receives the requests from that one, and adds its address to `X-Forwarded-For`. Mira then takes that address for the client, and every client counts as the CDN or the second proxy again.

| In front | What to do |
|---|---|
| A second proxy you run | Add its address to **Known proxies** too. Have the proxy next to Mira keep the `X-Forwarded-For` it receives from it. |
| A CDN | Have the proxy next to Mira trust the CDN's published address ranges, so that it keeps the client the CDN names and drops the header from anyone else. Listing the CDN's ranges under **Known proxies** works too, while they fit in 64 entries. |

<a id="example-traefik-with-docker-labels"></a>

## Example with Traefik and Docker labels

This adds Traefik to the `compose.yaml` of [Installation](installation.md), with certificates from Let's Encrypt.

- Traefik's container is named `traefik`.
- Mira joins Traefik's `proxy` network, and keeps its port for the local network.

```yaml
services:
  traefik:
    image: traefik:v3.5
    container_name: traefik
    command:
      - --providers.docker=true
      - --providers.docker.exposedbydefault=false
      - --providers.docker.network=mira_proxy
      - --entrypoints.web.address=:80
      - --entrypoints.web.http.redirections.entrypoint.to=websecure
      - --entrypoints.web.http.redirections.entrypoint.scheme=https
      - --entrypoints.websecure.address=:443
      - --certificatesresolvers.le.acme.tlschallenge=true
      - --certificatesresolvers.le.acme.email=you@example.org
      - --certificatesresolvers.le.acme.storage=/letsencrypt/acme.json
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - letsencrypt:/letsencrypt
    networks:
      - proxy
    restart: unless-stopped

  mira:
    # As in compose.yaml, plus:
    labels:
      - traefik.enable=true
      - traefik.http.routers.mira.rule=Host(`media.example.org`)
      - traefik.http.routers.mira.entrypoints=websecure
      - traefik.http.routers.mira.tls.certresolver=le
      - traefik.http.services.mira.loadbalancer.server.port=8096
    networks:
      - default
      - proxy

networks:
  proxy:

volumes:
  letsencrypt:
```

Compose names the `proxy` network `mira_proxy`, after the project's name, `mira`. Point `media.example.org` at the server, open ports 80 and 443, and start the stack. Then:

1. In **Settings › Security**, add `traefik` to **Known proxies**. **System › Health** then shows the address it resolves to on the `proxy` network.
2. In **Settings › Notifications**, set **Public address** to `https://media.example.org`.

Traefik passes WebSockets, and sets `X-Forwarded-For`, `X-Forwarded-Proto`, `X-Forwarded-Host` and `X-Real-Ip` by itself.

In apps, the server's address is `https://media.example.org`.

- Mira also answers apps under `https://media.example.org/jellyfin`, as at the root. A proxy that already serves Jellyfin at `/jellyfin` can forward that path to Mira as it is (see [Under /jellyfin](#under-jellyfin)).
- The admin app stays at the root only.

Traefik routes to Mira only once its container is healthy, which it becomes within seconds of starting (see [Behind Traefik](installation.md#behind-traefik)).

## Under /jellyfin

Mira answers apps under `/jellyfin` as at the root, as a Jellyfin server whose base URL is `/jellyfin` does. Some apps given an address without a path try it there first (see [Compatible apps](jellyfin-compatibility.md#compatible-apps)).

A proxy can forward only `/jellyfin` and what is under it to Mira, the path kept, and serve something else at the root:

- The addresses Mira gives an app that came under `/jellyfin` keep it: the stream addresses of its media sources, its live TV playlists' entries, the images it lists for editing, and `LocalAddress` in `/System/Info/Public`. Their requests reach the proxy under `/jellyfin` too.
- Its other addresses, such as `TranscodingUrl`, subtitle addresses and the entries of its HLS playlists, are relative to the server's address or to the playlist, which already ends in `/jellyfin`.
- Mira reads the base from the path only. It ignores `X-Forwarded-Prefix`: a proxy that removes `/jellyfin` before forwarding makes Mira answer as at the root, with addresses without it.
- The admin app, `/healthz`, `/readyz` and `/metrics` stay at the root, so they are out of reach through such a proxy.

**Compared with Jellyfin**

- Jellyfin with the base URL `/jellyfin` answers only under it, and redirects other paths to its web client there ([Startup.cs, lines 173 to 177](https://github.com/jellyfin/jellyfin/blob/v12.4/Jellyfin.Server/Startup.cs#L173-L177), [BaseUrlRedirectionMiddleware.cs, lines 43 to 76](https://github.com/jellyfin/jellyfin/blob/v12.4/Jellyfin.Api/Middleware/BaseUrlRedirectionMiddleware.cs#L43-L76)). Mira answers under `/jellyfin` and at the root alike, with no setting.
- Jellyfin's absolute addresses carry its base URL ([ApplicationHost.cs, lines 988 to 1008](https://github.com/jellyfin/jellyfin/blob/v12.4/Emby.Server.Implementations/ApplicationHost.cs#L988-L1008)): `LocalAddress` ([SystemManager.cs, line 77](https://github.com/jellyfin/jellyfin/blob/v12.4/Emby.Server.Implementations/SystemManager.cs#L77)) and the addresses of its own live streams ([SharedHttpStream.cs, line 66](https://github.com/jellyfin/jellyfin/blob/v12.4/src/Jellyfin.LiveTv/TunerHosts/SharedHttpStream.cs#L66)). Mira's carry the base the request came under.
- Jellyfin's `TranscodingUrl` is relative ([MediaInfoHelper.cs, lines 310 to 326](https://github.com/jellyfin/jellyfin/blob/v12.4/Jellyfin.Api/Helpers/MediaInfoHelper.cs#L310-L326), [StreamInfo.cs, lines 877 to 898](https://github.com/jellyfin/jellyfin/blob/v12.4/MediaBrowser.Model/Dlna/StreamInfo.cs#L877-L898)), and so are its master playlist's entries ([DynamicHlsHelper.cs, line 188](https://github.com/jellyfin/jellyfin/blob/v12.4/Jellyfin.Api/Helpers/DynamicHlsHelper.cs#L188)), as Mira's.
- Jellyfin reads no `X-Forwarded-Prefix` either: from a known proxy, it takes `X-Forwarded-For`, `X-Forwarded-Proto` and `X-Forwarded-Host` only ([ApiServiceCollectionExtensions.cs, lines 166 to 188](https://github.com/jellyfin/jellyfin/blob/v12.4/Jellyfin.Server/Extensions/ApiServiceCollectionExtensions.cs#L166-L188)).
