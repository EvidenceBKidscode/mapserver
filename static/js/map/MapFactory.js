import wsChannel from '../WebSocketChannel.js';
import SimpleCRS from './SimpleCRS.js';
import CoordinatesDisplay from './CoordinatesDisplay.js';
import WorldInfoDisplay from './WorldInfoDisplay.js';
import TopRightControl from './TopRightControl.js';
import SnapShotControl from './SnapShotControl.js';
import { OverlaySetup, GetLocalizedOverlays } from './Overlaysetup.js';
import CustomOverlay from './CustomOverlay.js';
import RealtimeTileLayer from './RealtimeTileLayer.js';

import config from '../config.js';

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

  var tileLayer = new RealtimeTileLayer(wsChannel, layerId, map);
  tileLayer.addTo(map);

  //All overlays
  var overlays = {};
	overlays.Top25 = new L.ImageOverlay(
		"http://localhost:8080/api/rastermaps/ign_scan25.png",
		[[-2560, -2560], [2560, 2560]], {opacity: 0.5});
	overlays.Arien = new L.ImageOverlay(
		"http://localhost:8080/api/rastermaps/ign_ortho.png",
		[[-2560, -2560], [2560, 2560]], {opacity: 0.5});

  OverlaySetup(cfg, map, overlays);
  CustomOverlay(map, overlays);

  new CoordinatesDisplay({ position: 'bottomleft' }).addTo(map);
  new WorldInfoDisplay(wsChannel, { position: 'bottomright' }).addTo(map);
  new TopRightControl({ position: 'topright' }).addTo(map);
  new SnapShotControl({ position: 'topright' }).addTo(map);

  // Layer Control
  L.control.layers({}, GetLocalizedOverlays(overlays), { position: "topright" }).addTo(map);

  return map;
}
