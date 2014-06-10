// ractive init  -------------------------------------------------------------

var ractive = new Ractive({
  el: 'app',

  template: '#template',

  data: {}

});

// keep a shadow copy of all values
var values = {};

// true before any changes made, needed for uri observer to not disable
// draft button on page load.
var firstLoad = true;


// utility functions  --------------------------------------------------------

// deleteQuery generates the SPARQL query to remove a resource from the graph.
var deleteQuery = function( published ) {
  var graph = published ? ractive.data.publicGraph : ractive.data.draftsGraph;
  return 'DELETE { GRAPH ' + graph + ' { ' +
    ractive.get( 'existingURI' ) + ' ?p ?o } }\n' +
    'WHERE { ' + ractive.get( 'existingURI' ) + ' ?p ?o }';
};

// insertQuery generates the SPARQL query to insert the resource into the graph.
var insertQuery = function( publish ) {
  var uri = ractive.get( 'overview.uri' );
  var now = new Date();
  var meta = [
    { 'p': 'a', 'o': ractive.data.overview.type },
    { 'p': internalPred( 'profile' ), 'o': '"' + urlParams.profile + '"' },
    { 'p': internalPred( 'displayLabel' ), 'o': ractive.data.overview.displayLabel },
    { 'p': internalPred( 'searchLabel' ), 'o': ractive.data.overview.searchLabel },
    { 'p': internalPred( 'updated' ), 'o': dateFormat( now.toISOString() ) }
  ];
  if ( !ractive.data.uriFn ) {
    meta.push( { 'p': internalPred( 'id' ), 'o': ractive.get( 'overview.idNumber' ) } );
  }
  if ( ractive.data.existingResource ) {
    meta.push( {"p": internalPred( "created" ),
                "o": dateFormat(ractive.data.overview.created) } );
    if ( ractive.data.overview.published && publish ) {
      meta.push( {"p": internalPred( "published" ),
                  "o": dateFormat( ractive.data.overview.published ) } );
    }
  } else {
    meta.push( { 'p': internalPred( 'created' ), "o": dateFormat( now.toISOString() ) } );
  }
  if ( publish && !ractive.data.overview.published ) {
    meta.push( {'p': internalPred( 'published' ), 'o': dateFormat( now.toISOString() ) } );
  }
  metaPreds = _.reduce(meta, function(s, e) {
    return s + uri + " " + e.p + " " + e.o + " .\n";
  }, "");

  var preds = "";
  _.each(values, function(v, k) {
    _.each(v, function(e) {
      preds += uri + " " + e.predicate + " " + e.value + " . \n";
    });
  });
  var graph = publish ? ractive.data.publicGraph : ractive.data.draftsGraph;
  return 'INSERT { GRAPH ' + graph + ' {\n' + metaPreds + preds + '} }';
};

// doQuery sends a SPARQL query the endpoint, and takes a success callback function.
var doQuery = function( query, callback ) {
  var postData = 'query=' + encodeURIComponent( query );
  var req = new XMLHttpRequest();
  req.open( 'POST', '/RDF/resource', true );
  req.setRequestHeader('Content-Type',
                       'application/x-www-form-urlencoded; charset=UTF-8');
  req.onload = function() {
    if ( req.status >= 200 && req.status < 400 ) {
      //console.log( req.responseText );
      callback( JSON.parse( req.responseText ) );
    } else {
      console.log( 'SPARQL endpoint responed with an error' );
      ractive.set( { 'draftDisabled': true, 'deleteDisabled': true, 'publishDisabled': true } );
    }
  };

  req.onerror = function() {
    console.log( 'Failed to execute SPQRL query' );
  };

  req.send( postData );
};

// enqueue sends an uri to a named queue.
// It's currently fire & forget: no success/failure handling.
var enqueue = function ( queue, uri) {
  var postData = 'uri=' + encodeURIComponent( uri );
  var req = new XMLHttpRequest();
  req.open( 'POST', '/queue/' + queue, true) ;
  req.setRequestHeader( 'Content-Type',
                        'application/x-www-form-urlencoded; charset=UTF-8' );
  req.send( postData );
}

// addToIndex sends an uri to the indexing queue.
var addToIndex = function( uri ) {
  enqueue( "add", uri );
}

// removeFromIndex sends an uri to the remove-from-index queue.
var removeFromIndex = function( uri) {
  enqueue( "remove", uri );
}


// event handlers ------------------------------------------------------------

listener = ractive.on({
  saveDraft: function( event ) {
    var published = ractive.get( 'overview.published' ) ? true : false;
    var q;
    if ( ractive.get( 'existingURI' ) ) {
      q = deleteQuery( published ) + ';\n' + insertQuery( false );
    } else {
      q = insertQuery( false );
    }
    doQuery( q, function() {
      // update the index
      addToIndex( ractive.get( 'overview.uri' ) );

      // Forward to saved uri
      setTimeout( function () {
        window.location.replace( window.location.origin +
                                  window.location.pathname +
                                  "?uri=" + trimURI( ractive.get( 'overview.uri' ) ) );
      }, 200);
    });
  },
  publish: function( event ) {
    var published = ractive.get( 'overview.published' ) ? true : false;
    var q;
    if ( ractive.get( 'existingURI' ) ) {
      q = deleteQuery( published ) + ';\n' + insertQuery( true );
    } else {
      q = insertQuery( true, 'forward' );
    }
    doQuery( q, function() {
      // update the index
      addToIndex( ractive.get( 'overview.uri' ) );

      // Forward to saved uri
      setTimeout( function () {
        window.location.replace( window.location.origin +
                                 window.location.pathname +
                                 "?uri=" + trimURI( ractive.get( 'overview.uri' ) ) );
      }, 200);
    } );
  },
  delResource: function( event) {
    var published = ractive.get( 'overview.published' ) ? true : false;
    var q = deleteQuery( published );
    doQuery( q, function() {
      // update the index
      removeFromIndex( ractive.get( 'overview.uri' ) );

      // forward to create new resource
      setTimeout( function () {
        window.location.replace( window.location.origin +
                                    window.location.pathname +
                                    "?profile=" + urlParams.profile );
      }, 200);
    } );
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
      ractive.set( 'searchResults', [] );
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
    if (event.original.keyCode == 27) {
      ractive.fire( "searchBlur", event );
      return
    }
    var q = event.node.value.trim();
    if (q === "") {
      ractive.set( event.keypath + ".searching", false);
      return;
    }

    var searchTypes = ractive.get( event.keypath + '.searchTypes' ).join(',');
    var searchQuery = { "query": {} };
    var query = {};
    var queryData;
    if ( q.length == 1 ) {
      // Do a prefix query if query string is only one character
      query.prefix = { "searchLabel": q };
    } else {
      // Otherwise normal match query (matches ngram size 2-20)
     query.match = { "searchLabel": q };
    }
    // filter the current URI if we're editing a resource
    if ( ractive.get( 'existingResource' ) ) {
      searchQuery.query = {"filtered": { "filter": {"not": {"ids": {"values": [ trimURI( ractive.get( 'overview.uri' ) ) ] } } },
                                         "query": query} };
    } else {
      searchQuery.query = query;
    }
    queryData = JSON.stringify( searchQuery );
    var req = new XMLHttpRequest();
    req.open( 'POST', '/search/public/'+ searchTypes, true) ;
    req.setRequestHeader( 'Content-Type', 'application/json; charset=UTF-8' );

    req.onerror = function( e ) {
      console.log( "failed to reach search endoint: " + e.target.status );
    }

    req.onload = function( e) {
      //console.log( e.target.responseText );
      ractive.set( 'searchResults',
                   JSON.parse( e.target.responseText).hits.hits );
    }

    req.send( queryData );

    ractive.merge( event.keypath + ".searching", true);
  }, 100),
  selectURI: function( event ) {
    var label, uri, predicate, predicateLabel, source;
    label = event.context._source.displayLabel;
    uri = '<' + event.context._source.uri + '>';
    source = 'local';
    var idx = event.index;
    predicate = ractive.data.views[idx.i1].elements[idx.i2].predicates[0].uri;
    predicateLabel = ractive.data.views[idx.i1].elements[idx.i2].predicates[0].label;
    var exsitingURI = _.find(ractive.data.views[idx.i1].elements[idx.i2].values, function( e ) {
      return e.value === uri;
    });

    if ( !exsitingURI ) {
      ractive.data.views[idx.i1].elements[idx.i2].values.push(
        {"predicate": predicate, "predicateLabel": predicateLabel, "value": uri,
         "URILabel": label, "source": source});
    }

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
  validateInteger: function(event) {
    var value = event.node.value.trim();

    // validate integer
    if ( !value.match(/^[0-9]+$/) ) {
      ractive.merge( event.keypath + ".errorInfo",
                    "ugyldig verdi: må være et heltall" );
      setTimeout( function () {
        event.node.focus();
      }, 0 );
      return;
    }

    ractive.fire( "newValue", event );
  },
  validateGYear: function(event) {
    var value = event.node.value.trim();

    // validate float
    if ( !value.match(/^[0-9]{1,4}$/) ) {
      ractive.merge( event.keypath + ".errorInfo",
                    "ugyldig verdi: må være et årstall" );
      setTimeout( function () {
        event.node.focus();
      }, 0 );
      return;
    }

    // add xsd:gYear datatype
    event.node.value = '"' + event.node.value + '"^^xsd:gYear';

    ractive.fire( "newValue", event );
  },
  validateString: function( event ) {
    event.node.value = '"' + event.node.value + '"';
    ractive.fire( 'newValue', event );
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
  // keep values in sync
  values = {};
  var missingValues = false;
  newValue.forEach(function(view, i) {
    view.elements.forEach(function(elem, j) {
      if (!values[elem.id]) {
        values[elem.id] = [];
      }
      // Check if all required attributes in the schema has a value
      if ( elem.required && elem.values.length === 0 ) {
        missingValues = true;
      }
      elem.values.forEach(function(v) {
        values[elem.id].push(v);
      });
    });
  });
  // Don't allow publish a resource with misssing required values.
  if ( missingValues ) {
    ractive.set( 'publishDisabled', true );
  } else if ( !ractive.get( 'duplicateURI') ) {
    ractive.set( { 'publishDisabled': false, 'draftDisabled': false } );
  }


  // Use uriFn if it exists
  if ( ractive.get('uriFn') ) {
    var createURI = _.every(ractive.data.uriNeedIds, function(id) {
      return ( values[id].length > 0 );
    });
    if ( createURI) {
      // got all needed values to generate uri
      ractive.set( 'overview.uri', ractive.data.uriFn( values ) );
    } else {
      ractive.set( 'overview.uri', '' );
    }
  }

  firstLoad = false; //
  var sl = ractive.data.searchLabel(values);
  var dl = ractive.data.displayLabel(values);

   // create searchLabel and displayLabel
  ractive.set( 'overview.searchLabel', sl);
  ractive.set( 'overview.displayLabel', dl);
});

ractive.observe( 'overview.uri', function( newURI, oldURI, keyPath ) {
  if ( ractive.get( 'uriFn') ) {
    // Check if URI allready exists in local RDF repo.
    if ( newURI !== "" && newURI !== ractive.get( 'existingURI' ) ) {
      var q = 'ASK WHERE { ' + newURI + '?p ?o }';
      doQuery( q, function( data) {
        var exists = data.boolean;
        ractive.set( 'duplicateURI', exists );
        ractive.set( 'draftDisabled', exists );
        if ( exists && !ractive.get( 'publishDisabled') ) {
          ractive.set( 'publishDisabled', true );
        }
      });
    } else {
      ractive.set( 'duplicateURI', false );
    }
    if ( newURI === "" && !firstLoad ) {
      ractive.set( 'draftDisabled', true );
    }
    // notify user if URI has changed
    ractive.set( 'changedURI', ( ractive.get( 'existingURI' ) && ractive.get( 'existingURI' ) != newURI && newURI !== "" ) );
  }
});

// load profile and (optionally) resource data -------------------------------

// loadScript dynamically loads a javascript file, executing callback on success.
var loadScript = function( src, callback ) {
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

// Get the query parameters.
var urlParams;
(window.onpopstate = function () {
  var match,
      pl     = /\+/g,
      search = /([^&=]+)=?([^&]*)/g,
      decode = function (s) { return decodeURIComponent(s.replace(pl, ' ')); },
      query  = window.location.search.substring(1);

  urlParams = {};
  while ( match = search.exec(query) )
    urlParams[decode(match[1])] = decode(match[2]);
})();

// createSchema creates a schema acording to a loaded profile.
var createSchema = function() {
  // set values to empty array
  profile.views.forEach(function(view, i) {
    view.elements.forEach(function(elem, j) {
      profile.views[i].elements[j].values = [];
    });
  });
  ractive.set(_.extend(profile, common));

  if (urlParams.uri) {
    // loading an existing resource
    return
  }

  // Get a ID number for new resource
  var req = new XMLHttpRequest();
  req.open( 'GET', '/id/'+ urlParams.profile + '?nocacheplease=' + new Date().valueOf(), true) ;

  req.onerror = function( e ) {
      console.log( "failed to reach idService endpoint: " + e.target.status );
  }

  req.onload = function( e) {
    var id = e.target.responseText;
    ractive.set( 'overview.idNumber', id );
    ractive.set( 'overview.uri', '<http://data.deichman.no/' +
                                 urlParams.profile + '/' + id +'>' );
    ractive.update();
  }

  req.send();
}

// Load resource if uri query parameter is given.
if ( urlParams.uri ) {
  ractive.set( 'existingResource', true );
  ractive.set( 'existingURI', "<" + urlParams.uri + ">" );

  var q = 'SELECT * WHERE { { <' + urlParams.uri + '> ?p ?o } UNION ' +
          '{ <' + urlParams.uri + '> ?p ?o .' +
          ' ?o ' + internalPred( 'displayLabel') + ' ?l } }';
  doQuery( q, function( rdfRes ) {
    // If the SPARQL query returns an empty set, forward to create new resource page.
    // TODO display flash message 'resource not found' for the user
    if ( rdfRes.results.bindings.length === 0 && urlParams.profile ) {
      window.location.replace( window.location.origin +
                              window.location.pathname +
                              "?profile=" + urlParams.profile ); // TODO urlParams.profile is not set here
    }

    // get the profile from the response
    var p = "";
    rdfRes.results.bindings.forEach(function(b) {
      if ( b.p.value ===  trimURI( internalPred( 'profile' ) ) ) {
        p = b.o.value;
        urlParams.profile = p;
      }
    });

    if ( p === "" ) {
      console.log( 'ERROR: no profile in resource.' );
    } else {
      loadScript( '/public/profiles/' + p + ".js", function() {
        createSchema();
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
          if ( b.type === 'typed-literal' ) {
            return b.value;
          }
          if ( b.type === 'uri' ) {
            return '<' + b.value + '>';
          }
          if ( b.type === 'literal' ) {
            if ( b['xml:lang'] ) {
              return '"' + b.value + '"@' + b["xml:lang"];
            }
            return '"' + b.value + '"';
          }
        };

        // Find bindings which are labels for URIs
        var uriLabels = {};
        rdfRes.results.bindings.forEach( function( b ) {
          if ( b.l ) {
            uriLabels['<' + b.o.value + '>'] = b.l.value;
          }
        });

        rdfRes.results.bindings.forEach(function(b) {
          var pred = "<" + b.p.value + ">";
          var source = 'local';
          var kp = findElement(pred);
          if ( kp ) {
            var v = getValue(b.o);
            var predLabel = ractive.get(kp).label;
            var res = {"predicate": pred, "predicateLabel": predLabel, "value": v, "source": source}
            if ( uriLabels[v] ) {
              res.URILabel = uriLabels[v]
            }
            if ( !b.l ) { // skip URI label bindings
              ractive.get(kp + ".values").push(res);
            }
          } else {
            switch ( pred ) {
              case internalPred( 'displayLabel' ):
                ractive.set( 'overview.displayLabel', getValue( b.o ) );
                break;
              case internalPred( 'searchLabel' ):
                ractive.set( 'overview.searchLabel', getValue( b.o ) );
                break;
              case internalPred( 'created' ):
                ractive.set( 'overview.created', getValue( b.o ) );
                break;
              case internalPred( 'updated' ):
                ractive.set( 'overview.updated', getValue( b.o ) );
                break;
              case internalPred( 'published' ):
                ractive.set( 'overview.published', getValue( b.o ) );
                break;
              case internalPred( 'id' ):
                ractive.set( 'overview.idNumber', getValue( b.o ) );
                ractive.set( 'overview.uri', '<' + urlParams.uri + '>' );
                break;
            }
          }
        }); // end rdfRes.results.bindings.forEach
      }); // end loadScript
    } // end p (profile) !== ""
  });
  ractive.set( { 'draftDisabled': false, 'deleteDisabled': false, 'publishDisabled': false } );
}


// Load profile if supplied in query string, and not allready fetched via
// loaded resource.
// TODO onerror: what if profile is not found?
if ( urlParams.profile && !urlParams.uri ) {
  // No URI given; assuming creating a new resource.
  ractive.set('existingResource', false);
  loadScript( '/public/profiles/' + urlParams.profile + ".js", createSchema );
  ractive.set( { 'draftDisabled': true, 'deleteDisabled': true, 'publishDisabled': true } );


}

