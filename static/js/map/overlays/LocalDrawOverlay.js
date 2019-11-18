import { localStorageAvailable, getLocalObject, setLocalObject } from "../LocalStorage.js"


export default L.GeoJSON.extend({
  initialize: function() {
    L.GeoJSON.prototype.initialize.call(this);
		if (localStorageAvailable()) {
	    this.drawControl = new L.Control.Draw({
	      position: 'topleft',
	      draw: {
	        polygon: {
	          allowIntersection: false, // Restricts shapes to simple polygons
	          drawError: {
	            color: '#e1e100', // Color the shape will turn when intersects
	            message: '<strong>Oh snap!<strong> you can\'t draw that!', // Message that will show when intersect
	          },
	          shapeOptions: {
	            color: '#97009c',
	            stroke:true,
	          }
	        },
	        polyline: false,
	        circle: false, // Turns off this drawing tool
	        rectangle: false,
	        marker: false,
	        circlemarker: false,
	      },
	      edit: { featureGroup: this, }
	    });
    } else {
		 console.error("Local storage not available for LocalDraw layer.")
		 this.drawControl = new L.Control.Draw({
			 position: 'topleft',
			 draw: false,
			 edit: false,
		 });
	 }
  },

  getMaxDisplayedZoom: function(){
    return 1;
  },

  getMinDisplayedZoom: function(){
    return 10;
  },

  onDrawCreated: function(e) {
    this.addLayer(e.layer);
		var jsonstring = JSON.stringify(this.toGeoJSON());
		console.log(jsonstring);
		if (localStorageAvailable())
			setLocalObject("test", this.toGeoJSON());
  },

  onAdd: function(map) {
    this.map = map;
    map.on("draw:created", this.onDrawCreated, this);
    map.addControl(this.drawControl);
		if (localStorageAvailable()) {
			var stored = getLocalObject("test", this);
			if (stored)
				this.addData(stored);
		}
  },

  onRemove: function(map) {
    this.clearLayers();
    map.off("draw:created", this.onDrawCreated, this);
    map.removeControl(this.drawControl);
  },
});
