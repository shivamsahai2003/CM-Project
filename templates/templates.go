package templates

// KeywordTemplate for the Publisher Page
const KeywordTemplate = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>{{.Title}}</title><meta name="viewport" content="width=device-width, initial-scale=1"></head>
<body style="font:15px/1.5 Arial, sans-serif; margin:20px;">
  <h1>{{.Title}}</h1>
  <p style="color:#555;">Publisher: {{.PubKey}} | Total fetched: {{.TotalFetched}} | Total shown: {{.TotalShown}}</p>
  <p style="color:#555;">Slot: {{.Slot}} | CC: {{.CC}} | Domain: {{.D}} | LID: {{.LID}} | PID: {{.PID}} | Size: {{.TSize}}</p>
  <p style="color:#777;">Referrer: {{.KwRf}}</p>
  <p style="color:#777;">Page Title: {{.PTitle}}</p>
  <p style="color:#777;">Page URL: {{.RURL}}</p>
  <p style="color:#777;">Keyword ID: {{.KID}}</p>
  {{if .IsBot}}
    <div style="margin:10px 0; padding:10px; background:#fff7ed; border:1px solid #fed7aa; color:#9a3412; border-radius:6px;">Bot detected.</div>
  {{end}}
  <hr>
  <div style="display:grid; grid-template-columns: 1fr; gap:12px; margin-top:16px;">
    {{range .Groups}}
      <section style="border:1px solid #e5e7eb; border-radius:8px; padding:12px;">
        <div style="font-weight:600; margin-bottom:8px;">{{.Label}} ({{len .Keywords}})</div>
        {{if .Keywords}}
          <ul style="margin:0; padding-left:18px;">
            {{range .Keywords}}
              <li><span style="color:#0b57d0; font-weight:600;">{{.Name}}</span></li>
            {{end}}
          </ul>
        {{else}}
          <div style="color:#999;">No keywords in this group</div>
        {{end}}
      </section>
    {{end}}
  </div>
  <p style="margin-top:12px;"><a href="javascript:history.back()">Back</a></p>
</body></html>`

// SerpTemplate is the HTML template for the SERP page
const SerpTemplate = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>{{.Title}}</title><meta name="viewport" content="width=device-width, initial-scale=1"></head>
<body style="font:15px/1.5 Arial, sans-serif; margin:20px;">
  <h1>{{.Title}}</h1>
  <p style="color:#555;">Slot: {{.Slot}} | CC: {{.CC}} | Domain: {{.D}} | LID: {{.LID}} | PID: {{.PID}} | Size: {{.TSize}}</p>
  <p style="color:#777;">Referrer: {{.KwRf}}</p>
  <p style="color:#777;">Page Title: {{.PTitle}}</p>
  <p style="color:#777;">Page URL: {{.RURL}}</p>
  <p style="color:#777;">Keyword ID: {{.KID}}</p>
  {{if .IsBot}}
    <div style="margin:10px 0; padding:10px; background:#fff7ed; border:1px solid #fed7aa; color:#9a3412; border-radius:6px;">Bot detected: ad clicks are disabled.</div>
  {{end}}
  <hr>
  <div class="sponsored-ads" style="margin-top:16px;">
    {{if .HasAds}}
      {{range .Ads}}
        <div class="ad-item" style="border:1px solid #e5e7eb; border-radius:8px; padding:12px; margin-bottom:10px;">
          {{if .RenderLinks}}
            <a href="{{.ClickHref}}" rel="nofollow noopener" target="_blank" style="font-weight:600; color:#0b57d0; text-decoration:none;">{{.TitleHTML}}</a>
          {{else}}
            <span style="font-weight:600; color:#0b57d0; text-decoration:none; cursor:not-allowed;">{{.TitleHTML}}</span>
          {{end}}
          {{if .DescHTML}}<div class="ad-desc" style="color:#374151; margin-top:6px;">{{.DescHTML}}</div>{{end}}
          {{if .Host}}<div class="ad-host" style="color:#6b7280; margin-top:6px; font-size:13px;">{{.Host}}</div>{{end}}
        </div>
      {{end}}
    {{else}}
      <div style="color:#999;">No sponsored ads available</div>
    {{end}}
  </div>
  <p><a href="javascript:history.back()">Back</a></p>
</body></html>`

// FirstCallJS is the JavaScript for the first call endpoint
// Uses pub parameter or host-based configuration
const FirstCallJS = `(function(){
  var src = (document.currentScript && document.currentScript.src) || '';
  var ORIGIN;
  var PUB_PARAM = ''; // Publisher identifier from script src
  try { 
    var srcUrl = new URL(src, location.href);
    ORIGIN = srcUrl.origin;
    PUB_PARAM = srcUrl.searchParams.get('pub') || '';
  } catch(e){ ORIGIN = location.origin; }

  // Publisher configuration by pub param or host
  var PUB_CONFIG = {
    'blue': { pid: 100, lid: 224, actno: 5, maxno: 5, cc: 'US', tsize: '300x250', pubKey: 'blue' },
    'red':  { pid: 200, lid: 224, actno: 5, maxno: 5, cc: 'US', tsize: '300x250', pubKey: 'red' }
  };
  var HOST_CONFIG = {
    'blue.localhost': PUB_CONFIG['blue'],
    'red.localhost':  PUB_CONFIG['red'],
    'localhost':      { pid: 0, lid: 224, actno: 5, maxno: 5, cc: 'US', tsize: '300x250', pubKey: 'default' }
  };
  var DEFAULT_CONFIG = { pid: 0, lid: 224, actno: 5, maxno: 5, cc: 'US', tsize: '300x250', pubKey: 'default' };

  function getConfig() {
    // First check pub param from script src
    if (PUB_PARAM && PUB_CONFIG[PUB_PARAM]) {
      return PUB_CONFIG[PUB_PARAM];
    }
    // Fallback to host-based config
    var host = (location.hostname || '').replace(/^www\./, '').split(':')[0];
    if (HOST_CONFIG[host]) return HOST_CONFIG[host];
    for (var h in HOST_CONFIG) {
      if (host === h || host.indexOf(h) === 0) return HOST_CONFIG[h];
    }
    return DEFAULT_CONFIG;
  }

  function getPageDefaults(){
    var config = getConfig();
    var qs = new URLSearchParams(window.location.search);
    return {
      actno: config.actno,
      maxno: config.maxno,
      cc: config.cc,
      lid: config.lid,
      pid: config.pid,
      pub: config.pubKey, // Pass publisher key to backend
      d: location.hostname || '',
      rurl: location.href,
      ptitle: qs.get('ptitle') || (document.title || ''),
      tsize: config.tsize,
      kwrf: document.referrer || ''
    };
  }

  function slotIdFromEl(el){
    if (!el) return '';
    var dataSlot = el.getAttribute('data-kw-slot');
    if (dataSlot) return String(dataSlot);
    var id = el.id || '';
    if (id) return id;
    return '';
  }

  function injectForSlot(el){
    if (!el || el.__kwInjected) return;
    var defaults = getPageDefaults();
    var slotId = slotIdFromEl(el);
    if (!slotId) return;
    var p = [];
    p.push('slot=' + encodeURIComponent(slotId));
    p.push('actno=' + encodeURIComponent(String(defaults.actno)));
    p.push('maxno=' + encodeURIComponent(String(defaults.maxno)));
    p.push('cc=' + encodeURIComponent(String(defaults.cc)));
    p.push('lid=' + encodeURIComponent(String(defaults.lid)));
    if (defaults.d) p.push('d=' + encodeURIComponent(String(defaults.d)));
    if (defaults.rurl) p.push('rurl=' + encodeURIComponent(String(defaults.rurl)));
    if (defaults.ptitle) p.push('ptitle=' + encodeURIComponent(String(defaults.ptitle)));
    if (defaults.tsize) p.push('tsize=' + encodeURIComponent(String(defaults.tsize)));
    if (defaults.kwrf) p.push('kwrf=' + encodeURIComponent(String(defaults.kwrf)));
    p.push('pid=' + encodeURIComponent(String(defaults.pid))); // Always send PID (even if 0)
    if (defaults.pub) p.push('pub=' + encodeURIComponent(String(defaults.pub))); // Publisher key
    var s = document.createElement('script');
    s.async = true;
    s.src = ORIGIN + '/render.js?' + p.join('&');
    (document.head || document.documentElement || document.body).appendChild(s);
    el.__kwInjected = true;
  }

  function findSlots(){
    var list = [];
    try {
      var a = document.querySelectorAll('[data-kw-slot]');
      for (var i=0; i<a.length; i++) list.push(a[i]);
      var b = document.querySelectorAll('[id^="kw-slot-"]');
      for (var j=0; j<b.length; j++) { if (list.indexOf(b[j]) === -1) list.push(b[j]); }
    } catch(e){}
    return list;
  }

  function run(){ var slots = findSlots(); for (var i=0; i<slots.length; i++){ injectForSlot(slots[i]); } }
  if (document.readyState === 'complete' || document.readyState === 'interactive') { run(); }
  else { document.addEventListener('DOMContentLoaded', run); }
})();`
