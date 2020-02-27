/*
wsChannel.addListener("incremental-render-progress", function(ev){
      console.log("Received incremental-render-progress event\n");
*/

var mapServerStatusRender = function(title, progress){
  if (progress < 1)
    return m("div", [
      m("p", [ title, " (", Math.round(progress * 100), "%)" ]),
      m("div", {class: "serverstatus-bar"},
        m("div", {class: "serverstatus-progress", style:"width:"+Math.round(progress * 100)+"%"}, "")
      )
    ])
  else
    return m("div", [
      m("p", "Carte à jour"),
      m("div", {class: "serverstatus-bar"}, "")
    ])
};

export default L.Control.extend({
    initialize: function(wsChannel, opts) {
        L.Control.prototype.initialize.call(this, opts);
        this.wsChannel = wsChannel;
    },

    onAdd: function() {
      var div = L.DomUtil.create('div', 'leaflet-bar leaflet-custom-display mapserver-status');
      m.render(div, mapServerStatusRender(1))
      this.wsChannel.addListener("incremental-render-progress", function(info){
        m.render(div, mapServerStatusRender("Mise à jour de la carte", info.progress));
      });
      this.wsChannel.addListener("initial-render-progress", function(info){
        m.render(div, mapServerStatusRender("Rendu initial de la carte", info.progress));
      });

      return div;
    }
});
