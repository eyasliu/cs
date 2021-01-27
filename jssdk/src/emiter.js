export default class Emiter {
  constructor() {
    this._listeners = new Map();
  }
  on(type, handler) {
    if (!type || !handler) {
      return
    }
    const hs = this._listeners.get(type)
    const nextHs = hs && hs.push(handler)
    if (!nextHs) {
      this._listeners.set(type, [handler])
    }
  }
  off(type, handler) {
    if (!type || !handler) {
      return
    }
    const hs = this._listeners.get(type)
    if (hs) {
      hs.splice(hs.indexOf(handler) >>> 0, 1);
    }
  }
  emit(type, evt) {
    (this._listeners.get(type) || []).slice().forEach((handler) => { handler(evt); });
    (this._listeners.get('*') || []).slice().forEach((handler) => { handler(type, evt); });
  }
}