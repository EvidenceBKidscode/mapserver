/* exported SearchService */
/* globals SearchStore: true */

var SearchService = {

    search: function(){
      SearchStore.show = true;
      this.fetchData();
    },

    fetchData: function(){
      SearchStore.result = [];

      if (!SearchStore.query){
        return;
      }

      SearchStore.busy = true;

      function searchFor(type, key, valuelike){
        return api.getMapObjects({
          pos1: { x:-2048, y:-2048, z:-2048 },
          pos2: { x:2048, y:2048, z:2048 },
          type: type,
          attributelike: {
            key: key,
            value: "%" + valuelike +"%"
          }
        });
      }

      var prom_list = [
        searchFor("shop", "out_item", SearchStore.query),
        searchFor("poi", "name", SearchStore.query),
        searchFor("train", "station", SearchStore.query),
        searchFor("travelnet", "station_name", SearchStore.query),
        searchFor("bones", "owner", SearchStore.query),
        searchFor("locator", "name", SearchStore.query)
      ];

      Promise.all(prom_list)
      .then(function(results){

        var arr = [];
        results.forEach(function(r) {
          arr = arr.concat(r);
        });

        SearchStore.result = arr;
        SearchStore.busy = false;
      });

    },

    clear: function(){
      SearchStore.result = [];
      SearchStore.show = false;
    }
};
