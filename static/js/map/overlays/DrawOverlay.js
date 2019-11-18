
export default L.GeoJSON.extend({
  initialize: function() {
    L.GeoJSON.prototype.initialize.call(this);
    this.jsonstring = ""
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
  },

  getMaxDisplayedZoom: function(){
    return 1;
  },

  getMinDisplayedZoom: function(){
    return 10;
  },

  onDrawCreated: function(e) {
    this.addLayer(e.layer);
		this.jsonstring = "";
		var json = this.toGeoJSON();
		this.jsonstring = JSON.stringify(json);
		m.request({
	    method: "POST",
	    url: "api/draw/",
	    data: json
	  });
  },

  onAdd: function(map) {
    this.map = map;
    map.on("draw:created", this.onDrawCreated, this);
    map.addControl(this.drawControl);
    if (this.jsonstring != "") {
      this.addData(JSON.parse(this.jsonstring));
    }
  },

  onRemove: function(map) {
    this.jsonstring = "";
    var jsonstring = JSON.stringify(this.toGeoJSON());
    this.jsonstring = jsonstring;
    this.clearLayers();
    map.off("draw:created", this.onDrawCreated, this);
    map.removeControl(this.drawControl);
  },
});
