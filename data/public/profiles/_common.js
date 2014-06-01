var common = {
  "languages": [
    {"label": "ikke angitt språk", "value": ""},
    {"label": "norsk", "value": "no"},
    {"label": "norsk (bokmål)", "value": "nb-NO"},
    {"label": "norsk (nynorsk)", "value": "nn-NO"},
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
