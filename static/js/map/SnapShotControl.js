import '../lib/html2canvas.js';

export default L.Control.extend({
  _takeSnapShot: function(e) {
    html2canvas(document.body, {
      ignoreElements: function(element) {
        return (element.className == "leaflet-control-container");
      },
    }).then(function(canvas) {
      var link = document.createElement("a");
      link.download = "map.png";
      link.href = canvas.toDataURL('image/png');
      link.click();
    });
  },

  onAdd: function(map) {
    this._div = L.DomUtil.create('div', 'snapshot-control leaflet-control');
    this._button = L.DomUtil.create('a', 'snapshot-button', this._div);
    L.DomEvent.on(this._button, 'click', this._takeSnapShot, this);
    return this._div;
  },

  onRemove: function(map) {
    L.DomEvent.off(this._button, 'click', this._takeSnapShot, this);
  },
});
