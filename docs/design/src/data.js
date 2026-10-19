// Mock data shared by the three concept boards. Plan content follows the coach's real plans
// (letters A–E, warm-up/cool-down cardio, sets × reps, per-leg/per-side, "+" supersets);
// athlete names are made up.

const fa = (v) => String(v).replace(/\d/g, (d) => "۰۱۲۳۴۵۶۷۸۹"[d]);

const PRIVATE = [
  ["مهسا", "رضایی", "#E4572E"], ["سارا", "کریمی", "#F3A712"], ["نگار", "احمدی", "#29A36A"],
  ["الهام", "موسوی", "#2E86DE"], ["نازنین", "حسینی", "#8E44AD"], ["زهرا", "قاسمی", "#E84393"],
  ["مریم", "صادقی", "#16A3A3"], ["فاطمه", "رحیمی", "#8D6E63"], ["پریسا", "نوری", "#C0392B"],
  ["شیما", "جعفری", "#7FB800"], ["لیلا", "کاظمی", "#F06292"], ["هانیه", "مرادی", "#607D8B"],
  ["رویا", "باقری", "#FF7F50"], ["آزاده", "یزدانی", "#5C6BC0"], ["ترانه", "شریفی", "#00897B"],
].map(([first, last, color], i) => ({ first, last, color, days: i % 4 === 3 ? 4 : 5 }));

const PUBLIC_COUNT = 9;

// A movement part: r = reps, s = seconds, side = "هرپا" | "هرسمت".
// An item with several parts is a superset (written with "+" in her plans).
const DAY_A = {
  warm: "دوچرخه ۱۰ دقیقه",
  cool: "دوچرخه ۱۰ دقیقه",
  items: [
    { sets: 3, parts: [{ n: "هیپ تراست تک پا صفحه", r: 15, side: "هرپا" }, { n: "پل باسن هولد", s: 45 }] },
    { sets: 3, parts: [{ n: "کیک بک دستگاه", r: 15, side: "هرپا" }] },
    { sets: 3, parts: [{ n: "پشت پا خوابیده تک پا", r: 15, side: "هرپا" }] },
    { sets: 3, parts: [{ n: "ددلیفت تک پا آزاد با دمبل", r: 15, side: "هرپا" }] },
    { sets: 3, parts: [{ n: "اسکوات ساق پا در نیمه بالا", r: 15 }] },
    { sets: 3, parts: [{ n: "داخل ران دمبل نشسته", r: 15, side: "هرسمت" }] },
    { sets: 4, parts: [{ n: "داخل ران دستگاه، مکث ۲ ثانیه", r: 15 }] },
    { sets: 3, parts: [{ n: "شکم مورب تک پا ۹۰ درجه", r: 15, side: "هرسمت" }] },
    { sets: 3, parts: [{ n: "شکم وی", r: 15 }] },
  ],
};

// The day being written in the builder boards: five rows in, the fifth open for editing.
const DAY_C_DRAFT = {
  warm: "دوچرخه ۱۰ دقیقه",
  items: [
    { sets: 3, parts: [{ n: "جلو ران نیمه و کامل", r: 12 }] },
    { sets: 3, parts: [{ n: "لانژ پا جمع در شکم دمبل", r: 12, side: "هرپا" }] },
    { sets: 3, parts: [{ n: "اسکوات هالتر پاجمع نیمه", r: 12 }] },
    { sets: 3, parts: [{ n: "وال سیت صفحه", s: 30 }] },
    { sets: 3, parts: [{ n: "جلو ران کراس نشسته", r: 12, side: "هرپا" }] },
  ],
};

const SEARCH_QUERY = "پلان";
const SEARCH_HITS = ["پلانک", "پلانک + اسکوات هولد", "پلانک کوماندو", "پلانک چرخشی"];
const PRESETS = ["۱۲ × ۳", "۱۵ × ۳", "۲۰ × ۳", "۱۵ × ۴", "۳۰ ثانیه × ۳"];

const part = (p) => (p.s ? `${fa(p.s)} ثانیه` : `${fa(p.r)} تکرار${p.side ? " " + p.side : ""}`);
const rx = (it) => `${it.parts.map(part).join(" + ")} × ${fa(it.sets)} ست`;

const ICON = {
  search: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/></svg>',
  back: '<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2.2"><path d="m9 6 6 6-6 6"/></svg>',
  edit: '<svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 20h4L19 9l-4-4L4 16z"/></svg>',
  plus: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.4"><path d="M12 5v14M5 12h14"/></svg>',
  grip: '<svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor"><circle cx="9" cy="6" r="1.6"/><circle cx="15" cy="6" r="1.6"/><circle cx="9" cy="12" r="1.6"/><circle cx="15" cy="12" r="1.6"/><circle cx="9" cy="18" r="1.6"/><circle cx="15" cy="18" r="1.6"/></svg>',
  link: '<svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2.2"><path d="M10 14a4 4 0 0 0 5.7 0l3-3a4 4 0 0 0-5.7-5.7l-1 1"/><path d="M14 10a4 4 0 0 0-5.7 0l-3 3a4 4 0 0 0 5.7 5.7l1-1"/></svg>',
  copy: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><rect x="8" y="8" width="12" height="12" rx="2"/><path d="M16 8V5a1 1 0 0 0-1-1H5a1 1 0 0 0-1 1v10a1 1 0 0 0 1 1h3"/></svg>',
  clock: '<svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="8"/><path d="M12 8v4l3 2"/></svg>',
  trash: '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 7h14M10 7V5h4v2M7 7l1 13h8l1-13"/></svg>',
  users: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><circle cx="9" cy="8" r="3.5"/><path d="M3 20c0-3.3 2.7-6 6-6s6 2.7 6 6M16 4.5a3.5 3.5 0 0 1 0 7M21 20c0-2.6-1.6-4.8-4-5.6"/></svg>',
  dumbbell: '<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 7v10M3 9v6M18 7v10M21 9v6M6 12h12"/></svg>',
  chevL: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="m15 6-6 6 6 6"/></svg>',
  down: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="m6 9 6 6 6-6"/></svg>',
  alert: '<svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2.2"><path d="M12 4 2.5 20h19z"/><path d="M12 10v4M12 17v.5"/></svg>',
  x: '<svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 6l12 12M18 6 6 18"/></svg>',
  check: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2.4"><path d="m5 12 5 5 9-10"/></svg>',
  calendar: '<svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="5" width="16" height="15" rx="2"/><path d="M4 10h16M9 3v4M15 3v4"/></svg>',
  user: '<svg viewBox="0 0 24 24" width="17" height="17" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="8" r="4"/><path d="M4 21c0-4.4 3.6-8 8-8s8 3.6 8 8"/></svg>',
  phone: '<svg viewBox="0 0 24 24" width="15" height="15" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a1 1 0 0 1-1 1A16 16 0 0 1 4 5a1 1 0 0 1 1-1z"/></svg>',
  lock: '<svg viewBox="0 0 24 24" width="12" height="12" fill="none" stroke="currentColor" stroke-width="2.4"><rect x="5" y="11" width="14" height="10" rx="2"/><path d="M8 11V8a4 4 0 0 1 8 0v3"/></svg>',
  eye: '<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2"><path d="M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7S2 12 2 12z"/><circle cx="12" cy="12" r="3"/></svg>',
  spark: '<svg viewBox="0 0 24 24" width="14" height="14" fill="currentColor"><path d="M12 2c.5 4.5 2.5 6.6 7 7-4.5.5-6.5 2.5-7 7-.5-4.5-2.5-6.5-7-7 4.5-.4 6.5-2.5 7-7z"/><path d="M19 14c.2 1.9 1 2.8 3 3-2 .2-2.8 1-3 3-.2-2-1-2.8-3-3 2-.2 2.8-1.1 3-3z"/></svg>',
};

// Fields the athlete cards in the classic concept show under the name (made-up values).
const GOALS = ["کاهش وزن", "فرم‌دهی بدن", "افزایش حجم", "افزایش قدرت", "حجم پایین‌تنه"];
const LEVELS = ["مبتدی", "متوسط", "حرفه‌ای"];
const CATEGORIES = ["گرم کردن", "قدرتی", "هوازی"];

// Swaps every ICON_<key> placeholder in the page's markup for ICON[key], e.g. ICON_chevL.
function fillIcons() {
  document.body.innerHTML = document.body.innerHTML.replace(/ICON_(\w+)/g, (m, key) => ICON[key] ?? m);
}

// Fills every element carrying data-fill="<name>" from the matching render function.
function fill(renderers) {
  document.querySelectorAll("[data-fill]").forEach((el) => {
    el.innerHTML = renderers[el.dataset.fill]();
  });
}
