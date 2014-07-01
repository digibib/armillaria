var common = {
  "languages": [
    {"label": "ikke angitt språk", "value": ""},
    {"label": "norsk", "value": "no"},
    {"label": "norsk (bokmål)", "value": "nb-no"},
    {"label": "norsk (nynorsk)", "value": "nn-no"},
    {"label": "engelsk", "value": "en"}
  ],
  "externalQueriesPending": 0,
  "logLines": [],
  "defaultLang": "nb-no",
  "internalNamespace": "armillaria://internal/",
  "publicGraph": "<http://data.deichman.no/books>",
  "draftsGraph": "<http://data.deichman.no/drafts>",
  "trimURI": function(s) { return s.substr(1, s.length -2)},
  "hiddenFields": function( view ) { return _.filter(view.elements, function( e ) { return e.hidden == true && e.values.length == 0})},
  "eng2norsk": function( s ) {
    var translations = {
      "classification": "klassifikasjon",
      "event": "hendelse",
      "location": "sted",
      "language": "språk",
      "manifestation": "manifestasjon",
      "person": "person",
      "script": "skriftsystem",
      "subject": "emne",
      "work": "verk"
    }
    if ( translations[s] ) {
      return translations[s];
    }
    return s;
  }
};

var isURL = function( s ) {
  return /(http|ftp|https):\/\/[\w-]+(\.[\w-]+)+([\w.,@?^=%&amp;:\/~+#-]*[\w@?^=%&amp;\/~+#-])?/.test(s);
};

var isValidISBN = function( value ) {
  // lifted from dojo toolkit (BSD)

  var len, sum = 0, weight;
  value = value.replace(/[- ]/g,''); //ignore dashes and whitespaces
  len = value.length;

  switch(len){
    case 10:
      weight = len;
      // ISBN-10 validation algorithm
      for(var i = 0; i < 9; i++){
        sum += parseInt(value.charAt(i)) * weight;
        weight--;
      }
      var t = value.charAt(9).toUpperCase();
      sum += t == 'X' ? 10 : parseInt(t);
      return sum % 11 == 0; // Boolean
      break;
    case 13:
      weight = -1;
      for(var i = 0; i< len; i++){
        sum += parseInt(value.charAt(i)) * (2 + weight);
        weight *= -1;
      }
      return sum % 10 == 0; // Boolean
      break;
  }
  return false;
};

var convertISBN10To13 = function( isbn10 ) {
  var chars = isbn10.split("");
  chars.unshift( "9", "7", "8");
  chars.pop();

  var sum = 0;
  for ( i=0; i < 12; i++) {
    sum += chars[i] * ( (i % 2) ? 3 : 1);
  }
  chars.push( (10 - ( sum % 10 ) ) % 10 );
  var isbn13 = chars.join("");
  return isbn13;
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

String.prototype.supplant = function (o) {
    return this.replace(/{([^{}]*)}/g,
        function (a, b) {
            var r = o[b];
            return typeof r === 'string' || typeof r === 'number' ? r : a;
        }
    );
};
