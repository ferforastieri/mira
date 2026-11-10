<script setup lang="ts">
import { withBase } from 'vitepress'
import { onBeforeUnmount, onMounted, ref, type Directive } from 'vue'
import FeatureViewer from './FeatureViewer.vue'
import Icon from './Icon.vue'
import TourScreen from './TourScreen.vue'
import { apps, extras, repository } from './home'
import { icons } from './icons'

/** The README's quick start. */
const commands = [
  'git clone https://github.com/ferforastieri/mira.git',
  'cd mira',
  'cp .env.example .env',
  '# set POSTGRES_PASSWORD in .env: openssl rand -hex 24 makes one',
  'docker compose up -d',
]

/** The shelf's two rows of poster art, each written twice so that it loops. */
const shelf = [
  [
    'amber-road',
    'northwave',
    'ocean-life',
    'echo-valley',
    'paper-moons',
    'glass-garden',
    'night-train',
    'silver-coast',
  ],
  [
    'low-tide',
    'long-field',
    'paper-moons-sq',
    'amber-road-sq',
    'northwave',
    'echo-valley',
    'ocean-life',
    'glass-garden',
    'silver-coast',
  ],
]

const heroPosters = shelf[0].slice(0, 6)

/** Image cards first, then the text cards, so each grid row has a similar height. */
const extraCards = [0, 1, 7, 3, 4, 5, 2, 6].map((index) => extras[index])
const tour = ref<InstanceType<typeof TourScreen>>()
const copied = ref(false)
type Theme = 'light' | 'dark'
const theme = ref<Theme>('light')
let copiedTimer: ReturnType<typeof setTimeout> | undefined

function applyTheme(next: Theme) {
  theme.value = next
  document.documentElement.style.colorScheme = next
  try {
    localStorage.setItem('mira-theme', next)
  } catch {
    // A browser that refuses storage keeps the theme for this page only.
  }
}

function toggleTheme() {
  applyTheme(theme.value === 'light' ? 'dark' : 'light')
}

async function copyCommands() {
  await navigator.clipboard.writeText(commands.join('\n'))
  copied.value = true
  clearTimeout(copiedTimer)
  copiedTimer = setTimeout(() => (copied.value = false), 2000)
}

/** Sections below the hero rise into place as they reach the screen. */
const revealer =
  typeof IntersectionObserver === 'undefined'
    ? undefined
    : new IntersectionObserver(
        (entries) => {
          for (const entry of entries) {
            if (!entry.isIntersecting) continue
            entry.target.classList.add('in')
            revealer?.unobserve(entry.target)
          }
        },
        { rootMargin: '0px 0px -6% 0px' },
      )

const vReveal: Directive<HTMLElement> = {
  mounted(element) {
    element.classList.add('rv')
    revealer?.observe(element)
  },
}

onMounted(() => {
  let saved: string | null = null
  try {
    saved = localStorage.getItem('mira-theme')
  } catch {
    // No storage: follow the system.
  }
  const system = matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
  applyTheme(saved === 'light' || saved === 'dark' ? saved : system)
  document.body.classList.add('mira-landing')
})
onBeforeUnmount(() => {
  document.body.classList.remove('mira-landing')
  revealer?.disconnect()
  clearTimeout(copiedTimer)
})
</script>

<template>
  <div class="pf-home" :data-theme="theme">
    <header class="m-nav">
      <div class="wrap m-nav-inner">
        <a class="m-brand" :href="withBase('/')">
          <img :src="withBase('/mira.svg')" alt="" width="42" height="42" />
          <span>Mira</span>
        </a>
        <nav class="m-links" aria-label="Main">
          <a href="#how">How it works</a>
          <a href="#features">Features</a>
          <a href="#built-in">Extras</a>
          <a href="#install">Install</a>
          <a :href="withBase('/docs/')">Docs</a>
        </nav>
        <div class="m-nav-actions">
          <button
            type="button"
            class="m-theme"
            :aria-label="theme === 'light' ? 'Switch to dark mode' : 'Switch to light mode'"
            :title="theme === 'light' ? 'Dark mode' : 'Light mode'"
            @click="toggleTheme"
          >
            <svg
              v-if="theme === 'light'"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.6"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8Z" />
            </svg>
            <svg
              v-else
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.6"
              stroke-linecap="round"
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <circle cx="12" cy="12" r="4" />
              <path
                d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"
              />
            </svg>
          </button>
          <a class="m-github" :href="repository" target="_blank" rel="noreferrer">GitHub</a>
          <a class="m-docs-mobile" :href="withBase('/docs/')">Docs</a>
          <a class="m-cta" href="#install">Get started</a>
        </div>
      </div>
    </header>

    <section id="top" class="hero" aria-labelledby="hero-title">
      <div class="wrap hero-grid">
        <div class="hero-copy">
          <p class="kicker">
            <span class="kicker-dot" aria-hidden="true"></span>Self-hosted · Jellyfin compatible
          </p>
          <h1 id="hero-title">Your media,<br /><span class="grad">everywhere.</span></h1>
          <p class="lead">
            Mira turns Stremio addons, music addons, IPTV and your own files into one library that
            every Jellyfin app can play — with real accounts, transcoding and no Jellyfin server.
          </p>
          <div class="actions">
            <a class="btn btn-ink" href="#install"
              >Install Mira<span class="btn-knob"><Icon :svg="icons.arrowRight" :size="15" /></span
            ></a>
            <button type="button" class="btn btn-line" @click="tour?.open()">
              <span class="play-dot" aria-hidden="true"></span>Watch the tour
              <span class="mono">1:23</span>
            </button>
          </div>
          <ul class="hero-facts">
            <li><b>Free</b><span>MIT · self-hosted</span></li>
            <li><b>One command</b><span>Docker Compose</span></li>
            <li><b>Any app</b><span>Jellyfin API</span></li>
          </ul>
        </div>

        <div class="hero-art" aria-hidden="true">
          <span class="art-glow"></span>
          <div class="device">
            <div class="device-bar">
              <span class="device-dot"></span><span class="device-name">Mira</span>
              <span class="device-pill">Library</span>
            </div>
            <div class="device-grid">
              <img
                v-for="poster in heroPosters"
                :key="poster"
                :src="withBase(`/posters/${poster}.jpg`)"
                alt=""
              />
            </div>
            <div class="device-foot">
              <span>Continue watching</span>
              <i><b></b></i>
            </div>
          </div>
          <img class="float-mark" :src="withBase('/mira.svg')" alt="" width="96" height="96" />
          <span class="chip chip-a">Addons</span>
          <span class="chip chip-b">Live TV</span>
          <span class="chip chip-c">Transcoding</span>
        </div>
      </div>
    </section>

    <section class="works" aria-label="Works with">
      <div class="wrap works-line">
        <span class="works-label">Works with</span>
        <div class="marquee">
          <div class="marquee-track">
            <span v-for="(app, index) in [...apps, ...apps]" :key="index">{{ app }}</span>
          </div>
        </div>
        <a class="works-link" :href="withBase('/docs/jellyfin-compatibility')"
          >Compatibility <Icon :svg="icons.arrowUpRight" :size="14"
        /></a>
      </div>
    </section>

    <span id="how" class="jump-anchor" aria-hidden="true"></span>
    <section class="section m-ink" aria-labelledby="how-title">
      <div class="wrap">
        <div v-reveal class="head">
          <p class="kicker light">How it works</p>
          <h2 id="how-title">One server between<br />your sources and your screens.</h2>
          <p class="head-lead">
            Jellyfin apps can't read addons. Mira speaks the Jellyfin API on one side and your
            sources on the other, then gets out of the way.
          </p>
        </div>
        <div class="pipeline">
          <article v-reveal class="pipe-card">
            <span class="pipe-index">01 / In</span>
            <h3>Your sources</h3>
            <ul>
              <li>Stremio addons</li>
              <li>Music addons</li>
              <li>IPTV playlists and Xtream</li>
              <li>Local folders and network shares</li>
            </ul>
          </article>
          <div v-reveal class="pipe-hub">
            <span class="pipe-line" aria-hidden="true"></span>
            <img :src="withBase('/mira.svg')" alt="" width="76" height="76" />
            <b>Mira</b>
            <small>Jellyfin API</small>
          </div>
          <article v-reveal class="pipe-card">
            <span class="pipe-index">02 / Out</span>
            <h3>Any Jellyfin app</h3>
            <ul>
              <li>Phones, tablets and TVs</li>
              <li>The built-in web client</li>
              <li>DLNA players</li>
              <li>Signed in as they would to Jellyfin</li>
            </ul>
          </article>
        </div>
        <div class="pipe-foot">
          <span>Catalogs become libraries.</span>
          <span>Streams become versions.</span>
          <span>IPTV becomes Live TV.</span>
        </div>
      </div>
    </section>

    <span id="tour" class="jump-anchor" aria-hidden="true"></span>
    <section class="section tour" aria-labelledby="tour-title">
      <div class="wrap">
        <div v-reveal class="split-head">
          <div>
            <p class="kicker">The tour</p>
            <h2 id="tour-title">See it in the apps you already use.</h2>
          </div>
          <p class="head-lead">
            Phones, TVs and the web, one server behind them. A minute and a half, start to finish.
          </p>
        </div>
        <TourScreen ref="tour" v-reveal />
      </div>
    </section>

    <span id="features" class="jump-anchor" aria-hidden="true"></span>
    <section class="section features" aria-labelledby="features-title">
      <div class="wrap">
        <div v-reveal class="split-head">
          <div>
            <p class="kicker">Now showing</p>
            <h2 id="features-title">Everything your apps expect.</h2>
          </div>
          <p class="head-lead">
            Libraries, accounts, transcoding, Live TV and music, all through the Jellyfin API. Pick
            one, or let them play in turn.
          </p>
        </div>
        <FeatureViewer v-reveal />
      </div>
    </section>

    <section class="shelf" aria-label="Artwork">
      <div class="rows">
        <div v-for="(row, index) in shelf" :key="index" class="row">
          <div class="track">
            <img
              v-for="(poster, at) in [...row, ...row]"
              :key="at"
              :src="withBase(`/posters/${poster}.jpg`)"
              alt=""
              width="164"
              height="219"
              loading="lazy"
            />
          </div>
        </div>
      </div>
    </section>

    <span id="built-in" class="jump-anchor" aria-hidden="true"></span>
    <section class="section extras" aria-labelledby="built-title">
      <div class="wrap">
        <div v-reveal class="split-head">
          <div>
            <p class="kicker">Also built in</p>
            <h2 id="built-title">The parts a home server needs.</h2>
          </div>
          <p class="head-lead">
            The things that matter once more than one person watches. Each one has its page in the
            docs.
          </p>
        </div>
        <ul class="extra-grid">
          <li v-for="extra in extraCards" :key="extra.name" v-reveal>
            <a :href="withBase(extra.link)">
              <span v-if="extra.still" class="extra-shot">
                <img
                  :src="withBase(extra.still.src)"
                  :alt="extra.still.alt"
                  loading="lazy"
                  width="1200"
                  height="672"
                />
              </span>
              <span class="extra-body">
                <span class="extra-icon" :class="{ cyan: extra.facet === 'cyan' }"
                  ><Icon :svg="extra.icon" :size="20"
                /></span>
                <span class="extra-copy">
                  <b>{{ extra.name }}</b>
                  <span>{{ extra.text }}</span>
                </span>
                <Icon class="extra-go" :svg="icons.arrowUpRight" :size="17" />
              </span>
              <span v-if="extra.paths" class="extra-paths">
                <span v-for="path in extra.paths" :key="path.path"
                  >server:8096<em>{{ path.path }}</em></span
                >
              </span>
            </a>
          </li>
        </ul>
      </div>
    </section>

    <span id="install" class="jump-anchor" aria-hidden="true"></span>
    <section class="section m-ink install" aria-labelledby="install-title">
      <div class="wrap install-grid">
        <div v-reveal>
          <p class="kicker light">Install</p>
          <h2 id="install-title">Up and running in one command.</h2>
          <p class="head-lead">
            Docker with Compose v2 is all it needs. The image is published for linux/amd64 and
            linux/arm64.
          </p>
          <ol class="steps">
            <li>
              <span class="n">1</span
              ><span>Open <code>http://&lt;server&gt;:8096/admin/</code>.</span>
            </li>
            <li>
              <span class="n">2</span
              ><span
                >Enter the one-time setup code from <code>docker compose logs mira</code> to create
                the administrator.</span
              >
            </li>
            <li>
              <span class="n">3</span
              ><span
                >Continue with
                <a class="underline" :href="withBase('/docs/getting-started')">Getting started</a
                >.</span
              >
            </li>
          </ol>
          <p class="more">
            Unraid, GPUs, pinned versions and builds from source are in the
            <a class="underline" :href="withBase('/docs/installation')">installation guide</a>.
          </p>
        </div>

        <div v-reveal class="install-side">
          <div class="terminal">
            <div class="terminal-bar">
              <span class="terminal-dots" aria-hidden="true"><i></i><i></i><i></i></span>
              <span class="terminal-title">Shell</span>
              <button
                type="button"
                class="copy-button"
                :class="{ done: copied }"
                @click="copyCommands"
              >
                <Icon :svg="copied ? icons.check : icons.copy" :size="14" />
                {{ copied ? 'Copied' : 'Copy' }}
              </button>
            </div>
            <pre><code><template v-for="line in commands" :key="line"><span :class="line.startsWith('#') ? 'comment' : 'command'">{{ line }}</span>{{ '\n' }}</template></code></pre>
          </div>

          <a class="desktop" :href="withBase('/docs/desktop')">
            <span class="desktop-mark"><img :src="withBase('/mira.svg')" alt="" /></span>
            <span class="desktop-text">
              <small>Rather not run Docker?</small>
              <b>Mira Desktop</b>
              <span>The same server on Windows, macOS or Linux, from the tray.</span>
            </span>
            <span class="systems" aria-hidden="true">
              <Icon :svg="icons.windowsLogo" :size="18" />
              <Icon :svg="icons.appleLogo" :size="18" />
              <Icon :svg="icons.linuxLogo" :size="18" />
            </span>
            <Icon class="desktop-go" :svg="icons.arrowUpRight" :size="17" />
          </a>
        </div>
      </div>

      <div class="wrap">
        <p v-reveal class="note">
          <Icon :svg="icons.info" :size="18" />
          <span
            >Mira is in early development. Expect changes between releases, and report problems in
            the <a class="underline" :href="`${repository}/issues`">issues</a>.</span
          >
        </p>
      </div>
    </section>

    <footer class="footer">
      <div class="wrap">
        <div class="footer-top">
          <div>
            <a class="m-brand" :href="withBase('/')">
              <img :src="withBase('/mira.svg')" alt="" width="34" height="34" /><span>Mira</span>
            </a>
            <p>Your media, in every Jellyfin app.</p>
          </div>
          <nav class="footer-links" aria-label="Footer">
            <a :href="withBase('/docs/')">Docs</a>
            <a :href="repository" target="_blank" rel="noreferrer">GitHub</a>
            <a :href="`${repository}/releases`" target="_blank" rel="noreferrer">Releases</a>
            <a :href="`${repository}/discussions`" target="_blank" rel="noreferrer">Community</a>
            <a :href="`${repository}/blob/main/SECURITY.md`" target="_blank" rel="noreferrer"
              >Security</a
            >
          </nav>
        </div>
        <p class="footer-word" aria-hidden="true">MIRA</p>
        <div class="footer-rule" aria-hidden="true"></div>
        <p class="legal">
          Mira does not host, store, or distribute any content. It only relays what the addons and
          services configured by its operator provide. You are solely responsible for the addons and
          services you configure and for complying with the laws that apply to you.
        </p>
        <p class="legal">
          Mira is an independent project, not affiliated with or endorsed by Jellyfin or Stremio.
          Released under the
          <a class="underline" :href="`${repository}/blob/main/LICENSE`">MIT License</a>.
        </p>
      </div>
    </footer>
  </div>
</template>

<style scoped>
.pf-home {
  --m-bg: #f4f0e7;
  --m-paper: #fffdf8;
  --m-ink: #15131a;
  --m-ink-2: #4e4a57;
  --m-mute: #7c7784;
  --m-line: rgb(21 19 26 / 0.12);
  --m-line-strong: rgb(21 19 26 / 0.22);
  --m-dark: #0b0a12;
  --m-dark-2: #12101c;
  --m-dark-line: rgb(255 255 255 / 0.1);
  --m-nav-bg: rgb(244 240 231 / 0.84);
  --m-footer-bg: #ece6da;
  --m-strong: #15131a;
  --m-on-strong: #fff;
  --m-feature-s2: #ece7dd;
  --m-feature-s3: #e5dfd3;
  --m-violet: #6d4aff;
  --m-cyan: #22d3ee;
  position: relative;
  overflow-x: clip;
  min-height: 100vh;
  background: var(--m-bg);
  color: var(--m-ink-2);
  font-size: 16px;
  line-height: 1.6;
}

.pf-home[data-theme='dark'] {
  --m-bg: #0b0a12;
  --m-paper: #12101c;
  --m-ink: #f4f0e7;
  --m-ink-2: #b9b2c7;
  --m-mute: #8f88a0;
  --m-line: rgb(255 255 255 / 0.12);
  --m-line-strong: rgb(255 255 255 / 0.24);
  --m-dark: #05040a;
  --m-dark-2: #0b0a12;
  --m-dark-line: rgb(255 255 255 / 0.1);
  --m-nav-bg: rgb(11 10 18 / 0.84);
  --m-footer-bg: #07060c;
  --m-strong: #f4f0e7;
  --m-on-strong: #15131a;
  --m-feature-s2: #1a1724;
  --m-feature-s3: #221e2d;
}

.wrap {
  width: 100%;
  max-width: 1288px;
  margin-inline: auto;
  padding-inline: clamp(20px, 4vw, 48px);
}

h1,
h2,
h3,
b {
  font-family: var(--pf-display);
}

h1,
h2,
h3 {
  margin: 0;
  font-weight: 700;
  color: var(--m-ink);
  letter-spacing: -0.045em;
  line-height: 1;
  text-wrap: balance;
}

p {
  margin: 0;
  text-wrap: pretty;
}

code {
  padding: 2px 7px;
  border-radius: 6px;
  background: rgb(255 255 255 / 0.08);
  font-family: var(--vp-font-family-mono);
  font-size: 0.86em;
  color: inherit;
  white-space: nowrap;
}

.mono {
  font-family: var(--vp-font-family-mono);
  font-size: 12px;
}

.underline {
  color: inherit;
  text-decoration: underline;
  text-decoration-color: color-mix(in srgb, var(--m-cyan) 55%, transparent);
  text-decoration-thickness: 1.5px;
  text-underline-offset: 3px;
}

.underline:hover {
  text-decoration-color: var(--m-cyan);
}

/* ------------------------------------------------ nav */

.m-nav {
  position: sticky;
  top: 0;
  z-index: 50;
  border-bottom: 1px solid var(--m-line);
  background: var(--m-nav-bg);
  backdrop-filter: blur(18px) saturate(1.2);
}

.m-nav-inner {
  display: flex;
  align-items: center;
  gap: 30px;
  height: 74px;
}

.m-brand {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  color: var(--m-ink);
  font: 700 18px/1 var(--pf-display);
  letter-spacing: -0.03em;
}

.m-brand img {
  border-radius: 9px;
}

.m-nav .m-brand img {
  width: 42px;
  height: 42px;
  border-radius: 11px;
}

.m-links {
  display: flex;
  gap: 26px;
  margin-left: 10px;
}

.m-links a,
.m-github {
  color: var(--m-ink-2);
  font-size: 14px;
  font-weight: 500;
  transition: color 0.25s var(--pf-ease);
}

.m-links a:hover,
.m-github:hover {
  color: var(--m-ink);
}

.m-nav-actions {
  display: flex;
  align-items: center;
  gap: 20px;
  margin-left: auto;
}

.m-cta {
  display: inline-flex;
  align-items: center;
  height: 40px;
  padding: 0 18px;
  border-radius: 999px;
  background: var(--m-strong);
  color: var(--m-on-strong);
  font: 600 13.5px/1 var(--pf-display);
  transition:
    background-color 0.25s var(--pf-ease),
    transform 0.25s var(--pf-ease);
}

.m-cta:hover {
  background: var(--m-violet);
  color: #fff;
  transform: translateY(-1px);
}

.m-theme {
  display: grid;
  place-items: center;
  width: 38px;
  height: 38px;
  border: 1px solid var(--m-line);
  border-radius: 50%;
  background: transparent;
  color: var(--m-ink);
  cursor: pointer;
  transition:
    border-color 0.25s var(--pf-ease),
    background-color 0.25s var(--pf-ease);
}

.m-theme:hover {
  border-color: var(--m-line-strong);
  background: var(--m-paper);
}

.m-theme svg {
  width: 17px;
  height: 17px;
}

.m-docs-mobile {
  display: none;
  color: var(--m-ink);
  font-size: 14px;
  font-weight: 600;
}

/* ------------------------------------------------ type */

.kicker {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  margin-bottom: 22px;
  color: var(--m-mute);
  font: 500 11.5px/1 var(--vp-font-family-mono);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.kicker-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--m-violet), var(--m-cyan));
  box-shadow: 0 0 0 4px rgb(109 74 255 / 0.1);
}

.kicker.light {
  color: #958da6;
}

.grad {
  background: linear-gradient(105deg, var(--m-violet) 12%, #4c8ff7 52%, var(--m-cyan));
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
}

.lead {
  max-width: 33em;
  margin-top: 26px;
  color: var(--m-ink-2);
  font-size: clamp(16.5px, 1.3vw, 18.5px);
  line-height: 1.65;
}

.head {
  max-width: 820px;
  margin-bottom: clamp(44px, 6vw, 70px);
}

.head h2,
.split-head h2 {
  font-size: clamp(36px, 4.6vw, 62px);
}

.head-lead {
  max-width: 390px;
  color: var(--m-ink-2);
  font-size: 16px;
  line-height: 1.6;
}

.split-head {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 40px;
  margin-bottom: clamp(38px, 5vw, 60px);
  padding-bottom: 28px;
  border-bottom: 1px solid var(--m-line);
}

/* ------------------------------------------------ buttons */

.actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-top: 32px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  height: 50px;
  padding: 0 22px;
  border-radius: 999px;
  font: 600 14px/1 var(--pf-display);
  cursor: pointer;
  transition:
    transform 0.25s var(--pf-ease),
    background-color 0.25s var(--pf-ease),
    border-color 0.25s var(--pf-ease);
}

.btn:hover {
  transform: translateY(-2px);
}

.btn-ink {
  background: var(--m-strong);
  color: var(--m-on-strong);
}

.btn-ink:hover {
  background: var(--m-violet);
  color: #fff;
}

.btn-knob {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--m-on-strong) 14%, transparent);
}

.btn-line {
  border: 1px solid var(--m-line-strong);
  background: transparent;
  color: var(--m-ink);
}

.btn-line:hover {
  border-color: var(--m-ink);
  background: var(--m-paper);
}

.play-dot {
  width: 22px;
  height: 22px;
  border-radius: 50%;
  background: var(--m-cyan);
  position: relative;
}

.play-dot::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 52%;
  width: 0;
  height: 0;
  transform: translate(-50%, -50%);
  border-left: 6px solid var(--m-ink);
  border-block: 4px solid transparent;
}

.hero-facts {
  display: flex;
  flex-wrap: wrap;
  gap: 34px;
  margin: 46px 0 0;
  padding: 0;
  list-style: none;
}

.hero-facts b {
  display: block;
  color: var(--m-ink);
  font-size: 18px;
  letter-spacing: -0.03em;
}

.hero-facts span {
  color: var(--m-mute);
  font: 400 11.5px/1.4 var(--vp-font-family-mono);
}

/* ------------------------------------------------ hero */

.hero {
  padding: clamp(52px, 7vw, 92px) 0 clamp(52px, 6vw, 80px);
}

.hero-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.04fr) minmax(0, 0.96fr);
  align-items: center;
  gap: clamp(36px, 5vw, 72px);
}

.hero h1 {
  max-width: 10.5em;
  font-size: clamp(52px, 6.6vw, 94px);
  line-height: 0.92;
  letter-spacing: -0.065em;
}

.hero-art {
  position: relative;
  justify-self: end;
  width: min(100%, 520px);
  padding: 36px 0 30px;
}

.art-glow {
  position: absolute;
  inset: -18% -12% -10% -14%;
  z-index: -1;
  background:
    radial-gradient(circle at 62% 36%, rgb(109 74 255 / 0.26), transparent 52%),
    radial-gradient(circle at 28% 78%, rgb(34 211 238 / 0.18), transparent 50%);
  filter: blur(22px);
}

.device {
  position: relative;
  z-index: 1;
  padding: 16px;
  border: 1px solid rgb(255 255 255 / 0.12);
  border-radius: 26px;
  background: radial-gradient(circle at 80% 0%, rgb(109 74 255 / 0.2), transparent 45%), #0b0a12;
  box-shadow:
    0 34px 90px rgb(35 24 80 / 0.28),
    0 2px 0 rgb(255 255 255 / 0.08) inset;
  transform: rotate(1.2deg);
}

.device-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 2px 4px 14px;
  color: #d8d3e6;
  font: 500 12px/1 var(--vp-font-family-mono);
}

.device-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--m-cyan);
}

.device-name {
  color: #f5f2fb;
  font-weight: 600;
}

.device-pill {
  margin-left: auto;
  padding: 4px 8px;
  border: 1px solid rgb(255 255 255 / 0.12);
  border-radius: 999px;
  color: #a49bb9;
  font-size: 10px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.device-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 9px;
}

.device-grid img {
  display: block;
  width: 100%;
  aspect-ratio: 2 / 3;
  object-fit: cover;
  border-radius: 9px;
  box-shadow: 0 8px 18px rgb(0 0 0 / 0.35);
}

.device-foot {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 4px 2px;
  color: #a49bb9;
  font: 500 10.5px/1 var(--vp-font-family-mono);
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.device-foot i {
  flex: 1;
  height: 3px;
  border-radius: 999px;
  background: rgb(255 255 255 / 0.12);
}

.device-foot b {
  display: block;
  width: 38%;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--m-violet), var(--m-cyan));
}

.float-mark {
  position: absolute;
  top: -20px;
  left: -36px;
  z-index: 3;
  width: 96px;
  height: 96px;
  border-radius: 22px;
  filter: drop-shadow(0 18px 30px rgb(35 24 80 / 0.4));
  transform: rotate(-8deg);
  animation: float 7s ease-in-out infinite;
}

.chip {
  position: absolute;
  z-index: 3;
  padding: 9px 13px;
  border: 1px solid var(--m-line);
  border-radius: 999px;
  background: color-mix(in srgb, var(--m-paper) 92%, transparent);
  box-shadow: 0 12px 28px rgb(35 24 80 / 0.14);
  color: var(--m-ink);
  font: 600 11.5px/1 var(--pf-display);
  backdrop-filter: blur(8px);
}

.chip-a {
  top: 90px;
  right: -18px;
}

.chip-b {
  right: -8px;
  bottom: 102px;
}

.chip-c {
  bottom: 52px;
  left: -28px;
}

@keyframes float {
  0%,
  100% {
    transform: rotate(-8deg) translateY(0);
  }
  50% {
    transform: rotate(-6deg) translateY(-8px);
  }
}

/* ------------------------------------------------ works */

.works {
  border-block: 1px solid var(--m-line);
  padding: 17px 0;
}

.works-line {
  display: flex;
  align-items: center;
  gap: 24px;
}

.works-label {
  flex: none;
  color: var(--m-mute);
  font: 500 11px/1 var(--vp-font-family-mono);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.marquee {
  flex: 1;
  overflow: hidden;
  mask-image: linear-gradient(90deg, transparent, #000 6%, #000 94%, transparent);
}

.marquee-track {
  display: flex;
  width: max-content;
  gap: 38px;
  animation: marquee 34s linear infinite;
}

.marquee-track span {
  color: var(--m-ink-2);
  font: 600 13.5px/1 var(--pf-display);
  letter-spacing: -0.01em;
  white-space: nowrap;
}

.marquee:hover .marquee-track {
  animation-play-state: paused;
}

@keyframes marquee {
  to {
    transform: translateX(-50%);
  }
}

.works-link {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 5px;
  color: var(--m-ink);
  font-size: 13px;
  font-weight: 500;
}

/* ------------------------------------------------ sections */

.section {
  padding: clamp(76px, 9vw, 130px) 0;
}

/* The in-page ids sit on zero-height anchors just above each section:
   the router then lands with the sticky nav clear of the heading. */
.jump-anchor {
  display: block;
  position: relative;
  top: -92px;
  height: 0;
  visibility: hidden;
  pointer-events: none;
}

.m-ink {
  background: var(--m-dark);
  color: #b9b2c7;
}

.m-ink h2,
.m-ink h3,
.m-ink b {
  color: #f8f5ef;
}

.m-ink .head-lead {
  color: #a9a2b8;
}

.m-ink .split-head {
  border-bottom-color: var(--m-dark-line);
}

.m-ink .kicker {
  color: #8f88a0;
}

/* ------------------------------------------------ how */

.pipeline {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 210px minmax(0, 1fr);
  align-items: stretch;
  gap: 18px;
}

.pipe-card {
  padding: 30px;
  border: 1px solid var(--m-dark-line);
  border-radius: 24px;
  background: linear-gradient(160deg, rgb(255 255 255 / 0.04), transparent 60%), var(--m-dark-2);
}

.pipe-index {
  display: block;
  margin-bottom: 28px;
  color: #8f88a0;
  font: 500 11px/1 var(--vp-font-family-mono);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.pipe-card h3 {
  margin-bottom: 22px;
  font-size: 25px;
}

.pipe-card ul {
  display: grid;
  gap: 12px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.pipe-card li {
  display: flex;
  align-items: baseline;
  gap: 10px;
  color: #b9b2c7;
  font-size: 14.5px;
}

.pipe-card li::before {
  content: '';
  flex: none;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--m-violet), var(--m-cyan));
}

.pipe-hub {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 9px;
  text-align: center;
}

.pipe-hub img {
  position: relative;
  z-index: 1;
  border-radius: 18px;
  box-shadow: 0 18px 40px rgb(0 0 0 / 0.4);
}

.pipe-hub b {
  position: relative;
  z-index: 1;
  margin-top: 4px;
  font-size: 18px;
  letter-spacing: -0.03em;
}

.pipe-hub small {
  position: relative;
  z-index: 1;
  color: #8f88a0;
  font: 500 10.5px/1 var(--vp-font-family-mono);
  letter-spacing: 0.12em;
  text-transform: uppercase;
}

.pipe-line {
  position: absolute;
  top: 50%;
  left: -18px;
  right: -18px;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--m-violet), var(--m-cyan), transparent);
  opacity: 0.45;
}

.pipe-foot {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px 30px;
  margin-top: 34px;
  color: #8f88a0;
  font: 500 11.5px/1.4 var(--vp-font-family-mono);
}

.pipe-foot span {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.pipe-foot span + span::before {
  content: '+';
  color: var(--m-violet);
}

/* ------------------------------------------------ tour & features */

.tour {
  background: var(--m-bg);
}

.features {
  --pf-ink: var(--m-ink);
  --pf-ink-2: var(--m-ink-2);
  --pf-mute: var(--m-mute);
  --pf-mute-2: #9a94a1;
  --pf-line: var(--m-line);
  --pf-s2: var(--m-feature-s2);
  --pf-s3: var(--m-feature-s3);
  --pf-display: 'Sora', ui-sans-serif, system-ui, sans-serif;
  background: var(--m-paper);
  border-block: 1px solid var(--m-line);
}

/* ------------------------------------------------ shelf */

.shelf {
  position: relative;
  padding: 26px 0 34px;
  overflow: clip;
  background: var(--m-bg);
}

.shelf::before,
.shelf::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  z-index: 2;
  width: 10vw;
  pointer-events: none;
}

.shelf::before {
  left: 0;
  background: linear-gradient(90deg, var(--m-bg), transparent);
}

.shelf::after {
  right: 0;
  background: linear-gradient(-90deg, var(--m-bg), transparent);
}

.rows {
  display: grid;
  gap: 14px;
}

.row {
  overflow: hidden;
}

.track {
  display: flex;
  width: max-content;
  gap: 14px;
  animation: shelf 62s linear infinite;
}

.row:nth-child(2) .track {
  animation-direction: reverse;
  animation-duration: 74s;
}

.track img {
  display: block;
  width: 150px;
  height: 200px;
  border-radius: 11px;
  object-fit: cover;
  box-shadow: 0 12px 26px rgb(35 24 80 / 0.14);
}

@keyframes shelf {
  to {
    transform: translateX(-50%);
  }
}

/* ------------------------------------------------ extras */

.extras {
  background: var(--m-bg);
}

.extra-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.extra-grid li {
  min-width: 0;
}

/* The row ends with Quick Connect: the web client card spans the other two
   columns so that the row is complete, without a hole. */
.extra-grid li:last-child {
  grid-column: span 2;
}

.extra-grid a {
  display: flex;
  flex-direction: column;
  justify-content: center;
  height: 100%;
  overflow: hidden;
  border: 1px solid var(--m-line);
  border-radius: 20px;
  background: var(--m-paper);
  transition:
    transform 0.3s var(--pf-ease),
    border-color 0.3s var(--pf-ease),
    box-shadow 0.3s var(--pf-ease);
}

.extra-grid a:hover {
  transform: translateY(-3px);
  border-color: var(--m-line-strong);
  box-shadow: 0 18px 40px rgb(35 24 80 / 0.1);
}

.extra-shot {
  display: block;
  overflow: hidden;
  border-bottom: 1px solid var(--m-line);
}

.extra-shot img {
  display: block;
  width: 100%;
  height: 170px;
  object-fit: cover;
  transition: transform 0.6s var(--pf-ease);
}

.extra-grid a:hover .extra-shot img {
  transform: scale(1.035);
}

.extra-body {
  display: flex;
  align-items: flex-start;
  gap: 13px;
  padding: 20px;
}

.extra-icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 38px;
  height: 38px;
  border-radius: 11px;
  background: linear-gradient(145deg, var(--m-violet), #4c8ff7);
  color: #fff;
}

.extra-icon.cyan {
  background: linear-gradient(145deg, #3e8ef7, var(--m-cyan));
  color: #071019;
}

.extra-copy {
  display: grid;
  gap: 5px;
  min-width: 0;
}

.extra-copy b {
  color: var(--m-ink);
  font-size: 16px;
  font-weight: 600;
  letter-spacing: -0.025em;
}

.extra-copy span {
  color: var(--m-ink-2);
  font-size: 13.5px;
  line-height: 1.5;
}

.extra-go {
  flex: none;
  margin-left: auto;
  margin-top: 3px;
  color: var(--m-mute);
}

.extra-paths {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 0 20px 20px;
}

.extra-paths span {
  padding: 6px 9px;
  border: 1px solid var(--m-line);
  border-radius: 7px;
  color: var(--m-mute);
  font: 500 10.5px/1 var(--vp-font-family-mono);
}

.extra-paths em {
  color: var(--m-ink);
  font-style: normal;
}

/* The web client card fills the row beside Quick Connect: its copy and
   its two paths sit side by side, so the wide card has no empty half. */
@media (min-width: 861px) {
  .extra-grid li:last-child a {
    flex-direction: row;
    align-items: center;
  }

  .extra-grid li:last-child .extra-body {
    flex: 1;
  }

  .extra-grid li:last-child .extra-paths {
    flex-direction: column;
    padding: 20px 22px;
    margin-left: auto;
  }
}

/* ------------------------------------------------ install */

.install-grid {
  display: grid;
  grid-template-columns: minmax(0, 0.92fr) minmax(0, 1.08fr);
  gap: clamp(40px, 6vw, 80px);
  align-items: start;
}

.install h2 {
  max-width: 10em;
}

.install .head-lead {
  margin-top: 22px;
}

.steps {
  display: grid;
  gap: 14px;
  margin: 34px 0 0;
  padding: 0;
  list-style: none;
}

.steps li {
  display: flex;
  align-items: flex-start;
  gap: 13px;
  color: #b9b2c7;
  font-size: 14.5px;
}

.steps .n {
  display: grid;
  place-items: center;
  flex: none;
  width: 25px;
  height: 25px;
  border: 1px solid var(--m-dark-line);
  border-radius: 50%;
  color: #f8f5ef;
  font: 600 11px/1 var(--vp-font-family-mono);
}

.more {
  margin-top: 28px;
  color: #8f88a0;
  font-size: 13.5px;
}

.install-side {
  display: grid;
  gap: 14px;
}

.terminal {
  overflow: hidden;
  border: 1px solid var(--m-dark-line);
  border-radius: 20px;
  background: #06050c;
  box-shadow: 0 24px 60px rgb(0 0 0 / 0.35);
}

.terminal-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 16px;
  border-bottom: 1px solid var(--m-dark-line);
  color: #8f88a0;
  font: 500 11px/1 var(--vp-font-family-mono);
}

.terminal-dots {
  display: inline-flex;
  gap: 5px;
}

.terminal-dots i {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #3a3548;
}

.terminal-dots i:first-child {
  background: #ff6b6b;
}

.terminal-dots i:nth-child(2) {
  background: #f4c867;
}

.terminal-dots i:nth-child(3) {
  background: #4adcf2;
}

.terminal-title {
  margin-left: 2px;
}

.copy-button {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  margin-left: auto;
  padding: 5px 9px;
  border: 1px solid var(--m-dark-line);
  border-radius: 7px;
  color: #b9b2c7;
  font: 500 10.5px/1 var(--vp-font-family-mono);
  cursor: pointer;
}

.copy-button:hover {
  color: #fff;
  border-color: rgb(255 255 255 / 0.24);
}

.copy-button.done {
  color: var(--m-cyan);
}

.terminal pre {
  margin: 0;
  padding: 20px 20px 22px;
  overflow-x: auto;
  font: 400 13px/1.8 var(--vp-font-family-mono);
}

.terminal code {
  padding: 0;
  background: none;
  font-size: inherit;
  white-space: pre;
}

.terminal .comment {
  color: #6f6880;
}

.terminal .command {
  color: #e9e5f0;
}

.desktop {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 17px;
  border: 1px solid var(--m-dark-line);
  border-radius: 17px;
  background: var(--m-dark-2);
  transition:
    border-color 0.3s var(--pf-ease),
    transform 0.3s var(--pf-ease);
}

.desktop:hover {
  border-color: rgb(255 255 255 / 0.22);
  transform: translateY(-2px);
}

.desktop-mark img {
  display: block;
  width: 38px;
  height: 38px;
  border-radius: 10px;
}

.desktop-text {
  display: grid;
  gap: 3px;
  min-width: 0;
}

.desktop-text small {
  color: #8f88a0;
  font: 500 10.5px/1 var(--vp-font-family-mono);
}

.desktop-text b {
  color: #f8f5ef;
  font-size: 15px;
  letter-spacing: -0.02em;
}

.desktop-text span {
  color: #a9a2b8;
  font-size: 12px;
  line-height: 1.45;
}

.systems {
  display: flex;
  gap: 7px;
  margin-left: auto;
  color: #8f88a0;
}

.desktop-go {
  flex: none;
  margin-top: 3px;
  color: #8f88a0;
}

.note {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  margin-top: clamp(48px, 6vw, 70px);
  padding-top: 22px;
  border-top: 1px solid var(--m-dark-line);
  color: #8f88a0;
  font-size: 13px;
}

/* ------------------------------------------------ footer */

.footer {
  background: var(--m-footer-bg);
  border-top: 1px solid var(--m-line);
  padding: clamp(34px, 4vw, 48px) 0 26px;
}

.footer-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 40px;
}

.footer-top p {
  margin-top: 12px;
  color: var(--m-mute);
  font-size: 13.5px;
}

.footer-links {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 28px;
  padding-top: 4px;
}

.footer-links a {
  color: var(--m-ink-2);
  font-size: 13.5px;
  font-weight: 500;
}

.footer-links a:hover {
  color: var(--m-ink);
}

.footer-word {
  margin: 24px 0 10px;
  color: transparent;
  font: 800 clamp(40px, 5vw, 64px) / 0.9 var(--pf-display);
  letter-spacing: -0.06em;
  -webkit-text-stroke: 1px color-mix(in srgb, var(--m-ink) 20%, transparent);
  user-select: none;
}

.footer-rule {
  height: 2px;
  margin-bottom: 22px;
  background: linear-gradient(90deg, var(--m-violet), #4c8ff7, var(--m-cyan));
  opacity: 0.55;
}

.legal {
  max-width: 940px;
  color: var(--m-mute);
  font-size: 12.5px;
  line-height: 1.6;
}

.legal + .legal {
  margin-top: 9px;
}

/* ------------------------------------------------ reveal */

.rv {
  transition:
    opacity 0.8s var(--pf-ease),
    transform 0.8s var(--pf-ease);
}

.rv:not(.in) {
  opacity: 0;
  transform: translateY(24px);
}

/* ------------------------------------------------ responsive */

@media (max-width: 1080px) {
  .hero-grid {
    grid-template-columns: minmax(0, 1fr) minmax(0, 0.82fr);
  }

  .m-links {
    gap: 18px;
  }

  .pipeline {
    grid-template-columns: 1fr 150px 1fr;
  }

  .pipe-card {
    padding: 24px;
  }

  .extra-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .extra-grid li:last-child {
    grid-column: auto;
  }
}

@media (max-width: 860px) {
  .m-links,
  .m-github {
    display: none;
  }

  .m-docs-mobile {
    display: inline;
  }

  .m-nav-actions {
    gap: 14px;
  }

  .hero-grid {
    grid-template-columns: 1fr;
  }

  .hero-art {
    justify-self: center;
    width: min(100%, 500px);
    margin-top: 10px;
  }

  .split-head,
  .footer-top {
    flex-direction: column;
    align-items: flex-start;
  }

  .pipeline {
    grid-template-columns: 1fr;
  }

  .pipe-hub {
    padding: 12px 0;
  }

  .pipe-line {
    display: none;
  }

  .install-grid {
    grid-template-columns: 1fr;
  }

  .extra-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 560px) {
  .hero h1 {
    font-size: clamp(46px, 14vw, 64px);
  }

  .float-mark {
    left: -10px;
    width: 78px;
    height: 78px;
  }

  .chip {
    display: none;
  }

  .hero-facts {
    gap: 22px;
  }

  .extra-grid {
    grid-template-columns: 1fr;
  }

  .works-link,
  .works-label {
    display: none;
  }

  .footer-word {
    font-size: 13vw;
  }
}
</style>
