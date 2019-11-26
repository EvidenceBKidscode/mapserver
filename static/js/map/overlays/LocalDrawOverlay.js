/*
  NOTE: GeoJSON is unable to store circles. So, as we need to store and edit
  any shapes, this LocalDrawOverlay will not use GeoJSON for storage but a
  custom format.
*/

// STILL TO DO:
// - Colors as options in ColorControl + better color choice
// - Undo feature
// - Map Key (legend)

import { localStorageAvailable, getLocalObject, setLocalObject } from "../LocalStorage.js";

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
  // TODO: Move colors to options
  colors: [],
  names: [],
  buttons: [],
  selectedColor: 0,
  edit: {},

  initialize: function (options) {
    var index = 0;
    for (name in options.colors) {
      this.colors[index] = options.colors[name];
      this.names[index] = name;
      index ++;
    }
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

    this.fire("colorselect", {
      color: this.colors[number],
      number: number, });
  },

  layerSelected: function(layer) {
    this.selectColor(layer.attributes.color);
  },

  onAdd: function(map) {
    var div = L.DomUtil.create('div', 'leaflet-bar localdrawoverlay-bar');

    for (let i = 0; i < this.colors.length; i++) {
      this.buttons[i] = L.DomUtil.create('div',
        'localdrawoverlay-button localdrawoverlay-color-button', div);
      this.buttons[i].style["background-color"] = this.colors[i];
      L.DomEvent.on(this.buttons[i], 'click',
        function(e) { this.selectColorNumber(i); }, this);
    }
    this.selectColorNumber(this.selectedColor);

    // Shape color -> color control
    if (this.options.edit.featureGroup != null)
      this.options.edit.featureGroup.on("layerselect", this.layerSelected, this);

    return div;
  },

  onRemove: function(map) {
    if (this.options.edit.featureGroup != null)
      this.options.edit.featureGroup.off("layerselect", this.layerSelected);
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

var DeleteControl = L.Control.extend({
  initialize: function (options) {
    if (options) {
      L.setOptions(this, options)
    }
  },

  _checkDeleteEnabled: function(layer) {
    this.deleteEnabled = (this.options.featureGroup != null &&
        this.options.featureGroup.getSelectedLayer() != null)

    if (this.deleteEnabled)
      this.deletebutton.classList.remove("disabled");
    else
      this.deletebutton.classList.add("disabled");
  },

  _clickDelete: function(e) {
    if (this.options.featureGroup != null &&
        this.options.featureGroup.getSelectedLayer() != null) {
      this.options.featureGroup.removeLayer(this.options.featureGroup.getSelectedLayer());
      this.options.featureGroup.save();
    }
  },

  onAdd: function(map) {
    var div = L.DomUtil.create('div', 'leaflet-bar localdrawoverlay-bar');
    this.deletebutton = L.DomUtil.create('div',
      'localdrawoverlay-button localdrawoverlay-delete-button', div);

    L.DomEvent.on(this.deletebutton, 'click', this._clickDelete, this);
    this._checkDeleteEnabled();
    if (this.options.featureGroup != null)
      this.options.featureGroup.on('layerselect layerunselect',
          this._checkDeleteEnabled, this);
    return div;
  },

  onRemove: function(map) {
    if (this.options.featureGroup != null)
      this.options.featureGroup.off('layerselect layerunselect',
          this._checkDeleteEnabled, this);
  },
})

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
        colors: {
          "Rouge"  : "#e60000",
          "Orange" : "#ffa612",
          "Jaune"  : "#fff600",
          "Vert"   : "#35fb1a",
          "Bleu"   : "#0043ff",
          "Violet" : "#bf00ff",
        },
      });
      this.deleteControl = new DeleteControl({
        position:'topleft',
        featureGroup: this,
      });
      this.colorControl.on("colorselect", this.colorSelected, this);
    } else {
      console.error("Local storage not available for LocalDraw layer.")
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
      this._updateStyle(this.selected_layer);
    }
    // Change draw control color
    if (this.drawControl != null) {
      this.drawControl.options.draw.polygon.shapeOptions.color = e.color;
      this.drawControl.options.draw.circle.shapeOptions.color = e.color;
      this.drawControl.options.draw.rectangle.shapeOptions.color = e.color;
    }
  },

  _updateStyle:function(layer) {
    if (layer == this.selected_layer)
      layer.setStyle({
        color: layer.attributes.color,
        dashArray: '10, 10',
      });
    else
    layer.setStyle({
      color: layer.attributes.color,
      dashArray: null,
      fillOpacity: 0.3,
    });
  },

  getSelectedLayer:function() {
    return this.selected_layer;
  },

  _unselectLayer:function() {
    if (this.selected_layer == null) return;
    var layer = this.selected_layer;
    this.selected_layer = null;
    layer.editing.disable();
    this._updateStyle(layer);
    // TODO: Comment gerer les sauvegardes et annulations ?
    this.save();
    this.fire("layerunselect", layer);
  },

  selectLayer:function(layer) {
    // Unselect
    if (layer == this.selected_layer || !this.hasLayer(layer)) {
      this._unselectLayer();
      return;
    }

    this._unselectLayer();
    this.selected_layer = layer;
    layer.editing.enable();
    layer.bringToFront();
    this._updateStyle(layer);
    this.fire("layerselect", layer);
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

  _load:function() {
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

  removeLayer:function(layer) {
    if (layer == this.selected_layer)
      this._unselectLayer();
    L.FeatureGroup.prototype.removeLayer.call(this, layer);
  },

  addLayer:function(layer) {
    var overlay = this;
    if (layer.attributes == null) {
      layer.attributes = {};
      layer.attributes.color = 'blue';
    }
    this._updateStyle(layer);

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
    if (this.deleteControl != null) map.addControl(this.deleteControl);
    this._load();
  },

  onRemove: function(map) {
    this._unselectLayer();
    map.off("draw:created", this.onDrawCreated, this);
    map.off("draw:edited", this.onDrawEdited, this);
    map.off("draw:deleted", this.onDrawDeleted, this);
    if (this.drawControl != null) map.removeControl(this.drawControl);
    if (this.colorControl != null) map.removeControl(this.colorControl);
    if (this.deleteControl != null) map.removeControl(this.deleteControl);
    this.clearLayers();
  },
});
