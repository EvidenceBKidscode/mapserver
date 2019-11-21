import config from '../config.js';

var LOCALPREFIX = "kidscode"

export function localStorageAvailable() {
  const cfg = config.get();

  if (! cfg.worldid) {
    return false
  }
  var storage;
  try {
    storage = window["localStorage"];
    var x = '__storage_test__';
    storage.setItem(x, x);
    storage.removeItem(x);
    return true;
  }
  catch(e) {
    return e instanceof DOMException && (
      // everything except Firefox
      e.code === 22 ||
      // Firefox
      e.code === 1014 ||
      // test name field too, because code might not be present
      // everything except Firefox
      e.name === 'QuotaExceededError' ||
      // Firefox
      e.name === 'NS_ERROR_DOM_QUOTA_REACHED') &&
      // acknowledge QuotaExceededError only if there's something already stored
      (storage && storage.length !== 0);
  }
}

export function getLocalObject(name) {
  const cfg = config.get();

  return JSON.parse(localStorage.getItem(LOCALPREFIX + '-' + cfg.worldid + '-' + name));
}

export function setLocalObject(name, object) {
  const cfg = config.get();

  localStorage.setItem(LOCALPREFIX + '-' + cfg.worldid + '-' + name, JSON.stringify(object));
}
