/*
  NOTE: GeoJSON is unable to store circles. So, as we need to store and edit
  any shapes, this LocalDrawOverlay will not use GeoJSON for storage but a
  custom format.
*/

// STILL TO DO:
// - Colors as options in ColorControl + better color choice
// - Undo feature
// - Map Key (legend)

import "../../lib/tinycolor.js";
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

  if (layer instanceof L.Circle)
    return {
      type: 'circle',
      LatLngs:layer.getLatLng(),
      Radius:layer.getRadius(),
      attributes:layer.attributes,
    };

  if (layer instanceof L.Marker)
    return  {
      type: 'marker',
      LatLng:layer.getLatLng(),
      attributes:layer.attributes,
    };

  console.log("Error layer with unknown type could not be stored:");
  console.log(layer);
};

function storableToLayer(storable) {
  var layer = null;
  switch (storable.type) {
    case 'rectangle':
      layer = L.rectangle(storable.Bounds, { bubblingMouseEvents:false, });
      break;
    case 'polygon':
      layer = L.polygon(storable.LatLngs, { bubblingMouseEvents:false, });
      break;
    case 'circle':
      layer = L.circle(storable.LatLngs, storable.Radius, { bubblingMouseEvents:false, });
      break;
    case 'marker':
      layer = L.marker(storable.LatLng, { bubblingMouseEvents:false, });
      break;
    default:
      console.log("Unknown stored layer type: " + storable.type);
      return null;
  }
  if (storable.attributes == null) {
      if (layer instanceof L.path)
        layer.attributes = { color: '#fff' };
      if (layer instanceof L.marker)
        layer.attributes = { markerType: 'unknown' };
  } else {
    layer.attributes = storable.attributes;
  }
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
        marker: 'Dessiner du texte',
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
          start: 'Cliquer sur la carte pour ajouter du texte.'
         //start: 'Cliquer sur la carte pour placer un marqueur.'
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
  _colors: [],
  _names: [],
  _buttons: [],
  _selectedColor: 0,
  edit: {},

  initialize: function (options) {
    var index = 0;
    for (name in options.colors) {
      this._colors[index] = options.colors[name];
      this._names[index] = name;
      index ++;
    }
    if (options) {
      L.setOptions(this, options)
    }
  },

  getSelectedColor: function() {
    return this._colors[this._selectedColor];
  },

  selectColor: function(name) {
    for (let i = 0; i < this._colors.length; i++)
      if (this._colors[i] == name) {
        this.selectColorNumber(i);
        return;
      }
  },

  selectColorNumber: function(number) {
    if (number < 0 || number >= this._buttons.length)
      return;

    for (let i = 0; i < this._buttons.length; i++)
      this._buttons[i].classList.remove("selected");

    this._buttons[number].classList.add("selected");
    this._selectedColor = number;

    this.fire("colorselect", { color: this._colors[number], number: number, });
  },

  layerSelected: function(layer) {
    this.selectColor(layer.attributes.color);
  },

  _createButton: function(title, bgcolor, container, fn) {
    var link = L.DomUtil.create('a',
      'localdrawoverlay-button', container);
    link.href = '#';
    link.title = title;

    link.setAttribute('role', 'button');
    link.setAttribute('aria-label', title);
    var div = L.DomUtil.create('div', 'localdrawoverlay-color-button', link);
    div.style["background-color"] = bgcolor;
    L.DomEvent.disableClickPropagation(link);
    L.DomEvent.on(link, 'click', L.DomEvent.stop);
    L.DomEvent.on(link, 'click', fn, this);
    L.DomEvent.on(link, 'click', this._refocusOnMap, this);
    return link;
  },

  onAdd: function(map) {
    var div = L.DomUtil.create('div', 'localdrawoverlay-bar leaflet-bar');
    L.DomEvent.disableClickPropagation(div);

    for (let i = 0; i < this._colors.length; i++) {
      this._buttons[i] = this._createButton(this._names[i], this._colors[i], div,
        function(e) { this.selectColorNumber(i); });
    }
    this.selectColorNumber(this._selectedColor);

    // Shape color -> color control
    if (this.options.featureGroup != null)
      this.options.featureGroup.on("layerselect", this.layerSelected, this);

    return div;
  },

  onRemove: function(map) {
    if (this.options.featureGroup != null)
      this.options.featureGroup.off("layerselect", this.layerSelected);

    for (let i = 0; i < this._buttons.length; i++) {
      L.DomEvent.off(this._buttons[i], 'click', DomEvent.stop);
      L.DomEvent.off(this._buttons[i], 'click', fn);
      L.DomEvent.off(this._buttons[i], 'click', this._refocusOnMap);
    }
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
    this._deleteEnabled = (this.options.featureGroup != null &&
        this.options.featureGroup.getSelectedLayer() != null)

    if (this._deleteEnabled) {
      this._deletebutton.classList.remove("disabled");
      this._deletebutton.title = "Supprimer la forme sélectionnée";
    } else {
      this._deletebutton.classList.add("disabled");
      this._deletebutton.title = "Pas de forme sélectionnée";
    }
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
    L.DomEvent.disableClickPropagation(div);

    var link = L.DomUtil.create('a',
      'localdrawoverlay-button localdrawoverlay-delete-button', div);
    link.href = '#';
    link.setAttribute('role', 'button');
    link.setAttribute('aria-label', link.title);
    L.DomEvent.disableClickPropagation(link);
    L.DomEvent.on(link, 'click', L.DomEvent.stop);
    L.DomEvent.on(link, 'click', this._clickDelete, this);
    L.DomEvent.on(link, 'click', this._refocusOnMap, this);
    this._deletebutton = link;
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
        L.DomEvent.off(this._deletebutton, 'click', L.DomEvent.stop);
        L.DomEvent.off(this._deletebutton, 'click', this._clickDelete);
        L.DomEvent.off(this._deletebutton, 'click', this._refocusOnMap);
  },
})

// Legende

var LegendControl = L.Control.extend({
  initialize: function (options) {
    if (options) {
      L.setOptions(this, options);
    }
    this._colors = {};
  },

  _keypressedField: function(e) {
    if (e.code == "Enter" || e.code == "Escape")
      this._validateField(e);
  },

  _validateField: function(e) {
    for (name in this._colors) {
      var color = this._colors[name];
      if (color.dom != null) {
        L.DomUtil.removeClass(color.dom.statictext,
          'localdrawoverlay-legend-statictext-hidden');
        L.DomUtil.addClass(color.dom.textfield,
          'localdrawoverlay-legend-textfield-hidden');
        if (color.dom.textfield == e.target) {
          color.text = color.dom.textfield.value;
        }
      }
    }
    this._save();
    this._update();
  },

  _focusField: function(e) {
    for (name in this._colors) {
      var color = this._colors[name];
      if (this._colors[name].dom) {
        if (color.dom.entry.contains(e.target)) {
          L.DomUtil.addClass(color.dom.statictext,
            'localdrawoverlay-legend-statictext-hidden');
          L.DomUtil.removeClass(color.dom.textfield,
            'localdrawoverlay-legend-textfield-hidden');
          if (color.text != null)
            color.dom.textfield.value = color.text;
          color.dom.textfield.focus();
        } else {
          L.DomUtil.removeClass(color.dom.statictext,
            'localdrawoverlay-legend-statictext-hidden');
          L.DomUtil.addClass(color.dom.textfield,
            'localdrawoverlay-legend-textfield-hidden');
        }
      }
    }
  },

  _save: function() {
    if (localStorageAvailable()) {
      var storage = {};
      for (name in this._colors)
        if (this._colors[name].text != null)
          storage[name] = this._colors[name].text;
      setLocalObject("mylegend", JSON.stringify(storage));
    }
  },

  _load:function() {
    if (localStorageAvailable()) {
      var storage = JSON.parse(getLocalObject("mylegend"));
      for (name in storage) {
        if (this._colors[name] == null)
          this._colors[name] = {
            tiny:tinycolor(name),
          };
        this._colors[name].text = storage[name];
      }
    }
  },

  _update: function() {
    if (! this.options.featureGroup instanceof L.FeatureGroup)
      return;

    // Reset color visibility
    for (name in this._colors)
      this._colors[name].visible = false;

    // Make visible colors realy in use
    this.options.featureGroup.eachLayer(function(layer) {
      // Only path color, not markers
      if (layer instanceof L.Path) {
        if (this._colors[layer.attributes.color] == null)
          this._colors[layer.attributes.color] = {
            tiny:tinycolor(layer.attributes.color)
          };
        this._colors[layer.attributes.color].visible = true;
      }
    }, this);

    for (name in this._colors) {
      var color = this._colors[name];

      // Add missing dom elements for new colors
      if (color.dom == null) {
        color.dom = {}
        color.dom.entry = L.DomUtil.create('div',
          'localdrawoverlay-legend-entry', this._div);

        var sample = L.DomUtil.create('div',
           'localdrawoverlay-legend-color', color.dom.entry);
        sample.style["background-color"] = color.tiny.setAlpha(0.3).toHex8String();
        sample.style["borderColor"] = color.tiny.setAlpha(1.0).toHex8String();

        color.dom.statictext = L.DomUtil.create('div',
          'localdrawoverlay-legend-statictext', color.dom.entry);

        color.dom.textfield = L.DomUtil.create('input',
          'localdrawoverlay-legend-textfield localdrawoverlay-legend-textfield-hidden',
          color.dom.entry);
        color.dom.textfield.type = "text";
        L.DomEvent.on(color.dom.entry, 'click', this._focusField, this);
        L.DomEvent.on(color.dom.textfield, 'blur', this._validateField, this);
        L.DomEvent.on(color.dom.textfield, 'keydown', this._keypressedField, this);
      };

      // Set text
      if (color.text == null || color.text == "")
        color.dom.statictext.innerHTML = "<span class='empty'>Ajouter du texte</span>";
      else
        color.dom.statictext.innerHTML = color.text;

      // Show/hide wanted lines
      if (color.visible)
        color.dom.entry.style["display"] = "block";
      else
        color.dom.entry.style["display"] = "none";
    }
  },

  onAdd: function(map) {
    if (this._div == null)
      this._div = L.DomUtil.create('div', 'localdrawoverlay-legend-box');
    L.DomEvent.disableClickPropagation(this._div);
    this._load();
    this._update();

    if (this.options.featureGroup instanceof L.FeatureGroup) {
      this.options.featureGroup.on("layeradd", this._update, this);
      this.options.featureGroup.on("layerchange", this._update, this);
      this.options.featureGroup.on("layerremove", this._update, this);
    }
    return this._div;
  },

  onRemove: function(map) {
    if (this.options.featureGroup instanceof L.FeatureGroup) {
      this.options.featureGroup.off("layeradd", this._update, this);
      this.options.featureGroup.off("layerchange", this._update, this);
      this.options.featureGroup.off("layerremove", this._update, this);
    }
  },
});

/* Taken from a newer version of leafletjs */

var DivIcon = L.Icon.extend({
  options: {
    iconSize: null,
    html: false,
    bgPos: null,
    className: 'leaflet-div-icon',
  },

  createIcon: function (oldIcon) {
    var div = (oldIcon && oldIcon.tagName === 'DIV') ? oldIcon : document.createElement('div'),
        options = this.options;

    if (options.html instanceof Element) {
      L.DomUtil.empty(div);
      div.appendChild(options.html);
    } else {
      div.innerHTML = options.html !== false ? options.html : '';
    }

    if (options.bgPos) {
      var bgPos = point(options.bgPos);
      div.style.backgroundPosition = (-bgPos.x) + 'px ' + (-bgPos.y) + 'px';
    }
    this._setIconStyles(div, 'icon');

    return div;
  },

  createShadow: function () {
    return null;
  }
});

////////////////////////////////////////////////////////////////////////////////

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
          marker: {
            icon: new L.Icon({
              iconUrl: 'pics/marker-text.png',
              iconSize: [24, 24],
               iconAnchor: [24, 24],
            }),
          },
          circlemarker: false,
        },
        _selected_layer: null,
        _edited_layer: null,
      });

      this.colorControl = new ColorControl({
        position:'topleft',
        featureGroup: this,
        colors: {
          "Rouge"  : "#e60000",
          "Orange" : "#ffa612",
          "Jaune"  : "#fff600",
          "Vert"   : "#35fb1a",
          "Bleu"   : "#0043ff",
          "Violet" : "#bf00ff",
        },
      });
      this.colorControl.on("colorselect", this.colorSelected, this);

      this.deleteControl = new DeleteControl({
        position:'topleft',
        featureGroup: this,
      });

      this.legendControl = new LegendControl({
        position:'bottomright',
        featureGroup: this,
      });

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
    if (this._selected_layer != null &&
        this._selected_layer instanceof L.Path &&
        this._selected_layer.attributes.color != e.color) {
      this._selected_layer.attributes.color = e.color;
      this._updateStyle(this._selected_layer);
      this.fire("layerchange", this._selected_layer);
    }
    // Change draw control color
    if (this.drawControl != null) {
      this.drawControl.options.draw.polygon.shapeOptions.color = e.color;
      this.drawControl.options.draw.circle.shapeOptions.color = e.color;
      this.drawControl.options.draw.rectangle.shapeOptions.color = e.color;
    }
  },

  _updateStyle:function(layer) {
    if (layer instanceof L.Path) {
      if (layer == this._selected_layer)
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
    }
  },

  _setIcon:function(layer) {
    if (layer instanceof L.Marker &&
        layer.attributes.markerType == 'text') {
      var div = L.DomUtil.create("div", "");
      div.innerHTML = layer.attributes.text;
      layer.setIcon(new DivIcon({
        className: 'localdrawoverlay-text',
        html: div,
        iconSize: null,
      }));
    }
  },

  getSelectedLayer:function() {
    return this._selected_layer;
  },

  _unselectLayer:function() {
    if (this._selected_layer == null) return;

    var layer = this._selected_layer;
    this._selected_layer = null

    layer.editing.disable();
    this._updateStyle(layer);

    if (layer == this._edited_layer)
      this._endEdit();

    // TODO: Comment gerer les sauvegardes et annulations ?
    this.save();
    this.fire("layerunselect", layer);
},

  selectLayer:function(layer) {
    // Unselect
    if (layer == this._selected_layer || !this.hasLayer(layer)) {
      this._unselectLayer();
      return;
    }

    this._unselectLayer();
    this._selected_layer = layer;
    if (layer instanceof L.Path) {
      layer.bringToFront();
      this._updateStyle(layer);
    }
    layer.editing.enable();
    this.fire("layerselect", layer);
  },

  _endEdit:function(save = true) {
    if (this._edited_layer) {
      var layer = this._edited_layer;
      var value = this._edit_field.value;
      L.DomEvent.off(this._edit_field, 'blur', this._endEdit);
      L.DomEvent.off(this._edit_field, 'keydown', this._keypressedField);
      this._edit_field = null;
      this._edited_layer = null;
      this._selected_layer = null;
      layer.editing.disable();

      if (save) {
        // Remove marker if text empty
        if (value == "") {
          this.removeLayer(layer);
          this.save();
          return;
        }
        layer.attributes.text = value;
        this.save();
      }
      this._setIcon(layer);
    }
  },

  _keypressedField: function(e) {
    if (e.code == "Enter")
      this._endEdit(true);
    if (e.code == "Escape")
      this._endEdit(false);
  },

  _editLayer:function(layer) {
    this._endEdit();

    if (layer instanceof L.Marker && layer.attributes.markerType == 'text') {
      this._edited_layer = layer;
      var content = L.DomUtil.create("input", "");
      L.DomEvent.disableClickPropagation(content);
      content.type = "text";
      content.value = layer.attributes.text;
      layer.setIcon(new DivIcon({
        className: 'localdrawoverlay-text',
        html: content,
        iconSize: null,
      }));
      L.DomEvent.on(content, 'blur', this._endEdit, this);
      L.DomEvent.on(content, 'keydown', this._keypressedField, this);
      this._edit_field = content;
      content.focus();
    }
  },

  save:function() {
    if (localStorageAvailable()) {
      var storage = []
      this.eachLayer(function (layer) {
        var storable = layerToStorable(layer);
        if (storable != null)
          storage.push(storable);
      });
      setLocalObject("mydraw", JSON.stringify(storage));
    }
  },

  _load:function() {
    if (localStorageAvailable()) {
      this.clearLayers();
      var overlay = this
      var storage = JSON.parse(getLocalObject("mydraw"));
      if (storage != null)
        storage.forEach(function(storable) {
          var layer = storableToLayer(storable);
          if (layer != null) {
            overlay.addLayer(layer);
          }
        });
    }
  },

  removeLayer:function(layer) {
    if (layer == this._selected_layer)
      this._unselectLayer();
    L.FeatureGroup.prototype.removeLayer.call(this, layer);
  },

  addLayer:function(layer) {
    if (layer.attributes == null) {
      if (layer instanceof L.path)
        layer.attributes = { color: '#fff' };
      if (layer instanceof L.marker)
        layer.attributes = { markerType: 'unknown' };
    }
    this._updateStyle(layer);
    this._setIcon(layer);

    // Select layer on click
    layer.on('click', function(e) { this.selectLayer(e.target); }, this);
    layer.on('dblclick', function(e) { this._editLayer(e.target); }, this);
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
    if (layer instanceof L.Path) {
      layer.attributes.color = this.colorControl.getSelectedColor();
    }
    if (layer instanceof L.Marker) {
      layer.attributes.markerType = 'text';
      layer.attributes.text = "";
      this.addLayer(layer);
      this.selectLayer(layer);
      this._editLayer(layer);
    } else {
      this.addLayer(layer);
      this.selectLayer(layer);
    }
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
    map.on("click", this._unselectLayer, this);

    if (this.drawControl != null) map.addControl(this.drawControl);
    if (this.colorControl != null) map.addControl(this.colorControl);
    if (this.deleteControl != null) map.addControl(this.deleteControl);
    if (this.legendControl != null) map.addControl(this.legendControl);
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
    if (this.legendControl != null) map.removeControl(this.legendControl);

    this.clearLayers();
  },
});
