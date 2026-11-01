(function () {
  'use strict';

  function getEventTargetElement(event) {
    return event.target instanceof Element ? event.target : null;
  }

  // Utility functions
  function getSlots(container) {
    return Array.from(container.querySelectorAll('[data-tui-inputotp-slot]')).sort(
      (a, b) => parseInt(a.getAttribute('data-tui-inputotp-index')) - parseInt(b.getAttribute('data-tui-inputotp-index'))
    );
  }

  function focusSlot(slot) {
    if (!slot) return;
    slot.focus();
    setTimeout(() => slot.select(), 0);
  }

  function updateHiddenValue(container) {
    const hiddenInput = container.querySelector('[data-tui-inputotp-value-target]');
    const slots = getSlots(container);
    if (hiddenInput && slots.length) {
      hiddenInput.value = slots.map(s => s.value).join('');
    }
  }

  function findFirstEmptySlot(container) {
    const slots = getSlots(container);
    for (const slot of slots) {
      if (!slot.value) return slot;
    }
    return null;
  }

  function getNextSlot(container, currentSlot) {
    const slots = getSlots(container);
    const index = slots.indexOf(currentSlot);
    return index >= 0 && index < slots.length - 1 ? slots[index + 1] : null;
  }

  function getPrevSlot(container, currentSlot) {
    const slots = getSlots(container);
    const index = slots.indexOf(currentSlot);
    return index > 0 ? slots[index - 1] : null;
  }

  // fillSlots puts a code's digits in the slots from slot on, as when one is pasted or autofilled, and focuses the
  // next empty slot, or the last one filled
  function fillSlots(container, slot, digits) {
    const slots = getSlots(container);
    const startIndex = slots.indexOf(slot);
    for (let i = 0; i < digits.length && startIndex + i < slots.length; i++) {
      slots[startIndex + i].value = digits[i];
    }
    updateHiddenValue(container);
    focusSlot(findFirstEmptySlot(container) || slots[Math.min(startIndex + digits.length, slots.length - 1)]);
  }

  // Event handlers
  document.addEventListener('input', (e) => {
    const target = getEventTargetElement(e);
    if (!target?.matches('[data-tui-inputotp-slot]')) return;

    const slot = target;
    const container = slot.closest('[data-tui-inputotp]');
    if (!container) return;

    // A code is digits only, so anything else that gets in is dropped
    const digits = slot.value.replace(/\D/g, '');
    if (!digits) {
      slot.value = '';
      updateHiddenValue(container);
      return;
    }

    // More digits than one keystroke adds is a whole code autofilled, which fills the slots from this one on
    if (digits.length > 2) {
      fillSlots(container, slot, digits);
      return;
    }

    // Keep only last digit
    slot.value = digits.slice(-1);

    // Move to next slot if filled
    if (slot.value) {
      const nextSlot = getNextSlot(container, slot);
      if (nextSlot) focusSlot(nextSlot);
    }

    updateHiddenValue(container);
  });

  // Typing anything but a digit does nothing, so it neither shows nor moves on to the next slot
  document.addEventListener('beforeinput', (e) => {
    const target = getEventTargetElement(e);
    if (target?.matches('[data-tui-inputotp-slot]') && e.data && !/\d/.test(e.data)) e.preventDefault();
  });

  document.addEventListener('keydown', (e) => {
    const target = getEventTargetElement(e);
    if (!target?.matches('[data-tui-inputotp-slot]')) return;

    const slot = target;
    const container = slot.closest('[data-tui-inputotp]');
    if (!container) return;

    if (e.key === 'Backspace') {
      e.preventDefault();

      if (slot.value) {
        slot.value = '';
        updateHiddenValue(container);
      } else {
        const prevSlot = getPrevSlot(container, slot);
        if (prevSlot) {
          prevSlot.value = '';
          updateHiddenValue(container);
          focusSlot(prevSlot);
        }
      }
    } else if (e.key === 'ArrowLeft') {
      e.preventDefault();
      const prevSlot = getPrevSlot(container, slot);
      if (prevSlot) focusSlot(prevSlot);
    } else if (e.key === 'ArrowRight') {
      e.preventDefault();
      const nextSlot = getNextSlot(container, slot);
      if (nextSlot) focusSlot(nextSlot);
    }
  });

  document.addEventListener('focus', (e) => {
    const target = getEventTargetElement(e);
    if (!target?.matches('[data-tui-inputotp-slot]')) return;

    const slot = target;
    const container = slot.closest('[data-tui-inputotp]');
    if (!container) return;

    // Redirect to first empty slot
    const firstEmpty = findFirstEmptySlot(container);
    if (firstEmpty && firstEmpty !== slot) {
      focusSlot(firstEmpty);
      return;
    }

    setTimeout(() => slot.select(), 0);
  }, true);

  document.addEventListener('paste', (e) => {
    const target = getEventTargetElement(e);
    const slot = target?.closest('[data-tui-inputotp-slot]');
    if (!slot) return;

    e.preventDefault();
    const container = slot.closest('[data-tui-inputotp]');
    if (!container) return;

    const pastedData = (e.clipboardData || window.clipboardData).getData('text');
    fillSlots(container, slot, pastedData.replace(/\D/g, ''));
  });

  // Label click handling
  document.addEventListener('click', (e) => {
    const target = getEventTargetElement(e);
    if (!target?.matches('label[for]')) return;

    const targetId = target.getAttribute('for');
    const hiddenInput = document.getElementById(targetId);
    if (!hiddenInput?.matches('[data-tui-inputotp-value-target]')) return;

    e.preventDefault();
    const container = hiddenInput.closest('[data-tui-inputotp]');
    const slots = getSlots(container);
    if (slots.length > 0) focusSlot(slots[0]);
  });

  // Form reset
  document.addEventListener('reset', (e) => {
    const target = getEventTargetElement(e);
    if (!target?.matches('form')) return;

    target.querySelectorAll('[data-tui-inputotp]').forEach(container => {
      getSlots(container).forEach(slot => {
        slot.value = '';
      });
      updateHiddenValue(container);
    });
  });

  // MutationObserver for initial setup and autofocus
  new MutationObserver(() => {
    document.querySelectorAll('[data-tui-inputotp]').forEach(container => {
      const slots = getSlots(container);
      if (slots.length === 0) return;

      // Set initial value if provided
      const initialValue = container.getAttribute('data-tui-inputotp-value');
      if (initialValue && !slots[0].value) {
        for (let i = 0; i < slots.length && i < initialValue.length; i++) {
          if (!slots[i].value) slots[i].value = initialValue[i];
        }
        updateHiddenValue(container);
      }

      // Handle autofocus
      if (container.hasAttribute('autofocus') && !slots.some(s => s === document.activeElement)) {
        requestAnimationFrame(() => {
          if (slots[0] && !slots.some(s => s === document.activeElement)) {
            focusSlot(slots[0]);
          }
        });
      }
    });
  }).observe(document.body, { childList: true, subtree: true });
})();
