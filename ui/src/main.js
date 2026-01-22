import { invoke } from '@tauri-apps/api/core';

// State
let systems = [];
let selectedSystems = new Set();

// DOM elements
const systemList = document.getElementById('system-list');
const userStoreInput = document.getElementById('user-store');
const btnApply = document.getElementById('btn-apply');
const btnDoctor = document.getElementById('btn-doctor');
const btnStatus = document.getElementById('btn-status');
const outputSection = document.getElementById('output-section');
const logElement = document.getElementById('log');
const provisionsSection = document.getElementById('provisions-section');
const provisionsList = document.getElementById('provisions-list');

// Initialize
async function init() {
  try {
    systems = await invoke('get_systems');
    renderSystems();

    const config = await invoke('get_config');
    if (config.userStore) {
      userStoreInput.value = config.userStore;
    }
    if (config.systems) {
      for (const sysId of Object.keys(config.systems)) {
        selectedSystems.add(sysId);
      }
      renderSystems();
    }
  } catch (err) {
    log(`Error initializing: ${err}`);
  }
}

function renderSystems() {
  systemList.innerHTML = systems.map(sys => `
    <li>
      <label>
        <input type="checkbox"
               value="${sys.id}"
               ${selectedSystems.has(sys.id) ? 'checked' : ''}
               onchange="toggleSystem('${sys.id}', this.checked)">
        <strong>${sys.name}</strong>
        <small>${sys.description}</small>
      </label>
    </li>
  `).join('');
}

// Make toggleSystem available globally for inline handler
window.toggleSystem = function(sysId, checked) {
  if (checked) {
    selectedSystems.add(sysId);
  } else {
    selectedSystems.delete(sysId);
  }
};

function log(message) {
  outputSection.classList.remove('hidden');
  logElement.textContent += message + '\n';
  logElement.scrollTop = logElement.scrollHeight;
}

function clearLog() {
  logElement.textContent = '';
}

// Button handlers
btnApply.addEventListener('click', async () => {
  clearLog();
  btnApply.disabled = true;
  btnApply.setAttribute('aria-busy', 'true');

  try {
    // Save config first
    const systemsConfig = {};
    for (const sysId of selectedSystems) {
      const sys = systems.find(s => s.id === sysId);
      if (sys && sys.emulators.length > 0) {
        systemsConfig[sysId] = sys.emulators[0].id;
      }
    }

    await invoke('set_config', {
      userStore: userStoreInput.value,
      systems: systemsConfig
    });

    log('Applying configuration...');

    // Apply
    const result = await invoke('apply');
    for (const line of result) {
      log(line);
    }
    log('Done!');
  } catch (err) {
    log(`Error: ${err}`);
  } finally {
    btnApply.disabled = false;
    btnApply.removeAttribute('aria-busy');
  }
});

btnDoctor.addEventListener('click', async () => {
  clearLog();
  provisionsSection.classList.remove('hidden');

  try {
    const result = await invoke('doctor');
    provisionsList.innerHTML = '';

    for (const [sysId, provisions] of Object.entries(result)) {
      const sys = systems.find(s => s.id === sysId);
      const sysName = sys ? sys.name : sysId;

      let html = `<h3>${sysName}</h3>`;

      if (provisions.length === 0) {
        html += '<p>No provisions required.</p>';
      } else {
        html += '<ul>';
        for (const prov of provisions) {
          const statusClass = prov.status === 'found' ? 'status-ok' :
                              prov.required ? 'status-missing' : 'status-warn';
          const statusText = prov.status === 'found' ? 'OK' :
                             prov.required ? 'MISSING' : 'optional';
          html += `<li>
            <code>${prov.filename}</code>
            <span class="status-badge ${statusClass}">${statusText}</span>
            <small>${prov.description || ''}</small>
          </li>`;
        }
        html += '</ul>';
      }

      provisionsList.innerHTML += html;
    }
  } catch (err) {
    log(`Error: ${err}`);
  }
});

btnStatus.addEventListener('click', async () => {
  clearLog();

  try {
    const status = await invoke('status');
    log(`Emulation folder: ${status.userStore}`);
    log(`Enabled systems: ${status.enabledSystems.join(', ') || 'none'}`);
    log(`Installed emulators: ${status.installedEmulators.length}`);
    if (status.lastApplied) {
      log(`Last applied: ${new Date(status.lastApplied).toLocaleString()}`);
    }
  } catch (err) {
    log(`Error: ${err}`);
  }
});

// Start
init();
