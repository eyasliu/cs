import Emit from './emiter'
import { randStr } from './util'

const msgBlog = "blob"
const msgText = "text"

const defaultOptions = {
  headers: {},
  wsMsgType: msgText,
  wsHeartBeatTime: 10000,
  withCredentials: true,
}

class CS extends Emit {
  constructor(url, options) {
    super()
    if (typeof url === 'object') {
      options = url
      url = options.url
    }
    if (!url) {
      throw new Error("cs initial must require url")
    }


    this.url = url
    this.options = Object.assign({}, defaultOptions, options || {})
    this.sendTimeout = 10000
    this._progress = new Map()
    this._init()
  }
  get adapterName() {
    if (this.url.indexOf('ws') === 0) {
      return 'ws'
    }
    return 'http'
  }
  async resetUrl(url) {
    this.destroy()
    this.url = url
    this._init()
  }
  destroy() {
    if (!this.adapter) {
      return
    }
    this.adapter.close()
    this.adapter = null
    this.wshb && clearInterval(this.wshb)
    this.wshb = null
  }
  async send(cmd, data) {
    const body = { cmd, data }
    body.seqno = randStr()
    let resp
    if (this.adapterName === 'ws') {
      resp = await this._sendByWs(body)
    } else {
      resp = await fetch(this.url, {
        method: "POST",
        headers: {
          'Content-Type': "application/json",
          ...(this.options.headers || {}),
        },
        body: JSON.stringify(body),
        credentials: this.options.withCredentials ? 'include' : 'omit',
      }).then(r => r.json())
    }

    if (resp.code != 0) {
      const err = new Error(resp.msg)
      err.response = resp
      throw err
    }

    return resp.data
  }
  _init() {
    if (this.adapter) {
      this.destroy()
    }

    if (this.adapterName === 'ws') {
      this.adapter = this._initWs()
    } else {
      this.adapter = this._initHttp()
    }
    this._events()
  }
  _initWs() {
    const ws = new WebSocket(this.url)
    ws.addEventListener('close', e => {
      this.wshb && clearInterval(this.wshb)
      this.wshb = null
    })
    // 心跳
    this.wshb = setInterval(() => {
      const hb = this.options.wsMsgType == msgBlog ? new Blob([''], { type: 'text/plain' }) : "";
      ws.send(hb)
    }, this.options.wsHeartBeatTime)
    return ws
  }
  _initHttp() {
    if (this.options.withCredentials === false) {
      console.warn("[CS]http adapter required Cookie to work fine, but you set withCredentials=false, Server-Send Event maybe invalid")
    }
    return new EventSource(this.url, {
      headers: this.options.headers,
      withCredentials: this.options.withCredentials,
    })
  }
  _events() {
    this.adapter.onopen = e => {
      this.emit("cs.connected", e)
    }
    this.adapter.onclose = e => {
      this.emit('cs.closed', e)
      // sse 在浏览器自己会自动重连，websocket 需要手动触发重连
      if (this.adapterName === 'ws') {
        setTimeout(this._init.bind(this), 100)
      }
    }
    this.adapter.onmessage = this._onMessage.bind(this)
  }
  async _onMessage(e) {
    let raw = e.data
    if (raw instanceof Blob) {
      raw = await raw.text()
    }
    const body = JSON.parse(raw)
    this.emit('cs.message', body)
    const { cmd, seqno, data } = body
    this.emit(cmd, data)

    // sned progress
    const p = this._progress.get(seqno)
    if (p) {
      const [resolve] = p
      resolve(body)
    }
  }
  _sendByWs(body) {
    return new Promise((resolve, reject) => {
      const seqno = body.seqno
      const resolveH = resp => {
        const { cmd, code, msg, data } = resp
        if (code != 0) {
          const err = new Error(msg)
          err.response = resp
          reject(err)
          return
        }
        resolve(resp)
        clear()
      }
      const rejectH = err => {
        reject(err)
        clear()
      }
      const clear = () => {
        clearTimeout(stfd)
        this._progress.delete(seqno)
      }
      const stfd = setTimeout(() => {
        rejectH(new Error("websocket send wait timeout"))
      }, this.sendTimeout)

      this._progress.set(seqno, [resolveH])
      let s = JSON.stringify(body)
      if (this.options.wsMsgType !== msgText) {
        s = new Blob([s], { type: 'text/plain' })
      }
      this.adapter.send(s)
    })
  }


}

export default CS