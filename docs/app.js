(function () {
  'use strict';

  var reduceMotion = window.matchMedia('(prefers-reduced-motion: reduce)');
  var finePointer = window.matchMedia('(hover: hover) and (pointer: fine)');

  initCopyButtons();
  initReveals();
  initDashboard();
  initTilt();
  initFilm();

  function initCopyButtons() {
    var status = document.getElementById('copy-status');

    document.querySelectorAll('[data-copy]').forEach(function (button) {
      button.addEventListener('click', function () {
        var source = document.getElementById(button.getAttribute('data-copy'));
        if (!source) {
          return;
        }

        var text = source.textContent.replace(/\s+$/, '');
        copyText(text).then(
          function () {
            button.textContent = 'Copied';
            announce(status, 'Command copied to the clipboard.');
            window.setTimeout(function () {
              button.textContent = 'Copy';
            }, 1600);
          },
          function () {
            announce(status, 'Could not copy the command. Select it manually.');
          }
        );
      });
    });
  }

  function copyText(text) {
    if (navigator.clipboard && window.isSecureContext) {
      return navigator.clipboard.writeText(text);
    }

    return new Promise(function (resolve, reject) {
      var field = document.createElement('textarea');
      field.value = text;
      field.setAttribute('readonly', '');
      field.style.position = 'fixed';
      field.style.top = '-1000px';
      document.body.appendChild(field);
      field.select();

      var copied = false;
      try {
        copied = document.execCommand('copy');
      } catch (err) {
        copied = false;
      }
      document.body.removeChild(field);

      if (copied) {
        resolve();
      } else {
        reject(new Error('copy command failed'));
      }
    });
  }

  function announce(status, message) {
    if (status) {
      status.textContent = message;
    }
  }

  function initReveals() {
    var items = document.querySelectorAll('[data-reveal]');
    if (!items.length) {
      return;
    }

    if (reduceMotion.matches || !('IntersectionObserver' in window)) {
      items.forEach(function (el) {
        el.classList.add('is-revealed');
      });
      return;
    }

    var seen = new Map();
    items.forEach(function (el) {
      var index = seen.get(el.parentElement) || 0;
      el.style.setProperty('--reveal-index', String(index));
      seen.set(el.parentElement, index + 1);
    });

    var observer = new IntersectionObserver(
      function (entries) {
        entries.forEach(function (entry) {
          if (!entry.isIntersecting) {
            return;
          }
          entry.target.classList.add('is-revealed');
          observer.unobserve(entry.target);
        });
      },
      { rootMargin: '0px 0px -8% 0px', threshold: 0.12 }
    );

    items.forEach(function (el) {
      observer.observe(el);
    });
  }

  function initDashboard() {
    var dash = document.querySelector('.dash');
    if (!dash) {
      return;
    }

    var numbers = dash.querySelectorAll('[data-count]');
    var bars = dash.querySelectorAll('.bar-fill[data-progress]');

    if (reduceMotion.matches) {
      numbers.forEach(finalNumber);
      bars.forEach(setBar);
      return;
    }

    numbers.forEach(function (el) {
      el.textContent = formatNumber(0, Number(el.dataset.countDecimals || 0), el.dataset.countSuffix || '');
    });

    var started = false;

    function start() {
      if (started) {
        return;
      }
      started = true;
      numbers.forEach(animateNumber);
      bars.forEach(setBar);
    }

    if (!('IntersectionObserver' in window)) {
      start();
      return;
    }

    var observer = new IntersectionObserver(
      function (entries) {
        if (!entries.some(function (entry) { return entry.isIntersecting; })) {
          return;
        }
        start();
        observer.disconnect();
      },
      { threshold: 0.15 }
    );

    observer.observe(dash);
  }

  function formatNumber(value, decimals, suffix) {
    var text = decimals > 0 ? value.toFixed(decimals) : String(Math.round(value));
    return text + suffix;
  }

  function finalNumber(el) {
    var decimals = Number(el.dataset.countDecimals || 0);
    el.textContent = formatNumber(
      Number(el.dataset.count),
      decimals,
      el.dataset.countSuffix || ''
    );
  }

  function animateNumber(el) {
    var target = Number(el.dataset.count);
    var decimals = Number(el.dataset.countDecimals || 0);
    var suffix = el.dataset.countSuffix || '';
    var duration = 1100;
    var startTime = null;

    function frame(now) {
      if (startTime === null) {
        startTime = now;
      }
      var progress = Math.min((now - startTime) / duration, 1);
      var eased = 1 - Math.pow(1 - progress, 3);
      el.textContent = formatNumber(target * eased, decimals, suffix);
      if (progress < 1) {
        requestAnimationFrame(frame);
      }
    }

    requestAnimationFrame(frame);
  }

  function setBar(el) {
    requestAnimationFrame(function () {
      el.style.width = el.dataset.progress + '%';
    });
  }

  function initTilt() {
    var zone = document.querySelector('.hero-object');
    var panel = document.querySelector('.dash-figure');
    if (!zone || !panel || reduceMotion.matches || !finePointer.matches) {
      return;
    }

    var maxRotateX = 2.4;
    var maxRotateY = 3.4;
    var targetX = 0;
    var targetY = 0;
    var currentX = 0;
    var currentY = 0;
    var frameId = 0;

    function step() {
      currentX += (targetX - currentX) * 0.12;
      currentY += (targetY - currentY) * 0.12;
      panel.style.transform =
        'perspective(1200px) rotateX(' + currentX.toFixed(2) + 'deg) rotateY(' + currentY.toFixed(2) + 'deg)';

      if (Math.abs(targetX - currentX) > 0.02 || Math.abs(targetY - currentY) > 0.02) {
        frameId = requestAnimationFrame(step);
      } else {
        frameId = 0;
      }
    }

    function requestStep() {
      if (!frameId) {
        frameId = requestAnimationFrame(step);
      }
    }

    zone.addEventListener('pointermove', function (event) {
      var rect = zone.getBoundingClientRect();
      var offsetX = (event.clientX - rect.left) / rect.width - 0.5;
      var offsetY = (event.clientY - rect.top) / rect.height - 0.5;
      targetY = offsetX * maxRotateY * 2;
      targetX = -offsetY * maxRotateX * 2;
      requestStep();
    });

    zone.addEventListener('pointerleave', function () {
      targetX = 0;
      targetY = 0;
      requestStep();
    });

    if (typeof reduceMotion.addEventListener === 'function') {
      reduceMotion.addEventListener('change', function (event) {
        if (event.matches) {
          targetX = 0;
          targetY = 0;
          requestStep();
        }
      });
    }
  }

  function initFilm() {
    var video = document.getElementById('brag-film');
    var frame = document.querySelector('.film-frame');
    var button = document.querySelector('[data-film-play]');
    if (!video || !frame || !button) {
      return;
    }

    button.addEventListener('click', function () {
      var played = video.play();
      if (played && typeof played.catch === 'function') {
        played.catch(function () {
          frame.classList.remove('is-playing');
        });
      }
    });

    video.addEventListener('play', function () {
      frame.classList.add('is-playing');
      button.blur();
    });

    video.addEventListener('pause', function () {
      frame.classList.remove('is-playing');
    });

    video.addEventListener('ended', function () {
      frame.classList.remove('is-playing');
    });
  }
})();
