(function () {
  'use strict';

  // A password box not to be saved (input.Props.NotSaved) is a text box masked with -webkit-text-security; a browser
  // that cannot mask one gets a password box back
  function unmaskable() {
    if (CSS.supports('-webkit-text-security', 'disc')) return;
    document.querySelectorAll('[data-tui-input-masked]').forEach((input) => {
      input.type = 'password';
    });
  }
  unmaskable();
  document.addEventListener('DOMContentLoaded', unmaskable);
  document.addEventListener('htmx:afterSettle', unmaskable);

  document.addEventListener('click', (e) => {
    const button = e.target.closest('[data-tui-input-toggle-password]');
    if (!button) return;

    const inputId = button.getAttribute('data-tui-input-toggle-password');
    const input = document.getElementById(inputId);
    if (!input) return;

    const iconOpen = button.querySelector('.icon-open');
    const iconClosed = button.querySelector('.icon-closed');

    // A masked text box is shown by unmasking it, as making it a password box would have it offered to be saved
    const masked = input.hasAttribute('data-tui-input-masked') && input.type !== 'password';
    const hide = masked ? input.style.webkitTextSecurity === 'none' : input.type !== 'password';
    if (masked) {
      input.style.webkitTextSecurity = hide ? 'disc' : 'none';
    } else {
      input.type = hide ? 'password' : 'text';
    }
    if (iconOpen) iconOpen.classList.toggle('hidden', !hide);
    if (iconClosed) iconClosed.classList.toggle('hidden', hide);
  });
})();
