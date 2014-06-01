
// ractive init  -------------------------------------------------------------

var ractive = new Ractive({
  el: 'app',

  template: '#template',

  data: {}

});

// keep a shadow copy of all values
var values = {};

// event handlers ------------------------------------------------------------

listener = ractive.on({
  previewSparql: function(event) {
    var uri = ractive.data.overview.uri;
    var preds = "";
    _.each(values, function(v, k) {
      _.each(v, function(e) {
        preds += uri + " " + e.predicate + " " + e.value + " . \n";
      });
    });
    if (ractive.data.existingResource) {
      console.log("WITH " + ractive.data.draftsGraph +
                  "\nDELETE { " + uri + " ?p ?o }\n" +
                  "INSERT {\n" + preds + "}\n" +
                  "WHERE { " + ractive.data.existingURI + " ?p ?o }");
    } else {
      console.log("INSERT INTO " + ractive.data.draftsGraph +
                  " {\n" + preds + "}\n");
    }
  },
  remove: function( event ) {
    var idx = event.index;
    ractive.data.views[idx.i1].elements[idx.i2].values.splice(idx.i3, 1);
  },
  searchBlur: function ( event ) {
    // delay a bit so that the on-click event has time to fire in case of searchhit select
    setTimeout( function () {
          event.node.value = "";
          ractive.set( event.keypath + ".searching", false);
        }, 100 );
  },
  newValue: function(event) {
    var value, predicate, predicateLabel, source;
    value = event.node.value.trim();

    predicate = event.context.predicates[0].uri;
    predicateLabel = event.context.predicates[0].label;
    source = 'local';

    var idx = event.index;
    ractive.data.views[idx.i1].elements[idx.i2].values.push(
      {"predicate": predicate, "predicateLabel": predicateLabel, "value": value, "source": source});
    event.node.value = "";
    ractive.merge( event.keypath + ".errorInfo", "");

  },
  searchForURI: _.debounce(function( event) {
    var q = event.node.value.trim();
    if (q === "") {
      ractive.set( event.keypath + ".searching", false);
      return;
    }
    console.log("searching for " + q);
    ractive.merge( event.keypath + ".searching", true);
  }, 100),
  selectURI: function( event ) {
    var label, uri, predicate, predicateLabel, source;
    label = event.context.label;
    uri = event.context.uri;
    source = 'local';

    var idx = event.index;
    predicate = data.views[idx.i1].elements[idx.i2].predicates[0].uri;
    predicateLabel = data.views[idx.i1].elements[idx.i2].predicates[0].label;
    ractive.data.views[idx.i1].elements[idx.i2].values.push(
      {"predicate": predicate, "predicateLabel": predicateLabel, "value": uri,
       "URILabel": label, "source": source});
  },
  validateFloat: function(event) {
    var value = event.node.value.trim();

    // validate float
    if ( !value.match(/^[0-9]+(?:\.[0-9]+)?$/) ) {
      ractive.merge( event.keypath + ".errorInfo",
                     "ugyldig verdi: må være et tall" );
      setTimeout( function () {
          event.node.focus();
        }, 0 );
      return;
    }

    ractive.fire( "newValue", event );
  },
  validateLangString: function( event ) {
    var value, lang;
    var idx = event.index;
    value = event.node.value.trim();
    lang = ractive.data.views[idx.i1].elements[idx.i2].selected;

    // associate language tag if it is chosen
    if ( lang === "") {
      event.node.value = "\"" + event.node.value + "\"";
    } else {
      event.node.value = "\"" + event.node.value + "\"@" + lang;
    }

    ractive.fire("newValue", event);
  }
});

// observers  ----------------------------------------------------------------

ractive.observe('views', function( newValue, oldValue, keypath) {
  values = {};
  newValue.forEach(function(view, i) {
    view.elements.forEach(function(elem, j) {
      if (!values[elem.id]) {
        values[elem.id] = [];
      }
      elem.values.forEach(function(v) {
        values[elem.id].push(v);
      });
    });
  });

  var createURI = _.every(ractive.data.uriNeedIds, function(id) {
    return (values[id].length > 0);
  });

  if (createURI) {
    // create URI if the needed id's are present
    ractive.set('overview.uri', ractive.data.uriFn(values));
  } else {
    // or remove uri if there is one
    ractive.set('overview.uri', "");
  }

  // create searchLabel and displayLabel
  ractive.set('overview.searchLabel', ractive.data.searchLabel(values));
  ractive.set('overview.displayLabel', ractive.data.displayLabel(values));
});

// load profile and (optionally) resource data -------------------------------

var loadScript = function(src, callback) {
  var s = document.createElement('script');
  s.type = 'text/javascript';
  s.async = false;
  s.src = src;

  s.onreadystatechange = s.onload = function() {
    var state = s.readyState;

    if (!callback.done && (!state || /loaded|complete/.test(state))) {
      callback.done = true;
      callback();
    }
  };

  (document.body || document.head).appendChild(s);
};
var urlParams;
(window.onpopstate = function () {
  var match,
      pl     = /\+/g,
      search = /([^&=]+)=?([^&]*)/g,
      decode = function (s) { return decodeURIComponent(s.replace(pl, ' ')); },
      query  = window.location.search.substring(1);

  urlParams = {};
  while (match = search.exec(query))
    urlParams[decode(match[1])] = decode(match[2]);
})();

// Load profile
loadScript('/public/profiles/' + urlParams.profile + ".js", function() {
  // set values to empty array
  profile.views.forEach(function(view, i) {
    view.elements.forEach(function(elem, j) {
      profile.views[i].elements[j].values = [];
    });
  });
  ractive.set(_.extend(profile, common));
  // TODO onerror: what if profile is not found?

  // Populate schema if uri is given
  if ( urlParams.uri ) {
    ractive.set( 'existingResource', true );
    ractive.set( 'existingURI', "<" + urlParams.uri + ">" );

    req = new XMLHttpRequest();
    req.open( 'GET', '/RDF/resource?uri=' + encodeURIComponent(urlParams.uri), true );

    req.onload = function() {
      if (req.status >= 200 && req.status < 400) {
        console.log(req.responseText);

        rdfRes = JSON.parse(req.responseText);

        // findElement returns the keypath of a predicate, or false if no match.
        var findElement = function(pred) {
          var kp = false;
          ractive.data.views.forEach(function(v, i) {
            v.elements.forEach(function(e, j) {
              e.predicates.forEach(function(p, k) {
                if (p.uri === pred) {
                  kp = "views."+i+".elements."+j;
                }
              });
            });
          });
          return kp;
        };

        // getValue returns the value of a binding, including surrounding quotes
        // for strings and language tag if present.
        var getValue = function(b) {
          if ( b.type === 'uri' || b.type === 'typed-literal' ) {
            return b.value;
          }
          if ( b.type === 'literal' ) {
            if ( b['xml:lang'] ) {
              return '"' + b.value + '"@' + b["xml:lang"];
            }
            return '"' + b.value + '"';
          }
        };

        rdfRes.results.bindings.forEach(function(b) {
          var pred = "<" + b.p.value + ">";
          var kp = findElement(pred);
          var v = getValue(b.o);
          var predLabel = ractive.get(kp).label;
          var source = 'local';
          ractive.get(kp + ".values").push(
            {"predicate": pred, "predicateLabel": predLabel, "value": v, "source": source});
        });
      } else {
        console.log("server error");
      }
    };

    req.onerror = function() {
      console.log("connection error");
    };

    req.send();
  } else {
    // No URI given; assuming creating a new resource.
    ractive.set('existingResource', false);
  }

});
