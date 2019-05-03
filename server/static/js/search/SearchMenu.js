/* exported SearchMenu */
/* globals SearchResult: true */
/* globals SearchService: true */
/* globals SearchStore: true */

var SearchMenu = {
  view: function(vnode){

    var style = {};

    if (!SearchStore.show) {
      style.display = "none";
    }

    function close(){
      SearchService.clear();
    }

    function getContent(){
      if (SearchStore.busy){
        return m("div", m("i", { class: "fa fa-spinner"}));
      } else {
        return m(SearchResult, { map: vnode.attrs.map });
      }
    }

    return m("div", { class: "card", id: "search-menu", style: style }, [
      m("div", { class: "card-header" }, [
        m("i", { class: "fa fa-search"}),
        "Search",
        m("i", { class: "fa fa-times float-right", onclick: close }),
      ]),
      m("div", { class: "card-body", style: {overflow: "auto"} }, getContent())
    ]);
  }
};
