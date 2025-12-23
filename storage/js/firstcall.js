(function(){
  var scriptEl = document.currentScript;

  if (!scriptEl) {
    var scripts = document.getElementsByTagName('script');
    for (var i = 0; i < scripts.length; i++) {
      if (scripts[i].src && scripts[i].src.indexOf('firstcall.js') !== -1) {
        scriptEl = scripts[i];
        break;
      }
    }
  }

  // Get PID from data-pid attribute or URL param
  var PID = scriptEl ? (scriptEl.getAttribute('data-pid') || '') : '';
  var ORIGIN = location.origin;

  // Always extract ORIGIN from the script's src URL (if available)
  if (scriptEl && scriptEl.src) {
    try {
      var srcUrl = new URL(scriptEl.src, location.href);
      ORIGIN = srcUrl.origin;
      // Extract pid from query string if not already set via data-pid
      if (!PID) {
        PID = srcUrl.searchParams.get('pid') || '';
      }
    } catch(e) {}
  }

  // Default config - actual config (maxno, max_ads) is determined by templates on the server
  var DEFAULT_CONFIG = { pid: 0, lid: 224, cc: 'US', tsize: '300x250' };

  function getConfig() {
    var config = Object.assign({}, DEFAULT_CONFIG);
    if (PID) {
      config.pid = parseInt(PID, 10) || 0;
    }
    return config;
  }

  function getPageDefaults(){
    var config = getConfig();
    var qs = new URLSearchParams(window.location.search);
    return {
      cc: config.cc,
      lid: config.lid,
      pid: config.pid,
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
    p.push('cc=' + encodeURIComponent(String(defaults.cc)));
    p.push('lid=' + encodeURIComponent(String(defaults.lid)));
    if (defaults.d) p.push('d=' + encodeURIComponent(String(defaults.d)));
    if (defaults.rurl) p.push('rurl=' + encodeURIComponent(String(defaults.rurl)));
    if (defaults.ptitle) p.push('ptitle=' + encodeURIComponent(String(defaults.ptitle)));
    if (defaults.tsize) p.push('tsize=' + encodeURIComponent(String(defaults.tsize)));
    if (defaults.kwrf) p.push('kwrf=' + encodeURIComponent(String(defaults.kwrf)));
    p.push('pid=' + encodeURIComponent(String(defaults.pid))); // Always send PID
    var s = document.createElement('script');
    s.async = true;
    s.src = ORIGIN + '/keyword_render?' + p.join('&');   // todo change path here
    console.log('[firstcall.js] Calling:', s.src);  // Debug: shows the full URL being called
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
})();// todo-> what problem this fucntion is solving
