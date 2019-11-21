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
      options:layer.options,
    };
  }
  if (layer instanceof L.Polygon)
    return {
      type: 'polygon',
      LatLngs:layer.getLatLngs(),
      options:layer.options,
    };

  else if (layer instanceof L.Circle)
    return {
      type: 'circle',
      LatLngs:layer.getLatLng(),
      Radius:layer.getRadius(),
      options:layer.options,
    };

  console.log("Error layer with unknown type could not be stored:");
  console.log(layer);
};

function storableToLayer(storable) {
  switch (storable.type) {
    case 'rectangle':
      return L.rectangle(storable.Bounds, storable.options);
      break;
    case 'polygon':
      return L.polygon(storable.LatLngs, storable.options);
      break;
    case 'circle':
      return L.circle(storable.LatLngs, storable.Radius, storable.options);
      break;
    default:
      console.log("Unknown stored layer type: "+storable.type);
  }
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

export default L.FeatureGroup.extend({
  initialize: function() {
    L.FeatureGroup.prototype.initialize.call(this);
    if (localStorageAvailable()) {
      this.drawControl = new L.Control.Draw({
        position: 'topleft',
        edit: {
          featureGroup: this,
        },
        draw: {
          polygon: {
            allowIntersection: false, // Restricts shapes to simple polygons
            shapeOptions: {
              color: '#00ff0080',
              stroke:true,
            }
          },
          circle: {
            shapeOptions: {
              color: '#00ff0080',
              stroke:true,
            }
          },
          rectangle: {
            shapeOptions: {
              color: '#00ff0080',
              stroke:true,
            }
          },
          polyline: false,
          marker: false,
          circlemarker: false,
        },
      })
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

  style:function() {
    this.setStyle(function(feature) {
      return {
          fillColor: feature.properties.fillColor,
          color: feature.properties.strokeColor,
        };
      }
    );
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
//      this.style();
    }
  },

  onDrawEdited:function(e) {
    console.log("onDrawEdited");
    console.log(e);
    var overlay = this;
    e.layers.eachLayer(function (layer) {
      overlay.addLayer(layer);
    });
    this.save();
  },

  onDrawCreated: function(e) {
    var storable = layerToStorable(e.layer);
    console.log(storable);
    if (storable == null) return;
    var layer = storableToLayer(storable);
    console.log(layer);
    if (layer == null) return;
    this.addLayer(layer);
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
    map.addControl(this.drawControl);
    this.load();
  },

  onRemove: function(map) {
    this.clearLayers();
    map.off("draw:created", this.onDrawCreated, this);
    map.off("draw:edited", this.onDrawEdited, this);
    map.off("draw:deleted", this.onDrawDeleted, this);
    map.removeControl(this.drawControl);
  },
});
