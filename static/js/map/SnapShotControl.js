import '../lib/html2canvas.js';

export default L.Control.extend({
  initialize: function(wsChannel, opts) {
      L.Control.prototype.initialize.call(this, opts);
  },

  _ongoingsnap: false,

  _takeSnapShot: function(e) {
    if (this._ongoingsnap) return;
    this._ongoingsnap = true;
    var me = this;
    html2canvas(document.body, {
      ignoreElements: function(element) {
        return (
          // Remove all controls
          (element.classList.contains('leaflet-control') &&
          // Except legend
            !element.classList.contains("localdrawoverlay-legend-box"))
          // Remove also players
          || element.classList.contains('mapserver-object-player')
        );
      },
    }).then(function(canvas) {
      me._ongoingsnap = false;
      var link = document.createElement("a");
      link.download = "map.png";
      link.href = canvas.toDataURL('image/png');
      link.click();
    }, function() {
      // In case things went wrong
      me._ongoingsnap = false;
    });
  },

  onAdd: function(map) {
    this._div = L.DomUtil.create('div', 'snapshot-control leaflet-control');
    L.DomEvent.disableClickPropagation(this._div);
    this._button = L.DomUtil.create('a', 'snapshot-button', this._div);
    L.DomEvent.on(this._button, 'click', this._takeSnapShot, this);
    return this._div;
  },

  onRemove: function(map) {
    L.DomEvent.off(this._button, 'click', this._takeSnapShot, this);
  },
});
