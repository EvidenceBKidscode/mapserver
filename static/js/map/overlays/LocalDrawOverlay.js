/*
  NOTE: GeoJSON is unable to store circles. So, as we need to store and edit
  any shapes, this LocalDrawOverlay will not use GeoJSON for storage but a
  custom format.
*/

import { localStorageAvailable, getLocalObject, setLocalObject } from "../LocalStorage.js"

function layerToStorable(layer) {
  // NOTE: Should test rectangle before polygon because rectangle IS a
  // polygon also
  if (layer instanceof L.Rectangle) {
    var bounds = layer.getBounds();
    return {
      type: 'rectangle',
      // Does not work with getBounds (?), have to pass this array
      Bounds:[[bounds.getNorthWest().lat, bounds.getNorthWest().lng],
        [bounds.getSouthEast().lat, bounds.getSouthEast().lng]],
      attributes:layer.attributes,
    };
  }
  if (layer instanceof L.Polygon)
    return {
      type: 'polygon',
      LatLngs:layer.getLatLngs(),
      attributes:layer.attributes,
    };

  else if (layer instanceof L.Circle)
    return {
      type: 'circle',
      LatLngs:layer.getLatLng(),
      Radius:layer.getRadius(),
      attributes:layer.attributes,
    };

  console.log("Error layer with unknown type could not be stored:");
  console.log(layer);
};

function storableToLayer(storable) {
  var layer = null;
  switch (storable.type) {
    case 'rectangle':
      layer = L.rectangle(storable.Bounds, {});
      break;
    case 'polygon':
      layer = L.polygon(storable.LatLngs, {});
      break;
    case 'circle':
      layer = L.circle(storable.LatLngs, storable.Radius, {});
      break;
    default:
      console.log("Unknown stored layer type: " + storable.type);
      return null;
  }
  layer.attributes = storable.attributes;
  return layer;
}

Object.assign(L.drawLocal, {
  draw: {
    toolbar: {
      actions: {
        title: 'Annuler le dessin',
          text: 'Annuler'
      },
      finish: {
        title: 'Terminer le dessin',
        text: 'Terminer'
      },
      undo: {
        title: 'Supprimer le dernier point dessiné',
        text: 'Supprimer le dernier point'
      },
      buttons: {
        polyline: 'Dessiner une polyligne',
        polygon: 'Dessiner un polygone',
        rectangle: 'Dessiner un rectangle',
        circle: 'Dessiner un cercle',
        marker: 'Dessiner un marqueur',
        circlemarker: 'Dessiner un marqueur-cercle'
      }
    },
    handlers: {
      circle: {
        tooltip: {
          start: 'Cliquer et glisser pour dessiner un cercle.'
        },
        radius: 'Rayon'
      },
      circlemarker: {
        tooltip: {
          start: 'Cliquer sur la carte pour placer un marqueur-cercle.'
        }
      },
      marker: {
        tooltip: {
          start: 'Cliquer sur la carte pour placer un marqueur.'
        }
      },
      polygon: {
        tooltip: {
          start: 'Cliquer pour commencer à dessiner une forme.',
          cont: 'Cliquer pour continuer à dessiner la forme.',
          end: 'Cliquer sur le premier point pour fermer la forme.'
        }
      },
      polyline: {
        error: '<strong>Erreur:</strong> les bords de la forme ne peuvent pas se croiser !',
        tooltip: {
          start: 'Cliquer pour commencer à dessiner une ligne.',
          cont: 'Cliquer pour continuer à dessiner la ligne.',
          end: 'Cliquer sur le dernier point pour terminer la ligne.'
        }
      },
      rectangle: {
        tooltip: {
          start: 'Cliquer et glisser pour dessiner un rectangle.'
        }
      },
      simpleshape: {
        tooltip: {
          end: 'Lacher la souris pour finir le dessin.'
        }
      }
    },
  },
  edit: {
    toolbar: {
      actions: {
        save: {
          title: 'Enregistrer les modifications',
          text: 'Enregistrer'
        },
        cancel: {
          title: 'Annuler les modifications',
          text: 'Annuler'
        },
        clearAll: {
          title: 'Supprimer toutes les formes',
          text: 'Tout supprimer'
        }
      },
      buttons: {
        edit: 'Modifier les formes',
        editDisabled: 'Pas de forme à modifier',
        remove: 'Supprimer des formes',
        removeDisabled: 'Pas de forme à supprimer'
      }
    },
    handlers: {
      edit: {
        tooltip: {
          text: 'Glisser les poignées ou les marqueurs pour modifier les formes.',
          subtext: 'Cliquer sur annuler pour annuler les modifications.'
        }
      },
      remove: {
        tooltip: {
          text: 'Cliquer sur une forme pour la supprimer.'
        }
      }
    }
  },
});

var ColorControl = L.Control.extend({
  colors: ["#ec7063", "#9b59b6", "#3498db", "#2ecc71", "#f4d03f", "#f39c12"],
  buttons: [],
  selectedColor: 0,

  initialize: function (options) {
    if (options) {
      L.setOptions(this, options)
    }
  },

  getSelectedColor: function() {
    return this.colors[this.selectedColor];
  },

  selectColor: function(name) {
    for (let i = 0; i < this.colors.length; i++)
      if (this.colors[i] == name) {
        this.selectColorNumber(i);
        return;
      }
  },

  selectColorNumber: function(number) {
    if (number < 0 || number >= this.buttons.length)
      return;

    for (let i = 0; i < this.buttons.length; i++)
      this.buttons[i].classList.remove("selected");

    this.buttons[number].classList.add("selected");
    this.selectedColor = number;

    this.fire("colorselected", {
      color: this.colors[number],
      number: number, });
  },

  onAdd: function(map) {
    var div = L.DomUtil.create('div', 'leaflet-bar localdrawoverlay-bar');

    for (let i = 0; i < this.colors.length; i++) {
      this.buttons[i] = L.DomUtil.create('div', 'localdrawoverlay-color-box', div);
      this.buttons[i].style["background-color"] = this.colors[i];
      L.DomEvent.on(this.buttons[i], 'click',
        function(e) { this.selectColorNumber(i); }, this);
    }
    this.selectColorNumber(this.selectedColor);
    return div;
  },

  onRemove: function(map) {
  },
});

// Add fire capability to ColorControl
var version = L.version.split('.');
//If Version is >= 1.2.0
if (parseInt(version[0], 10) === 1 && parseInt(version[1], 10) >= 2) {
  ColorControl.include(L.Evented.prototype);
} else {
  ColorControl.include(L.Mixin.Events);
}

export default L.FeatureGroup.extend({
  initialize: function() {
    L.FeatureGroup.prototype.initialize.call(this);
    if (localStorageAvailable()) {
      this.drawControl = new L.Control.Draw({
        position: 'topleft',
        draw: {
          polygon: {
            allowIntersection: false, // Restricts shapes to simple polygons
            shapeOptions: {
              stroke:true,
            }
          },
          circle: {
            shapeOptions: {
              stroke:true,
            }
          },
          rectangle: {
            shapeOptions: {
              stroke:true,
            }
          },
          polyline: false,
          marker: false,
          circlemarker: false,
        },
        selected_layer: null,
      });
      this.colorControl = new ColorControl({
        position:'topleft',
        edit: {
          featureGroup: this,
        },
      });
      this.colorControl.on("colorselected", this.colorSelected, this);
    } else {
      console.error("Local storage not available for LocalDraw layer.")
      this.drawControl = null;
      this.colorControl = null;
    }
  },

  getMaxDisplayedZoom: function(){
    return 1;
  },

  getMinDisplayedZoom: function(){
    return 10;
  },

  colorSelected:function(e) {
    // Change selected shape color
    if (this.selected_layer != null) {
      this.selected_layer.attributes.color = e.color;
      this.updateStyle(this.selected_layer);
    }
    // Change draw control color
    if (this.drawControl != null) {
      this.drawControl.options.draw.polygon.shapeOptions.color = e.color;
      this.drawControl.options.draw.circle.shapeOptions.color = e.color;
      this.drawControl.options.draw.rectangle.shapeOptions.color = e.color;
    }
  },

  updateStyle:function(layer) {
    if (layer == this.selected_layer)
      layer.setStyle({
        color: layer.attributes.color,
        dashArray: '10, 10',
      });
    else
    layer.setStyle({
      color: layer.attributes.color,
      dashArray: null,
    });
  },

  unselectLayer:function() {
    if (this.selected_layer != null) {
      var layer = this.selected_layer
      this.selected_layer = null;
      layer.editing.disable();
      this.updateStyle(layer);
      // TODO: Comment gerer les sauvegardes et annulations ?
      this.save();
    }
  },

  selectLayer:function(layer) {
    // Unselect
    if (layer == this.selected_layer || !this.hasLayer(layer)) {
      this.unselectLayer();
      return;
    }

    this.unselectLayer();
    this.selected_layer = layer;
    layer.editing.enable();
    layer.bringToFront();
    this.updateStyle(layer);
    // Send selected shape color to color control
    if (this.colorControl != null)
      this.colorControl.selectColor(this.selected_layer.attributes.color);
  },

  save:function() {
    if (localStorageAvailable()) {
      var storage = []
      this.eachLayer(function (layer) {
        var storable = layerToStorable(layer);
        if (storable != null)
          storage.push(storable);
      });
      setLocalObject("test", JSON.stringify(storage));
    }
  },

  load:function() {
    if (localStorageAvailable()) {
      this.clearLayers();
      var overlay = this
      var storage = JSON.parse(getLocalObject("test"));
      if (storage != null)
        storage.forEach(function(storable) {
          var layer = storableToLayer(storable);
          if (layer != null)
            overlay.addLayer(layer);
        });
    }
  },

  addLayer:function(layer) {
    var overlay = this;
    if (layer.attributes == null) {
      layer.attributes = {};
      layer.attributes.color = 'blue';
    }
    this.updateStyle(layer);

    // Select layer on click
    layer.on('click', function(e) { overlay.selectLayer(e.target); });

    L.FeatureGroup.prototype.addLayer.call(this, layer);
  },

  onDrawEdited:function(e) {
    var overlay = this;
    e.layers.eachLayer(function (layer) {
      overlay.addLayer(layer);
    });
    this.save();
  },

  onDrawCreated: function(e) {
    e.layer.attributes = {};
    var storable = layerToStorable(e.layer);
    if (storable == null) return;
    var layer = storableToLayer(storable);
    if (layer == null) return;
    layer.attributes.color = this.colorControl.getSelectedColor();
    this.addLayer(layer);
    // Automatically select new layer
    this.selectLayer(layer);
    this.save();
  },

  onDrawDeleted:function(e) {
    var overlay = this;
    e.layers.eachLayer(function (layer) {
      overlay.removeLayer(layer);
    })
    this.save();
  },

  onAdd: function(map) {
    this.map = map;
    map.on("draw:created", this.onDrawCreated, this);
    map.on("draw:edited", this.onDrawEdited, this);
    map.on("draw:deleted", this.onDrawDeleted, this);
    if (this.drawControl != null) map.addControl(this.drawControl);
    if (this.colorControl != null) map.addControl(this.colorControl);
    this.load();
  },

  onRemove: function(map) {
    this.clearLayers();
    map.off("draw:created", this.onDrawCreated, this);
    map.off("draw:edited", this.onDrawEdited, this);
    map.off("draw:deleted", this.onDrawDeleted, this);
    if (this.drawControl != null) map.removeControl(this.drawControl);
    if (this.colorControl != null) map.removeControl(this.colorControl);
  },
});
