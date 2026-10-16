'use strict';

// Get default accent colors
{{ $defaultDarkAccent   := .Site.Params.Style.darkAccent  | default .Site.Data.default.style.darkAccent }}
{{ $defaultLightAccent  := .Site.Params.Style.lightAccent | default .Site.Data.default.style.lightAccent }}

// Get CSS transition
{{ $changeTransition := .Site.Params.Style.changeTransition | default .Site.Data.default.style.changeTransition }}

// =================================================
// Mode switcher + Custom accent color
// Based on: https://gist.github.com/regpaq/04c67e8aceecbf0fd819945835412d1f
// =================================================

const rootElement = document.documentElement;

// localStorage can throw (privacy modes, blocked site data),
// which would stop the whole script
const storage = {
  get(key) {
    try {
      return localStorage.getItem(key);
    } catch (e) {
      return null;
    }
  },
  set(key, value) {
    try {
      localStorage.setItem(key, value);
    } catch (e) {}
  },
  remove(key) {
    try {
      localStorage.removeItem(key);
    } catch (e) {}
  }
};

// Set the dark mode
function setDark() {
  rootElement.setAttribute('data-mode', 'dark');
}

// Set the light mode
function setLight() {
  rootElement.setAttribute('data-mode', 'light');
}

/*
 * Initialization triggers dark/light mode based on 3 things
 * The priority follows:
 * 
 * 1. Local preference (localStorage)
 * 2. System settings (prefers-color-scheme)
 * 3. HTML data-* attribute (data-mode)
 */

const localMode = storage.get('mode');

{{ if not .Site.Params.Style.ignoreSystemSettings }}
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)');
  const prefersLight = window.matchMedia('(prefers-color-scheme: light)');
{{ end }}

if (localMode === 'dark') {
  setDark();
} else if (localMode === 'light') {
  setLight();

  {{ if not .Site.Params.Style.ignoreSystemSettings }}
    } else if (prefersDark.matches) {
      setDark();
    } else if (prefersLight.matches) {
      setLight();
  {{ end }}

}


{{ if hugo.IsDevelopment }}

  function capitalize(string) {
    return string.charAt(0).toUpperCase() + string.slice(1);
  }

  const currentModeTitle = capitalize(rootElement.getAttribute('data-mode'));

  if (localMode !== null) {
    console.log(currentModeTitle +
    ' mode loaded via local preference (localStorage).');
  } else if (typeof prefersDark !== 'undefined') {
    console.log(currentModeTitle +
    ' mode loaded via system settings (prefers-color-scheme).');
  } else {
    console.log(currentModeTitle +
    ' mode loaded via HTML data-* attribute (data-mode).');
  }
  
{{ end }}


function getAccent() {
  
  const currentMode = rootElement.getAttribute('data-mode');
  
  const localDarkAccent = storage.get('darkAccent');
  const localLightAccent = storage.get('lightAccent');
  
  let currentAccent;

  if (currentMode === 'dark') {
  
    if (localDarkAccent === null) {
      currentAccent = '{{ $defaultDarkAccent }}';
    } else {
      currentAccent = localDarkAccent;
    }
    
  } else {

    if (localLightAccent === null) {
      currentAccent = '{{ $defaultLightAccent }}';
    } else {
      currentAccent = localLightAccent;
    }
    
  }
  
  {{ if hugo.IsDevelopment }}
  
    if (
      (currentMode === 'dark') &&
      (storage.get('darkAccent') !== null) ||
      (currentMode === 'light') &&
      (storage.get('lightAccent') !== null)
    ) {
      console.log('Custom accent color defined. Loading custom ' +
      currentMode + ' accent (' + currentAccent + ').');
    } else {
      console.log('Custom accent color NOT defined. Loading default ' +
      currentMode + ' accent (' + currentAccent + ').');
    }
    
  {{ end }}
  
  return currentAccent;
}

const activeAccent = getAccent();

// Set the active accent color for these right after setting dark/light mode
// Should mitigate any flashing/flickering
const rootStyle = rootElement.style;

rootStyle.setProperty('--accent', activeAccent);


// Also meta-theme cuz, why not
const metaThemeColor = document.querySelector('meta[name=theme-color]');

metaThemeColor.setAttribute('content', activeAccent);



document.addEventListener('DOMContentLoaded', function() {

  const picker = document.querySelector('.mode-picker');
  const accentToggle = picker.querySelector('.accent-toggle');
  const accentMenu = picker.querySelector('.accent-menu');
  const customInput = accentMenu.querySelector('input[type=color]');
  const swatches = accentMenu.querySelectorAll('.swatch[data-color]');

  function accentKey() {
    return rootElement.getAttribute('data-mode') === 'dark' ?
      'darkAccent' : 'lightAccent';
  }

  // Apply the active accent color and mark it in the menu
  function updateAccent() {
    const activeAccent = getAccent();

    rootStyle.setProperty('--accent', activeAccent);
    metaThemeColor.setAttribute('content', activeAccent);

    const accent = activeAccent.toLowerCase();

    swatches.forEach(function(swatch) {
      swatch.setAttribute('aria-pressed', swatch.dataset.color === accent);
    });

    customInput.value = accent;
  }

  function saveAccent(color) {
    storage.set(accentKey(), color.toLowerCase());
    updateAccent();
  }

  function openMenu() {
    accentMenu.hidden = false;
    accentToggle.setAttribute('aria-expanded', 'true');

    const selected = accentMenu.querySelector('[aria-pressed=true]');
    (selected || swatches[0] || customInput).focus();
  }

  function closeMenu(returnFocus) {
    if (accentMenu.hidden) {
      return;
    }

    accentMenu.hidden = true;
    accentToggle.setAttribute('aria-expanded', 'false');

    if (returnFocus) {
      accentToggle.focus();
    }
  }

  accentToggle.addEventListener('click', function() {
    if (accentMenu.hidden) {
      openMenu();
    } else {
      closeMenu(false);
    }
  });

  swatches.forEach(function(swatch) {
    swatch.addEventListener('click', function() {
      saveAccent(swatch.dataset.color);
    });
  });

  // Live preview while picking, saved once the picker is closed
  customInput.addEventListener('input', function() {
    rootStyle.setProperty('--accent', customInput.value);
  });

  customInput.addEventListener('change', function() {
    saveAccent(customInput.value);
  });

  accentMenu.querySelector('.accent-reset')
    .addEventListener('click', function() {
      storage.remove(accentKey());
      updateAccent();
    });

  // Close with Escape or by clicking anywhere else
  document.addEventListener('keydown', function(event) {
    if (event.key === 'Escape') {
      closeMenu(picker.contains(document.activeElement));
    }
  });

  document.addEventListener('pointerdown', function(event) {
    if (!picker.contains(event.target)) {
      closeMenu(false);
    }
  });

  updateAccent();

  // Smooth transition, only when changing modes (and not loading pages)
  function smoothTransition() {
    document.body.style.transition =
    '{{ printf "background-color %[1]s, color %[1]s" $changeTransition }}';
  }
  
  // Change mode via localStorage
  function localModeChange() {
  
    smoothTransition();

    if (rootElement.getAttribute('data-mode') === 'light') {
      setDark();
      storage.set('mode', 'dark');
    } else {
      setLight();
      storage.set('mode', 'light');
    }
    
    {{ if hugo.IsDevelopment }}
      console.log('Local: ' +
      capitalize(storage.get('mode') + ' mode set.'));
    {{ end }}
    
    updateAccent();
  }

  
  {{ if hugo.IsDevelopment }}
  
    // TEST
    // Keyboard shortcut for mode change, here for testing purposes only
    // CTRL + ALT + M
    document.addEventListener('keydown', (event) => {
      const e = event || window.event;
      if (e.keyCode === 77 && e.ctrlKey && e.altKey) {
        localModeChange();
        return;
      }
    }, false);
    
  {{ end }}


  {{ if not .Site.Params.Style.ignoreSystemSettings }}
  
    // Change mode via system settings
    function systemModeChange() {
    
      smoothTransition();
      
      if (prefersDark.matches) {
        setDark();
      } else {
        setLight();
      }
      
      {{ if hugo.IsDevelopment }}
        console.log('System: ' +
        capitalize(rootElement.getAttribute('data-mode')) + ' mode set.');
      {{ end }}
      
      updateAccent();
      
      // System settings do not require localStorage
      // (the stored mode may have been set after the page loaded)
      storage.remove('mode');

    }

    // System settings listener
    prefersDark.addEventListener('change', systemModeChange);
  
  {{ end }}

  // Mode change button listener
  picker.querySelector('.mode-toggle')
    .addEventListener('click', localModeChange);
});
