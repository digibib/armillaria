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
// TODO find a better way to solve this
var firstLoad = true;


// utility functions  --------------------------------------------------------

var addClass = function( el, className ) {
  if (el.classList) {
    el.classList.add(className);
  }  else {
    el.className += ' ' + className;
  }

}

var removeClass = function( el, className ) {
  if (el.classList) {
    el.classList.remove(className);
  } else {
    el.className = el.className.replace(new RegExp('(^|\\b)' + className.split(' ').join('|') + '(\\b|$)', 'gi'), ' ');
  }
}

// log adds a line to the logLines div
var log = function( msg, isError ) {
  var now = new Date().toISOString().slice(11, 19);
  var t = isError ? "error" : "normal";
  if ( msg.length > 120 ) {
    msg = msg.substr(0, 120) + "...";
  }
  ractive.data.logLines.push({
    "message": msg,
    "type": t,
    "time": now
  });
};


// findElement returns the keypath of a predicate, or false if no match.
var findElement = function(pred) {
  var kp = false;
  ractive.data.views.forEach(function(v, i) {
    v.elements.forEach(function(e, j) {
      if (e.predicate === pred) {
        kp = "views."+i+".elements."+j;
      }
      // find path for multiPredicate inputs
      if (e.predicates) {
        e.predicates.forEach(function ( p ) {
          if ( p.uri === pred) {
            kp = "views."+i+".elements."+j;
          }
        });
      }
      // find path of inferred predicates
      if ( e.predicateInferred ) {
        e.predicateOptions.forEach(function ( p ) {
          if ( p === pred) {
            kp = "views."+i+".elements."+j;
          }
        });
      }
    });
  });
  return kp;
};

// findElementById returns the keypath of an id, or false if no match.
var findElementById = function( id ) {
  var kp = false;
  ractive.data.views.forEach(function(v, i) {
    v.elements.forEach(function(e, j) {
      if (e.id === id) {
        kp = "views."+i+".elements."+j;
      }
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

// deleteQuery generates the SPARQL query to remove a resource from the graph.
var deleteQuery = function( published ) {
  var graph = published ? ractive.data.publicGraph : ractive.data.draftsGraph;
  return 'DELETE { GRAPH ' + graph + ' { ' +
    ractive.get( 'existingURI' ) + ' ?p ?o } }\n' +
    'WHERE { ' + ractive.get( 'existingURI' ) + ' ?p ?o }';
};

// insertQuery generates the SPARQL query to insert the resource into the graph.
var insertQuery = function( publish ) {
  // generate overview triples using the internal namespace
  var uri = ractive.get( 'overview.uri' );
  var now = new Date();
  var meta = [
    { 'p': internalPred( 'profile' ), 'o': '"' + urlParams.profile + '"' },
    { 'p': internalPred( 'displayLabel' ), 'o': ractive.data.overview.displayLabel },
    { 'p': internalPred( 'searchLabel' ), 'o': ractive.data.overview.searchLabel },
    { 'p': internalPred( 'updated' ), 'o': dateFormat( now.toISOString() ) }
  ];
  for (var i=0; i<ractive.data.overview.type.length; i++) {
    meta.push( { 'p': 'a', 'o': ractive.get( 'overview.type.'+ i) });
  }
  if ( !ractive.data.uriFn && ractive.get( 'overview.idNumber' ) ) {
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
  for (var i=0; i<ractive.data.nonEditableProperties.length; i++) {
    meta.push( { 'p': ractive.get('nonEditableProperties.'+i+'.predicate'),
                 'o': ractive.get('nonEditableProperties.'+i+'.value')} );
  }

  var generated = "";
  if ( ractive.data.generatedValues ) {
    ractive.data.generatedValues.forEach( function( g ) {
      var gv = g(values);
      if ( gv ) {
        generated = generated + uri + ' ' + gv + " .\n";
      }
    });
  }
  metaPreds = _.reduce(meta, function(s, e) {
    return s + uri + " " + e.p + " " + e.o + " .\n";
  }, "");

  // generate a triple for each of the values in the form
  var preds = "";
  _.each(values, function(v, k) {
    _.each(v, function(e) {
      preds += uri + " " + e.predicate + " " + e.value + " . \n";
    });
  });
  var graph = publish ? ractive.data.publicGraph : ractive.data.draftsGraph;
  return 'INSERT { GRAPH ' + graph + ' {\n' + metaPreds + preds + generated + '} }';
};

// doQuery sends a SPARQL query the endpoint, and takes a success callback function.
// task parameter should be one of get/create/update/delete
var doQuery = function( query, task, callback ) {
  var postData = 'query=' + encodeURIComponent( query )
  postData += '&task=' +task;
  postData += '&uri=' + encodeURIComponent( ractive.get( 'overview.uri' ) );
  if ( ractive.get( 'overview.updated' ) ) {
    postData += '&updated=' + encodeURIComponent( ractive.get( 'overview.updated') );
  }
  var req = new XMLHttpRequest();
  req.open( 'POST', '/resource', true );
  req.setRequestHeader('Content-Type',
                       'application/x-www-form-urlencoded; charset=UTF-8');
  req.onload = function() {
    if ( req.status >= 200 && req.status < 400 ) {
      //console.log( req.responseText );
      callback( JSON.parse( req.responseText ) );
    } else {
      ractive.set( { 'serverError': req.responseText } );
    }
  };

  req.onerror = function() {
    console.log( 'Failed to execute SPAQRL query' );
  };

  req.send( postData );
};

// searchES queries ElasticSearch for URIs with searchLabel q
var searchES = _.debounce( function( q, kp) {
  // Trim whitespace from the query string.
  q = q.trim();
  // Return if empty string
  if (q === "") {
    ractive.set( kp + ".searching", false);
    return;
  }

  var searchTypes = ractive.get( kp + '.searchTypes' ).join(',');
  var searchQuery = { "query": { "filtered": { "query": {}, "filter": { "bool": {"must_not": [{"missing": {"field": "published"}}]}}} } };
  if ( q.length == 1 ) {
    // Do a prefix query if query string is only one character
    searchQuery.query.filtered.query.prefix = { "searchLabel": q };
  } else {
    // Otherwise normal match query (matches ngram size 2-20)
    searchQuery.query.filtered.query.match = { "searchLabel": { "query": q, "operator": "and" } };
  }
  // filter the current URI if we're editing a resource
  if ( ractive.get( 'existingResource' ) ) {
    searchQuery.query.filtered.filter.bool.must_not.push( {"ids": {"values": [trimURI( ractive.get( 'overview.uri' ) )]}} );
  }
  var queryData = JSON.stringify( searchQuery );
  var req = new XMLHttpRequest();
  req.open( 'POST', '/search/public/'+ searchTypes, true) ;
  req.setRequestHeader( 'Content-Type', 'application/json; charset=UTF-8' );

  req.onerror = function( e ) {
    console.log( "failed to reach search endoint: " + e.target.status );
  }

  req.onload = function( e) {
    //console.log( e.target.responseText );
    var qRes = JSON.parse( e.target.responseText );
    ractive.set( 'searchResults', qRes.hits.hits );
    ractive.set( 'searchSummary', qRes.hits.total + ' treff (' + qRes.took + ' ms)' );
  }

  req.send( queryData );

  ractive.merge( kp + ".searching", true);
  ractive.merge( kp + ".selectedResult", 0); // reset

}, 100); // debounce 100 ms


// event handlers ------------------------------------------------------------

listener = ractive.on({
  queryExternal: function( event ) {
    log("Querying external resources: " + ractive.get( 'externalSources' ).length, false);
    ractive.get( 'externalSources' ).forEach(function ( source ) {
      var q = source.genRequest( values );
      var req = new XMLHttpRequest();
      req.open( 'POST', '/external/'+ source.source, true );
      req.setRequestHeader( 'Content-Type', 'application/x-www-form-urlencoded; charset=UTF-8');

      req.onerror = function( event ) {
        // decrement pending counter
        ractive.subtract( 'externalQueriesPending' );
        if ( ractive.get( 'externalQueriesPending') == 0 ) {
          removeClass( document.body, 'wait' );
        }

        console.log( event );
        log( "Failed to send external query to server", true);
      }

      req.onload = function( event ) {
        // decrement pending counter
        ractive.subtract( 'externalQueriesPending' );
        if ( ractive.get( 'externalQueriesPending') == 0 ) {
          removeClass( document.body, 'wait' );
        }

        if ( req.status >= 200 && req.status < 400  ) {
          var parsedResponse = source.parseRequest( req.responseText );
          var v = parsedResponse[0]; // values
          var s = parsedResponse[1]; // suggestions
          log( source.source + ": OK. Values: " +v.length + ", suggestions: " + s.length, false);

          // 1. populate values
          v.forEach( function( val ) {
            val.source = source.source;
            var kp = findElement( val.predicate );
            if ( kp ) {
              val.predicateLabel = ractive.get(kp).label;
              if ( ractive.get(kp).type === "multiPredicateURI" ) {
                var predSelect = document.getElementById("multiPred-" + ractive.get(kp).id );
                for (var i =0; i < predSelect.options.length; i++) {
                  if ( pred === predSelect.options[i].value) {
                    val.predicateLabel = predSelect.options[i].innerHTML;
                  }
                }
              }
              // Check if value allready exists:
              var allreadyExists = false;
              for ( var i=0; i<ractive.get( kp + '.values' ).length; i++) {
                var exVal = ractive.get( kp + '.values.' + i );

                if ( exVal.value === val.value ) {
                  allreadyExists = true;
                  // If the value allready exists, simply add external source nome to source
                  if ( exVal.source.indexOf(val.source) == -1) {
                    ractive.set( kp + '.values.' + i + '.source', exVal.source + ', ' + val.source);
                  }
                }
              }
              if ( !allreadyExists ) {
                ractive.get(kp + ".values").push( val );
              }
            }
          });

          // 2.populate suggestions
          s.forEach( function( sug ) {
            sug.source = source.source;
            var kp = findElementById( sug.id );
            if ( kp ) {
              // Check if suggestion allready exists:
              var allreadyExists = false;
              for ( var i=0; i<ractive.get( kp + '.suggestions' ).length; i++) {
                var exVal = ractive.get( kp + '.suggestions.' + i );

                if ( exVal.value === sug.value ) {
                  allreadyExists = true;
                  // If the value allready exists, simply add external source nome to source
                  if ( exVal.source.indexOf(sug.source) == -1) {
                    ractive.set( kp + '.suggestions.' + i + '.source', exVal.source + ', ' + sug.source);
                  }
                }
              }
              if ( !allreadyExists ) {
                ractive.get(kp + ".suggestions").push( sug );
              }
            }
          });

        } else { // query status != 200 || 300
          log( source.source + ': Failed: "' + event.target.responseText + '"', true);
        }
      }

      req.send( "query=" + encodeURIComponent(q) );

      // inc pending counter
      ractive.add( 'externalQueriesPending' );

      // show waiting cursor
      addClass( document.body, 'wait' );

      log( source.source + ": Query sendt.", false);
    });
  },
  saveDraft: function( event ) {
    var published = ractive.get( 'overview.published' ) ? true : false;
    var q;
    if ( ractive.get( 'existingURI' ) ) {
      q = deleteQuery( published ) + ';\n' + insertQuery( false );
    } else {
      q = insertQuery( false );
    }
    var task = ractive.get( 'existingURI') ? 'updateDraft' : 'createDraft';
    doQuery( q, task, function() {
      // TODO check for return errors

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
    var task = published ? 'update' : 'create';
    doQuery( q, task, function() {
      // TODO check for returned errors

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
    var task = published ? 'delete' : 'deleteDraft';
    doQuery( q, task, function() {
      // TODO check for returned errors

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

    // Also delete dependant, if exists:
    if ( ractive.data.views[idx.i1].elements[idx.i2].dependant ) {
      var kp = findElementById( ractive.data.views[idx.i1].elements[idx.i2].dependant );
      ractive.get( kp+'.values').splice(0, 1);
    }
  },
  searchBlur: function ( event ) {
    // delay a bit so that the on-click event has time to fire in case of searchhit select
    setTimeout( function () {
      event.node.value = "";
      ractive.set( event.keypath + ".searching", false);
      ractive.set( 'searchResults', [] );
      ractive.set( 'searchSummary', '' );
    }, 200 );
  },
  newValue: function( event ) {
    var value, predicate, predicateLabel, source;

    value = event.node.value.trim();
    if ( event.context.predicateInferred ) {
      predicate = event.context.predicateFn( value );
    } else {
      predicate = event.context.predicate;
    }
    predicateLabel = event.context.label;
    source = 'local';

    // Don't add allow duplicate values:
    var exists = _.find( ractive.get( event.keypath ).values, function( e ) {
      return e.value === value;
    } );
    if (exists) {
      event.node.value = "";
      return
    }

    // Push to values array
    var idx = event.index;
    ractive.data.views[idx.i1].elements[idx.i2].values.push(
      {"predicate": predicate, "predicateLabel": predicateLabel, "value": value, "source": source});

    // Generate depdant values, if any
    if ( event.context.dependant ) {
      var depv = event.context.dependantTransform( value );
      if ( depv ) {
        var kp = findElementById( event.context.dependant );
        var dep = ractive.get( kp );

        ractive.set( kp + '.values.0', {
          "predicate": dep.predicate, "predicateLabel": dep.label,
          "value": depv, "source": source
        });
        ractive.update( kp + '.values' );
      }
    }

    // Reset, clear input field
    event.node.value = "";
    ractive.merge( event.keypath + ".errorInfo", "");
    ractive.set( event.keypath + '.currentValue', "");

    // If no more values allowed; attempt to focus on next input field.
    if ( !event.context.repeatable ) {
      var nextInput = event.node.parentElement.parentElement.parentElement.nextElementSibling.querySelector('input');
      if ( nextInput ) {
        nextInput.focus();
      }
    }
  },
  searchForURI: function( event ) {
    // Blur input field on Esc key.
    if (event.original.keyCode == 27) {
      ractive.fire( "searchBlur", event );
      return;
    }

    // Handle browsing and selecting search results by arrow keys.
    if ( ractive.get( event.keypath + '.searching' ) ) {
      var cur = ractive.get( event.keypath + ".selectedResult");

      // Browse search results down
      if (event.original.keyCode === 40) {
        if ( cur < ractive.get( 'searchResults').length-1 ) {
          ractive.add( event.keypath + ".selectedResult");
        }
        return;
      }

      // Browse search resuls up
      if (event.original.keyCode === 38) {
        if ( cur > 0 ) {
          ractive.subtract( event.keypath + ".selectedResult");
        }
        return;
      }

      // Select search result on Enter
      if (event.original.keyCode === 13) {
        var label, uri, predicate, predicateLabel, source;
        var selected = ractive.get( 'searchResults.' + cur);
        label = selected._source.displayLabel;
        uri = '<' + selected._source.uri + '>';
        source = 'local';
        var idx = event.index;
        predicate = ractive.data.views[idx.i1].elements[idx.i2].predicate;
        predicateLabel = ractive.data.views[idx.i1].elements[idx.i2].label;
        var predSelect = document.getElementById("multiPred-" + ractive.data.views[idx.i1].elements[idx.i2].id);
        if ( predSelect ) {
          predicateLabel = predSelect.options[predSelect.selectedIndex].innerHTML;
        }
        var exsitingURI = _.find(ractive.data.views[idx.i1].elements[idx.i2].values, function( e ) {
          return e.value === uri && e.predicate === predicate;
        });

        if ( !exsitingURI ) {
          ractive.data.views[idx.i1].elements[idx.i2].values.push(
            {"predicate": predicate, "predicateLabel": predicateLabel, "value": uri,
             "URILabel": label, "source": source});
        }
        ractive.set( event.keypath + '.searching', false);
        event.node.value = "";
        return;
      }
    }

    searchES( event.node.value, event.keypath );

  },
  selectURI: function( event ) {
    var label, uri, predicate, predicateLabel, source;
    label = event.context._source.displayLabel;
    uri = '<' + event.context._source.uri + '>';
    source = 'local';
    var idx = event.index;
    predicate = ractive.data.views[idx.i1].elements[idx.i2].predicate;
    predicateLabel = ractive.data.views[idx.i1].elements[idx.i2].label;
    var predSelect = document.getElementById("multiPred-" + ractive.data.views[idx.i1].elements[idx.i2].id);
    if ( predSelect ) {
      predicateLabel = predSelect.options[predSelect.selectedIndex].innerHTML;
    }
    var exsitingURI = _.find(ractive.data.views[idx.i1].elements[idx.i2].values, function( e ) {
      return e.value === uri && e.predicate === predicate;
    });

    if ( !exsitingURI ) {
      ractive.data.views[idx.i1].elements[idx.i2].values.push(
        {"predicate": predicate, "predicateLabel": predicateLabel, "value": uri,
         "URILabel": label, "source": source});
    }
    ractive.update( 'views.'+ idx.i1 + '.elements.' + idx.i2 + '.values' );

  },
  searchSuggestion: function( event ) {
    var input = event.node.parentElement.parentElement.querySelector('input');
    if ( input ) {
      var v = event.context.value;
      if ( v.match(/^(.*)\((.*\))$/) ) {
        // Strip pharentehes with identificator info, we don't want it
        // to be part of search query
        v = v.match(/^(.*)\((.*\))$/)[1];
      }
      setTimeout( function( ) {
        input.value = v.trim();
        input.focus();
        var idx = event.index;
        searchES( event.context.value, 'views.' + idx.i1 + '.elements.' + idx.i2 );
      }, 210);


    }
  },
  selectOption: function( event ) {
    var v = event.node.value;
    selected = _.find( ractive.get( event.keypath + '.options' ), function( o ) {
      return o.value === v;
    } );
    ractive.set( event.keypath +'.values.0', selected );
  },
  selectAddOption: function( event ) {
    var v = ractive.get( event.keypath+'.selected' );
    var selected = _.find( ractive.get( event.keypath + '.options' ), function( o ) {
      return o.value === v;
    } );

    // Don't add allow duplicate values:
    var exists = _.find( ractive.get( event.keypath).values, function( e ) {
      console.log(e);
      return e.value === v;
    } );
    if (!exists) {
      ractive.get( event.keypath ).values.push( selected );
    }
  },
  editLiteral: function( event ) {
    var kp = event.keypath.substr(0, event.keypath.indexOf('.values'));
    var v = event.context.value;
    var langTag = v.match(/(?:"@)(.*)/);

    // select language tag if we have one
    if ( langTag ) {
      ractive.set( kp + '.selectedLang', langTag[1] );
    } else {
      ractive.set( kp + '.selectedLang', "" );
    }

    // remove < > from URLs
    if ( ractive.get( kp ).type === "URL" ) {
      v = v.match(/^<(.*)>$/)[1];
    }

    // put value in input field
    ractive.set( kp + ".currentValue", cleanString( v ).replace(/\<br\/\>/g, "\n") );

    // set focus on input field
    setTimeout(function() {
      document.getElementById( ractive.get( kp ).id ).focus();
    }, 0);

    // remove the value we're editing from values array
    var idx = event.index;
    ractive.data.views[idx.i1].elements[idx.i2].values.splice(idx.i3, 1);
  },
  validateFloat: function( event ) {
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
  validateISBN: function( event ) {
    if ( event.node.value.trim() === "" ) {
      return
    }

    // validate ISBN
    if ( !isValidISBN( event.node.value ) ) {
      ractive.merge( event.keypath + ".errorInfo",
                    "ugyldig ISBN-nummer" );
      setTimeout( function () {
        event.node.focus();
      }, 0 );
      return;
    }

    event.node.value = '"' + event.node.value + '"';
    ractive.fire( "newValue", event );
  },
  validateString: function( event ) {
    if ( event.node.value.trim() === "" ) {
      // no action on empty input
      return
    }
    event.node.value = '"' + event.node.value + '"';
    ractive.fire( 'newValue', event );
  },
  validateLangString: function( event ) {
    var value, lang;
    var idx = event.index;
    value = event.node.value.trim();
    if ( value === "" ) {
      // no action on empty input
      return
    }
    lang = ractive.data.views[idx.i1].elements[idx.i2].selectedLang;

    // associate language tag if it is chosen
    if ( lang === "") {
      event.node.value = "\"" + event.node.value + "\"";
    } else {
      event.node.value = "\"" + event.node.value + "\"@" + lang;
    }

    ractive.fire("newValue", event);
  },
  validateURL: function( event ) {
    // no action on empty input
    var url = event.node.value.trim();
    if ( url === "" ) { return }

    // validate URL
    if ( !isURL( url ) ) {
      ractive.merge( event.keypath + ".errorInfo",
                    "ugyldig URL" );
      setTimeout( function () {
        event.node.focus();
      }, 0 );
      return;
    }

    event.node.value = '<' + url + '>';
    ractive.fire( 'newValue', event );
  },
  addText: function( event ) {
    var value, lang, pred;
    var idx = event.index;
    value = ractive.data.views[idx.i1].elements[idx.i2].currentValue;
    lang = ractive.data.views[idx.i1].elements[idx.i2].selectedLang;
    pred = ractive.data.views[idx.i1].elements[idx.i2].predicate;

    if ( value.trim() === "" ) {
      return;
    }

    // associate language tag if it is chosen
    if ( lang === "") {
      value = "\"" + value + "\"";
    } else {
      value = "\"" + value + "\"@" + lang;
    }
    ractive.data.views[idx.i1].elements[idx.i2].values.push(
      {"value": value.replace(/\n/g, "<br/>"), "predicate": pred, "source": "local"});

    ractive.data.views[idx.i1].elements[idx.i2].currentValue = "";
    ractive.update();
  },
  showHiddenField: function( event ) {
    event.context.hidden = false;
    ractive.update();
    // todo focus on input field
  }
});

// observers  ----------------------------------------------------------------

ractive.observe('views', function( newValue, oldValue, keypath) {
  // keep values in sync
  values = {};
  var missingValues = false;
  var missingForExternal = false;
  var tooManyValues = false;
  newValue.forEach(function(view, i) {
    view.elements.forEach(function(elem, j) {
      if (!values[elem.id]) {
        values[elem.id] = [];
      }
      // Check if all required attributes in the schema has a value
      if ( elem.required && elem.values.length === 0 ) {
        missingValues = true;
      }
      if ( !elem.repeatable && elem.values.length > 1) {
        tooManyValues = true;
      }
      if ( ractive.data.externalRequired && ractive.data.externalRequired.indexOf( elem.id ) >= 0 && elem.values.length === 0 ) {
        missingForExternal = true;
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

  // Don't allow to save when some fields have to many values.
  if ( tooManyValues ) {
    ractive.set( { 'publishDisabled': true, 'draftDisabled': true } );
  }

  // Toggle button for querying external sources.
  if ( missingForExternal || !ractive.data.externalRequired ) {
    ractive.set( 'externalDisabled', true);
  } else {
    ractive.set( 'externalDisabled', false);
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

  firstLoad = false; // TODO fix this
  var sl = ractive.data.searchLabel(values);
  var dl = ractive.data.displayLabel(values);

   // create searchLabel and displayLabel
  ractive.set( 'overview.searchLabel', sl);
  ractive.set( 'overview.displayLabel', dl);
});

ractive.observe( 'overview.uri', function( newURI, oldURI, keyPath ) {
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
  if ( ractive.get( 'uriFn') && !missingValues ) {
    // Check if URI allready exists in local RDF repo.
    if ( newURI !== "" && newURI !== ractive.get( 'existingURI' ) ) {
      var q = 'ASK WHERE { ' + newURI + '?p ?o }';
      doQuery( q, 'get', function( data) {
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
var createSchema = function( loadRes ) {
  // set values to empty array, if not allready popuated with predefined values.
  profile.views.forEach(function(view, i) {
    view.elements.forEach(function(elem, j) {
      // suggestions from external sources (where target values are URIs)
      profile.views[i].elements[j].suggestions = [];
      if ( !loadRes ) {
        if ( !profile.views[i].elements[j].values ) {
          profile.views[i].elements[j].values = [];
        }
      } else {
        // we don't want to populate with default values if we're loading
        // an existing resource.
        profile.views[i].elements[j].values = [];
      }
      profile.views[i].elements[j].currentValue = "";
      // set languageTag to defaultTag on all input fields
      profile.views[i].elements[j].selectedLang = common.defaultLang;
    });
  });

  // merge data with defaults from _common.js:
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
  doQuery( q, 'get', function( rdfRes ) {
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
        createSchema( true );
        ractive.set( 'nonEditableProperties', [] );

        // Find bindings which are labels for URIs
        var uriLabels = {};
        rdfRes.results.bindings.forEach( function( b ) {
          if ( b.l ) {
            uriLabels['<' + b.o.value + '>'] = b.l.value;
          }
        });
        ractive.data.overview.type = [];
        rdfRes.results.bindings.forEach(function(b) {
          var pred = "<" + b.p.value + ">";
          var source = 'local';
          var kp = findElement(pred);
          if ( kp ) {
            var v = getValue(b.o);
            var predLabel = ractive.get(kp).label;
            if ( ractive.get(kp).type === "multiPredicateURI" ) {
              // Find the predicateLabel for a multiPredicate input type
              var predSelect = document.getElementById("multiPred-" + ractive.get(kp).id );
              for (var i =0; i < predSelect.options.length; i++) {
                if ( pred === predSelect.options[i].value) {
                  predLabel = predSelect.options[i].innerHTML;
                }
              }
            }
            // If predicate is inferred from value:
            if ( ractive.get( kp ).predicateInferred ) {
              pred = ractive.get( kp ).predicateFn( v );
            }
            var res = {"predicate": pred, "predicateLabel": predLabel, "value": v, "source": source}
            if ( uriLabels[v] ) {
              res.URILabel = uriLabels[v]
            }
            if ( !b.l ) { // skip URI label bindings
              ractive.get(kp + ".values").push(res);
            }
          } else {
            var unknownPred = true;
            switch ( pred ) {
              case internalPred( 'displayLabel' ):
                ractive.set( 'overview.displayLabel', getValue( b.o ) );
                unknownPred = false;
                break;
              case internalPred( 'searchLabel' ):
                ractive.set( 'overview.searchLabel', getValue( b.o ) );
                unknownPred = false;
                break;
              case internalPred( 'created' ):
                ractive.set( 'overview.created', getValue( b.o ) );
                unknownPred = false;
                break;
              case internalPred( 'updated' ):
                ractive.set( 'overview.updated', getValue( b.o ) );
                unknownPred = false;
                break;
              case internalPred( 'published' ):
                ractive.set( 'overview.published', getValue( b.o ) );
                unknownPred = false;
                break;
              case internalPred( 'id' ):
                ractive.set( 'overview.idNumber', getValue( b.o ) );
                unknownPred = false;
                break;
              case "<http://www.w3.org/1999/02/22-rdf-syntax-ns#type>":
                ractive.data.overview.type.push( '<' + b.o + '>');
                unknownPred = false;
                break;
              case internalPred( 'profile' ):
                unknownPred = false;
                break;
            }
            ractive.set( 'overview.uri', '<' + urlParams.uri + '>' );

            // Store predicates and values whitch are NOT in profile schema:
            if (unknownPred) {
              ractive.data.nonEditableProperties.push({
              "predicate": pred, "value": getValue( b.o )
              });
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
  ractive.set('nonEditableProperties', []);
  loadScript( '/public/profiles/' + urlParams.profile + ".js",  createSchema);
  ractive.set( { 'draftDisabled': true, 'deleteDisabled': true, 'publishDisabled': true } );
}



