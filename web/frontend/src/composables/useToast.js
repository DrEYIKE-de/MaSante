export function useToast() {
  function toast(msg, type = 'info', duration = 4000) {
    let container = document.getElementById('ms-toasts')
    if (!container) {
      container = document.createElement('div')
      container.id = 'ms-toasts'
      container.style.cssText = 'position:fixed;top:20px;right:20px;z-index:99999;display:flex;flex-direction:column;gap:8px'
      document.body.appendChild(container)
    }
    const el = document.createElement('div')
    el.textContent = msg
    const colors = { success: '#2d7a4f', error: '#a63d3d', warning: '#b8860b', info: '#3d6b8a' }
    el.style.cssText = `padding:12px 20px;border-radius:8px;font-size:.85rem;font-weight:500;color:#fff;box-shadow:0 8px 30px rgba(0,0,0,.12);transform:translateX(120%);transition:transform .3s;max-width:400px;background:${colors[type] || colors.info}`
    container.appendChild(el)
    requestAnimationFrame(() => (el.style.transform = 'translateX(0)'))
    setTimeout(() => {
      el.style.transform = 'translateX(120%)'
      setTimeout(() => el.remove(), 300)
    }, duration)
  }

  return {
    success: (msg) => toast(msg, 'success'),
    error: (msg) => toast(msg, 'error'),
    warning: (msg) => toast(msg, 'warning'),
    info: (msg) => toast(msg, 'info'),
  }
}
