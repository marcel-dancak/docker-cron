
export default function WebsocketMessenger (url) {
  let listeners = []
  let openListeners = []
  const activeRequests = {}
  const socket = new WebSocket(url)
  let timer = null
  const ws = {
    connected: false,
    pluginConnected: false,
    clientInfo: '',

    bind (type, callback) {
      listeners.push({ type, callback })
      return () => this.unbind(type, callback)
    },
    unbind (type, callback) {
      listeners = listeners.filter(l => l.type !== type || l.callback !== callback)
    },
    send (name, data) {
      const msg = { type: name, data }
      socket.send(JSON.stringify(msg))
    },
    request (name, data) {
      return new Promise((resolve, reject) => {
        activeRequests[name] = { resolve, reject }
        this.send(name, data)
      })
    },
    close () {
      if (timer !== null) {
        clearInterval(timer)
        timer = null
      }
      socket.close()
    },
    onopen () {
      return new Promise(resolve => {
        if (ws.connected) {
          resolve()
        } else {
          openListeners.push(resolve)
        }
      })
    }
  }

  socket.onopen = () => {
    ws.connected = true
    openListeners.forEach(cb => cb())
    openListeners = []
  }
  socket.onclose = () => {
    ws.connected = false
  }
  socket.onmessage = (e) => {
    const msg = JSON.parse(e.data)
    if (activeRequests[msg.type]) {
      if (msg.status && msg.status >= 400) {
        activeRequests[msg.type].reject(msg)
      } else {
        activeRequests[msg.type].resolve(msg)
      }
      delete activeRequests[msg.type]
    }
    listeners.filter(l => l.type === msg.type).forEach(l => l.callback(msg))
  }
  timer = setInterval(() => {
    ws.send('ping')
  }, 30 * 1000)
  return ws
}
