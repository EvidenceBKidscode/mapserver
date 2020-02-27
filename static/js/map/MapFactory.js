import wsChannel from '../WebSocketChannel.js';
import SimpleCRS from './SimpleCRS.js';
import CoordinatesDisplay from './CoordinatesDisplay.js';
import WorldInfoDisplay from './WorldInfoDisplay.js';
import TopRightControl from './TopRightControl.js';
import SnapShotControl from './SnapShotControl.js';
import { OverlaySetup, GetLocalizedOverlays } from './Overlaysetup.js';
import CustomOverlay from './CustomOverlay.js';
import RealtimeTileLayer from './RealtimeTileLayer.js';
import MapServerStatus from './MapServerStatus.js';

import config from '../config.js';
import '../lib/L.Control.Opacity.js';

export function createMap(node, layerId, zoom, lat, lon){

  const cfg = config.get();

  const map = L.map(node, {
    minZoom: 7,
    maxZoom: 13,
    center: [lat, lon],
    zoom: zoom,
    crs: SimpleCRS
  });

  map.attributionControl.addAttribution('<a href="https://github.com/minetest-tools/mapserver">Minetest Mapserver</a>');

  // Quick and dirty image layers ~~> Shoud go into a separate file
  // Supose upperleft and lowerright corners are elements 3 and 1 of given coordinates
  if (cfg.rasteroverlays != null) {
    var bounds = [cfg.geometry.coordinatesGame[3], cfg.geometry.coordinatesGame[1]];
    var labels = {}
    cfg.rasteroverlays.forEach(overlay => {
      var layer = new L.ImageOverlay(
        "http://localhost:8080/api/rastermaps/" + overlay.texture, bounds, {opacity: 0});
      labels[overlay.label] = layer;
      layer.addTo(map);
    });

    L.control.opacity(labels, { label: "Cartes supplémentaires" }).addTo(map);
  }
  // End quick and dirty

  var tileLayer = new RealtimeTileLayer(wsChannel, layerId, map);
  tileLayer.addTo(map);

  //All overlays
  var overlays = {};
  OverlaySetup(cfg, map, overlays);
  CustomOverlay(map, overlays);

  new CoordinatesDisplay({ position: 'bottomleft' }).addTo(map);
  new WorldInfoDisplay(wsChannel, { position: 'bottomright' }).addTo(map);
  new TopRightControl({ position: 'topright' }).addTo(map);
  new SnapShotControl({ position: 'topright' }).addTo(map);

  // Layer Control
  L.control.layers({}, GetLocalizedOverlays(overlays), { position: "topright" }).addTo(map);

  // Mapserver status
  new MapServerStatus(wsChannel, { position: 'bottomright' }).addTo(map);

  return map;
}
