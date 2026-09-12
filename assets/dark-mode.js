/* Dark / light mode toggle for the docs site.
   Injects a nav item into the theme navbar (the theme itself is untouched),
   persists the choice in localStorage, and defaults to the OS preference. */
(function () {
  'use strict';

  var STORAGE_KEY = 'jenkins-desktop-theme';

  function stored() {
    try {
      var value = window.localStorage.getItem(STORAGE_KEY);
      if (value === 'dark' || value === 'light') {
        return value;
      }
    } catch (e) {
      /* storage unavailable (e.g. file://), fall through to OS default */
    }
    return null;
  }

  function osDefault() {
    if (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches) {
      return 'dark';
    }
    return 'light';
  }

  function apply(theme) {
    if (theme === 'dark' || theme === 'light') {
      /* Explicit choice: pin it so it beats the OS media query. */
      document.documentElement.setAttribute('data-theme', theme);
    } else {
      /* No stored choice: leave the attribute off so the
         prefers-color-scheme media query decides from first paint. */
      document.documentElement.removeAttribute('data-theme');
      theme = osDefault();
    }
    var icon = document.querySelector('#jd-theme-toggle i');
    if (icon) {
      icon.className = theme === 'dark' ? 'fas fa-sun' : 'fas fa-moon';
    }
    var link = document.querySelector('#jd-theme-toggle a');
    if (link) {
      link.setAttribute('title', theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode');
    }
  }

  function toggle() {
    var current = document.documentElement.getAttribute('data-theme') || osDefault();
    var next = current === 'dark' ? 'light' : 'dark';
    try {
      window.localStorage.setItem(STORAGE_KEY, next);
    } catch (e) {
      /* ignore */
    }
    apply(next);
  }

  /* Set the theme as early as possible to avoid a light-mode flash. */
  apply(stored());

  document.addEventListener('DOMContentLoaded', function () {
    if (document.getElementById('jd-theme-toggle')) {
      return;
    }
    var bar = document.querySelector('ul.navbar-right');
    if (!bar) {
      return;
    }
    var li = document.createElement('li');
    li.id = 'jd-theme-toggle';
    var a = document.createElement('a');
    a.href = '#';
    a.addEventListener('click', function (e) {
      e.preventDefault();
      toggle();
    });
    var icon = document.createElement('i');
    icon.className = 'fas fa-moon';
    a.appendChild(icon);
    a.appendChild(document.createTextNode(' Theme'));
    li.appendChild(a);
    bar.appendChild(li);
    apply(stored());
  });
})();
