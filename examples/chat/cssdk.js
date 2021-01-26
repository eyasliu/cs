(function (global, factory) {
  typeof exports === 'object' && typeof module !== 'undefined' ? module.exports = factory() :
  typeof define === 'function' && define.amd ? define(factory) :
  (global = typeof globalThis !== 'undefined' ? globalThis : global || self, global.CS = factory());
}(this, (function () { 'use strict';

  class Emiter {
    constructor() {
      this.evts = new Map();
    }
    on(type, handler) {
      if (!type || !handler) {
        return
      }
      const hs = this.evts.get(type);
      const nextHs = hs && hs.push(handler);
      if (!nextHs) {
        this.evts.set(type, [handler]);
      }
    }
    off(type, handler) {
      if (!type || !handler) {
        return
      }
      const hs = this.evts.get(type);
      if (hs) {
        hs.splice(hs.indexOf(handler) >>> 0, 1);
      }
    }
    emit(type, evt) {
      (this.evts.get(type) || []).slice().forEach((handler) => { handler(evt); });
      (this.evts.get('*') || []).slice().forEach((handler) => { handler(type, evt); });
    }
  }

  const randStr = () => Math.random().toString(36).substr(2);

  class CS extends Emiter {
    constructor(url) {
      super();
      this.url = url;
      this.sendTimeout = 10000;
      this._progress = new Map();
      this._ins = this._init();
      this._events();
    }
    get adapterName() {
      if (this.url.indexOf('ws') === 0) {
        return 'ws'
      }
      return 'http'
    }
    async send(cmd, data) {
      const body = { cmd, data };
      body.seqno = randStr();
      let resp;
      if (this.adapterName === 'ws') {
        resp = await this._sendByWs(body);
      } else {
        resp = await fetch(this.url, {
          method: "POST",
          headers: {
            'Content-Type': "application/json",
          },
          body: JSON.stringify(body),
          credentials: 'include',
        }).then(r => r.json());
      }

      if (resp.code != 0) {
        const err = new Error(resp.msg);
        err.response = resp;
        throw err
      }

      return resp.data
    }
    _init() {
      if (this.adapterName === 'ws') {
        // return new WS(this.url)
        return this._initWs()
      }
      // return new HTTP(this.url)
      return this._initHttp()
    }
    _initWs() {
      this.adapter = new WebSocket(this.url);

    }
    _initHttp() {
      this.adapter = new EventSource(this.url);
    }
    _events() {
      this.adapter.onopen = e => {
        this.emit("cs.connected", e);
      };
      this.adapter.onclose = e => {
        this.emit('cs.closed', e);
      };
      this.adapter.onmessage = this._onMessage.bind(this);
    }
    _onMessage(e) {
      const body = JSON.parse(e.data);
      this.emit('cs.message', body);
      const { cmd, seqno, data } = body;
      this.emit(cmd, data);

      // sned progress
      const p = this._progress.get(seqno);
      if (p) {
        const [resolve] = p;
        resolve(body);
      }
    }
    _invokeProgress() {

    }
    _sendByWs(body) {
      return new Promise((resolve, reject) => {
        const seqno = body.seqno;
        const resolveH = resp => {
          const { cmd, code, msg, data } = resp;
          if (code != 0) {
            const err = new Error(msg);
            err.response = resp;
            reject(err);
            return
          }
          resolve(resp);
          clear();
        };
        const rejectH = err => {
          reject(err);
          clear();
        };
        const clear = () => {
          clearTimeout(stfd);
          this._progress.delete(seqno);
        };
        const stfd = setTimeout(() => {
          rejectH(new Error("websocket send wait timeout"));
        }, this.sendTimeout);

        this._progress.set(seqno, [resolveH]);
        const s = JSON.stringify(body);
        this.adapter.send(s);
      })
    }


  }

  return CS;

})));
