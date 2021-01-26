export default class Emiter {
  constructor() {
    this.evts = new Map();
  }
  on(type, handler) {
    if (!type || !handler) {
      return
    }
    const hs = this.evts.get(type)
    const nextHs = hs && hs.push(handler)
    if (!nextHs) {
      this.evts.set(type, [handler])
    }
  }
  off(type, handler) {
    if (!type || !handler) {
      return
    }
    const hs = this.evts.get(type)
    if (hs) {
      hs.splice(hs.indexOf(handler) >>> 0, 1);
    }
  }
  emit(type, evt) {
    (this.evts.get(type) || []).slice().forEach((handler) => { handler(evt); });
    (this.evts.get('*') || []).slice().forEach((handler) => { handler(type, evt); });
  }
}