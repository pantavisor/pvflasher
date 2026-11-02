// pvflasher site: block-map demo, OS detection for downloads, latest release, copy buttons.
const root = document.documentElement;
root.classList.remove("no-js");
root.classList.add("js");

const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

/* ---------- Block map: write mapped blocks, skip empty ones, then read back ---------- */
function blockMap() {
  const grid = document.getElementById("bmap");
  if (!grid || reduceMotion) return;
  const phase = document.getElementById("bmap-phase");
  const count = document.getElementById("bmap-count");
  const cells = [...grid.children];
  const mapped = cells.filter((c) => c.classList.contains("m"));
  const mbPerBlock = 715 / cells.length; // the real image is 715 MB raw
  const fmt = (n) => Math.round(n * mbPerBlock);

  let timer = null;
  const tickMs = 9;     // per block while writing
  const verifyMs = 4;   // per block while reading back
  const perTick = 2;

  function setStatus(label, text) {
    phase.textContent = label;
    count.textContent = text;
  }

  function reset() {
    mapped.forEach((c) => c.classList.remove("w", "v"));
    cells.forEach((c) => c.classList.remove("head"));
  }

  function write() {
    reset();
    let i = 0;
    let written = 0;
    let skipped = 0;
    timer = setInterval(() => {
      for (let k = 0; k < perTick && i < cells.length; k++, i++) {
        const c = cells[i];
        cells[i - 1]?.classList.remove("head");
        if (c.classList.contains("m")) { c.classList.add("w", "head"); written++; }
        else skipped++;
      }
      setStatus("Writing", `${fmt(written)} MB written · ${fmt(skipped)} MB skipped`);
      if (i >= cells.length) {
        clearInterval(timer);
        cells.forEach((c) => c.classList.remove("head"));
        setTimeout(verify, 500);
      }
    }, tickMs);
  }

  function verify() {
    let i = 0;
    timer = setInterval(() => {
      for (let k = 0; k < perTick * 2 && i < mapped.length; k++, i++) mapped[i].classList.add("v");
      setStatus("Verifying", `${fmt(i)} of ${fmt(mapped.length)} MB checked`);
      if (i >= mapped.length) {
        clearInterval(timer);
        setStatus("Verified", `${fmt(mapped.length)} MB written · ${fmt(cells.length - mapped.length)} MB skipped`);
        timer = setTimeout(write, 4200);
      }
    }, verifyMs);
  }

  // Start once the card is on screen; pause when it scrolls away.
  let running = false;
  const io = new IntersectionObserver(([e]) => {
    if (e.isIntersecting && !running) { running = true; timer = setTimeout(write, 900); }
    else if (!e.isIntersecting && running) {
      running = false; clearInterval(timer); clearTimeout(timer);
      mapped.forEach((c) => c.classList.add("w", "v"));
      cells.forEach((c) => c.classList.remove("head"));
      setStatus("Verified", `${fmt(mapped.length)} MB written · ${fmt(cells.length - mapped.length)} MB skipped`);
    }
  }, { threshold: 0.4 });
  io.observe(grid);
}

/* ---------- Benchmark bars grow in when scrolled to ---------- */
function chart() {
  const el = document.querySelector(".chart");
  if (!el || reduceMotion) return;
  const io = new IntersectionObserver(([e]) => {
    if (e.isIntersecting) { el.classList.add("is-in"); io.disconnect(); }
  }, { threshold: 0.3 });
  io.observe(el);
}

/* ---------- Downloads ---------- */
function detectOS() {
  const p = (navigator.userAgentData?.platform || navigator.platform || navigator.userAgent).toLowerCase();
  if (p.includes("win")) return "windows";
  if (p.includes("mac") || p.includes("iphone") || p.includes("ipad")) return "macos";
  return "linux";
}

function tabs() {
  const tabButtons = [...document.querySelectorAll('.tabs [role="tab"]')];
  const panels = [...document.querySelectorAll('.panel[role="tabpanel"]')];
  if (!tabButtons.length) return;
  const select = (os, focus) => {
    tabButtons.forEach((b) => {
      const on = b.dataset.os === os;
      b.setAttribute("aria-selected", String(on));
      b.tabIndex = on ? 0 : -1;
      if (on && focus) b.focus();
    });
    panels.forEach((p) => p.setAttribute("aria-hidden", String(p.dataset.os !== os)));
  };
  tabButtons.forEach((b, i) => {
    b.addEventListener("click", () => select(b.dataset.os));
    b.addEventListener("keydown", (e) => {
      const d = e.key === "ArrowRight" ? 1 : e.key === "ArrowLeft" ? -1 : 0;
      if (!d) return;
      const next = tabButtons[(i + d + tabButtons.length) % tabButtons.length];
      select(next.dataset.os, true);
    });
  });
  const os = detectOS();
  select(os);
  const names = { linux: "Linux", windows: "Windows", macos: "macOS" };
  const hero = document.querySelector("#hero-download span:first-child");
  if (hero) hero.textContent = `Download for ${names[os]}`;
}

async function latestRelease() {
  const dl = document.querySelector("[data-dl-base]");
  if (!dl) return;
  try {
    const res = await fetch("https://api.github.com/repos/pantavisor/pvflasher/releases/latest", {
      headers: { Accept: "application/vnd.github+json" },
    });
    if (!res.ok) return;
    const { tag_name: tag } = await res.json();
    if (!/^v\d+\.\d+\.\d+/.test(tag || "")) return;
    const num = tag.slice(1);
    const base = dl.dataset.dlBase + tag + "/";
    document.querySelectorAll("[data-version]").forEach((el) => { el.textContent = tag; });
    document.querySelectorAll("[data-file]").forEach((a) => {
      const file = a.dataset.file.replaceAll("{v}", tag).replaceAll("{n}", num);
      a.href = base + file;
      a.querySelector(".dl-file")?.replaceChildren(file);
    });
  } catch { /* keep the version the site was built with */ }
}

/* ---------- Copy buttons ---------- */
function copyButtons() {
  document.querySelectorAll("[data-copy]").forEach((btn) => {
    btn.addEventListener("click", async () => {
      const code = document.getElementById(btn.dataset.copy);
      try {
        await navigator.clipboard.writeText(code.textContent);
        btn.textContent = "Copied";
      } catch {
        const r = document.createRange();
        r.selectNodeContents(code);
        const s = getSelection();
        s.removeAllRanges();
        s.addRange(r);
        btn.textContent = "Selected";
      }
      setTimeout(() => { btn.textContent = "Copy"; }, 1600);
    });
  });
}

blockMap();
chart();
tabs();
latestRelease();
copyButtons();
