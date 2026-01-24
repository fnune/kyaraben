import { invoke } from '@tauri-apps/api/core'

// State
let systems = []
const selectedSystems = new Set()

// DOM elements
const systemList = document.getElementById('system-list')
const userStoreInput = document.getElementById('user-store')
const btnApply = document.getElementById('btn-apply')
const btnDoctor = document.getElementById('btn-doctor')
const btnStatus = document.getElementById('btn-status')
const outputSection = document.getElementById('output-section')
const logElement = document.getElementById('log')
const provisionsSection = document.getElementById('provisions-section')
const provisionsList = document.getElementById('provisions-list')

// Install UI elements
const installSection = document.getElementById('install-section')
const installPrompt = document.getElementById('install-prompt')
const installStatus = document.getElementById('install-status')
const btnInstall = document.getElementById('btn-install')
const btnDismissInstall = document.getElementById('btn-dismiss-install')
const btnUninstall = document.getElementById('btn-uninstall')

// Initialize
async function init() {
  try {
    systems = await invoke('get_systems')
    renderSystems()

    const config = await invoke('get_config')
    if (config.userStore) {
      userStoreInput.value = config.userStore
    }
    if (config.systems) {
      for (const sysId of Object.keys(config.systems)) {
        selectedSystems.add(sysId)
      }
      renderSystems()
    }

    await checkInstallStatus()
  } catch (err) {
    log(`Error initializing: ${err}`)
  }
}

async function checkInstallStatus() {
  try {
    const status = await invoke('get_install_status')

    const dismissed = sessionStorage.getItem('install-dismissed')
    if (dismissed) return

    if (status.installed) {
      installSection.classList.remove('hidden')
      installPrompt.classList.add('hidden')
      btnInstall.classList.add('hidden')
      btnDismissInstall.classList.add('hidden')
      installStatus.classList.remove('hidden')
      installStatus.textContent = `Installed at: ${status.appPath}`
      btnUninstall.classList.remove('hidden')
    } else {
      installSection.classList.remove('hidden')
    }
  } catch {
    // Install features not available when not running as AppImage
  }
}

async function doInstall() {
  btnInstall.disabled = true
  btnInstall.setAttribute('aria-busy', 'true')

  try {
    await invoke('install_app')
    installPrompt.classList.add('hidden')
    btnInstall.classList.add('hidden')
    btnDismissInstall.classList.add('hidden')
    installStatus.classList.remove('hidden')
    installStatus.innerHTML =
      '<strong>Installed!</strong> Kyaraben is now in your applications menu.'
    btnUninstall.classList.remove('hidden')
  } catch (err) {
    installStatus.classList.remove('hidden')
    installStatus.textContent = `Install failed: ${err}`
  } finally {
    btnInstall.disabled = false
    btnInstall.removeAttribute('aria-busy')
  }
}

async function doUninstall() {
  btnUninstall.disabled = true

  try {
    await invoke('uninstall_app')
    installSection.classList.add('hidden')
    log('Kyaraben has been uninstalled from applications menu.')
  } catch (err) {
    log(`Uninstall failed: ${err}`)
  } finally {
    btnUninstall.disabled = false
  }
}

function dismissInstall() {
  sessionStorage.setItem('install-dismissed', 'true')
  installSection.classList.add('hidden')
}

function renderSystems() {
  systemList.innerHTML = systems
    .map(
      (sys) => `
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
  `,
    )
    .join('')
}

// Make toggleSystem available globally for inline handler
window.toggleSystem = (sysId, checked) => {
  if (checked) {
    selectedSystems.add(sysId)
  } else {
    selectedSystems.delete(sysId)
  }
}

function log(message) {
  outputSection.classList.remove('hidden')
  logElement.textContent += `${message}\n`
  logElement.scrollTop = logElement.scrollHeight
}

function clearLog() {
  logElement.textContent = ''
}

// Button handlers
btnApply.addEventListener('click', async () => {
  clearLog()
  btnApply.disabled = true
  btnApply.setAttribute('aria-busy', 'true')

  try {
    // Save config first
    const systemsConfig = {}
    for (const sysId of selectedSystems) {
      const sys = systems.find((s) => s.id === sysId)
      if (sys && sys.emulators.length > 0) {
        systemsConfig[sysId] = sys.emulators[0].id
      }
    }

    await invoke('set_config', {
      userStore: userStoreInput.value,
      systems: systemsConfig,
    })

    log('Applying configuration...')

    // Apply
    const result = await invoke('apply')
    for (const line of result) {
      log(line)
    }
    log('Done!')
  } catch (err) {
    log(`Error: ${err}`)
  } finally {
    btnApply.disabled = false
    btnApply.removeAttribute('aria-busy')
  }
})

btnDoctor.addEventListener('click', async () => {
  clearLog()
  provisionsSection.classList.remove('hidden')

  try {
    const result = await invoke('doctor')
    provisionsList.innerHTML = ''

    for (const [sysId, provisions] of Object.entries(result)) {
      const sys = systems.find((s) => s.id === sysId)
      const sysName = sys ? sys.name : sysId

      let html = `<h3>${sysName}</h3>`

      if (provisions.length === 0) {
        html += '<p>No provisions required.</p>'
      } else {
        html += '<ul>'
        for (const prov of provisions) {
          const statusClass =
            prov.status === 'found' ? 'status-ok' : prov.required ? 'status-missing' : 'status-warn'
          const statusText = prov.status === 'found' ? 'OK' : prov.required ? 'MISSING' : 'optional'
          html += `<li>
            <code>${prov.filename}</code>
            <span class="status-badge ${statusClass}">${statusText}</span>
            <small>${prov.description || ''}</small>
          </li>`
        }
        html += '</ul>'
      }

      provisionsList.innerHTML += html
    }
  } catch (err) {
    log(`Error: ${err}`)
  }
})

btnStatus.addEventListener('click', async () => {
  clearLog()

  try {
    const status = await invoke('status')
    log(`Emulation folder: ${status.userStore}`)
    log(`Enabled systems: ${status.enabledSystems.join(', ') || 'none'}`)
    log(`Installed emulators: ${status.installedEmulators.length}`)
    if (status.lastApplied) {
      log(`Last applied: ${new Date(status.lastApplied).toLocaleString()}`)
    }
  } catch (err) {
    log(`Error: ${err}`)
  }
})

btnInstall.addEventListener('click', doInstall)
btnDismissInstall.addEventListener('click', dismissInstall)
btnUninstall.addEventListener('click', doUninstall)

init()
