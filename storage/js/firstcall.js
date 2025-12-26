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

  var PID = scriptEl ? (scriptEl.getAttribute('data-pid') || '') : '';
  var ORIGIN = location.origin;

  if (scriptEl && scriptEl.src) {
    try {
      var srcUrl = new URL(scriptEl.src, location.href);
      ORIGIN = srcUrl.origin;
      if (!PID) PID = srcUrl.searchParams.get('pid') || '';
    } catch(e) {}
  }

  var CONFIG = { pid: PID ? parseInt(PID, 10) : 0, cc: 'US', tsize: '300x250', lid: '224' };

  function slotIdFromEl(el) {
    if (!el) return '';
    return el.getAttribute('data-kw-slot') || el.id || '';
  }

  function injectForSlot(el) {
    if (!el || el.__kwInjected) return;
    var slotId = slotIdFromEl(el);
    if (!slotId) return;

    var p = 'slot=' + encodeURIComponent(slotId) +
            '&cc=' + encodeURIComponent(CONFIG.cc) +
            '&pid=' + encodeURIComponent(CONFIG.pid) +
            '&tsize=' + encodeURIComponent(CONFIG.tsize) +
            '&lid=' + encodeURIComponent(CONFIG.lid) +
            '&d=' + encodeURIComponent(location.hostname) +
            '&ptitle=' + encodeURIComponent(document.title || '') +
            '&rurl=' + encodeURIComponent(location.href) +
            '&kwrf=' + encodeURIComponent(document.referrer || '');

    var dims = CONFIG.tsize.split('x');
    var iframe = document.createElement('iframe');
    iframe.src = ORIGIN + '/keyword_render?' + p;
    iframe.width = dims[0] || '300';
    iframe.height = dims[1] || '250';
    iframe.style.border = 'none';
    iframe.scrolling = 'no';
    iframe.frameBorder = '0';
    el.appendChild(iframe);
    el.__kwInjected = true;
  }

  function findSlots() {
    var list = [];
    var a = document.querySelectorAll('[data-kw-slot]');
    for (var i = 0; i < a.length; i++) list.push(a[i]);
    var b = document.querySelectorAll('[id^="kw-slot-"]');
    for (var j = 0; j < b.length; j++) {
      if (list.indexOf(b[j]) === -1) list.push(b[j]);
    }
    return list;
  }

  window.addEventListener('message', function(e) {
    if (e.data && e.data.type === 'impression' && e.data.url) {
      var i = new Image();
      i.src = e.data.url;
    }
  });

  function run() {
    var slots = findSlots();
    for (var i = 0; i < slots.length; i++) injectForSlot(slots[i]);
  }

  if (document.readyState === 'complete' || document.readyState === 'interactive') run();
  else document.addEventListener('DOMContentLoaded', run);
})();
