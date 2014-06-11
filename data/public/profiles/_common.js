var common = {
  "languages": [
    {"label": "ikke angitt språk", "value": ""},
    {"label": "norsk", "value": "no"},
    {"label": "norsk (bokmål)", "value": "nb-no"},
    {"label": "norsk (nynorsk)", "value": "nn-no"},
    {"label": "engelsk", "value": "en"}
  ],
  "internalNamespace": "armillaria://internal/",
  "publicGraph": "<http://data.deichman.no/public>",
  "draftsGraph": "<http://data.deichman.no/drafts>"
};

var cleanString = function(s) {
  var m = s.match(/"(.)+"/);
  if ( m ) {
    return m[0].substr(1, m[0].length - 2);
  }
  return s;
};

var internalPred = function(s) {
  return "<" + common.internalNamespace + s + ">";
};

var dateFormat = function(d) {
  return '"' + d + '"^^<http://www.w3.org/2001/XMLSchema#dateTime>'
}

var trimURI = function(s) {
  return s.substr( 1, s.length - 2 );
}
