// Beginning of logilab strptime implementation
var _DATE_FORMAT_REGEXES = {
    'Y': new RegExp('^-?[0-9]+'),
    'd': new RegExp('^[0-9]{1,2}'),
    'm': new RegExp('^[0-9]{1,2}'),
    'H': new RegExp('^[0-9]{1,2}'),
    'M': new RegExp('^[0-9]{1,2}')
}

/*
 * _parseData does the actual parsing job needed by `strptime`
 */
var _parseDate = function (datestring, format) {
    var parsed = {};
    for (var i1=0,i2=0;i1<format.length;i1++,i2++) {
    var c1 = format[i1];
    var c2 = datestring[i2];
	if (c1 == '%') {
	   c1 = format[++i1];
	   var data = _DATE_FORMAT_REGEXES[c1].exec(datestring.substring(i2));

	   if (!data.length) {
		    return null;
	   }

	   data = data[0];
	   i2 += data.length-1;
	   var value = parseInt(data, 10);

	   if (isNaN(value)) {
	      return null;
	   }

	   parsed[c1] = value;
	   continue;
	}

	if (c1 != c2) {
	   return null;
	}
    }
    return parsed;
};

/*
 * basic implementation of strptime. The only recognized formats
 * defined in _DATE_FORMAT_REGEXES (i.e. %Y, %d, %m, %H, %M)
 */
var strptime = function (datestring, format) {
    var parsed = _parseDate(datestring, format);

    if (!parsed) {
       return null;
    }

    // create initial date (!!! year=0 means 1900 !!!)
    var date = new Date(0, 0, 1, 0, 0);
    date.setFullYear(0); // reset to year 0

    if (parsed.Y) {
       date.setFullYear(parsed.Y);
    }

    if (parsed.m) {

       if (parsed.m < 1 || parsed.m > 12) {
          return null;
       }

       // !!! month indexes start at 0 in javascript !!!
       date.setMonth(parsed.m - 1);
    }

    if (parsed.d) {

       if (parsed.m < 1 || parsed.m > 31) {
          return null;
       }

       date.setDate(parsed.d);
    }

    if (parsed.H) {

       if (parsed.H < 0 || parsed.H > 23) {
          return null;
       }

       date.setHours(parsed.H);
    }

    if (parsed.M) {

       if (parsed.M < 0 || parsed.M > 59) {
          return null;
       }

       date.setMinutes(parsed.M);
    }

    return date;
};
// End of logilab strptime implementation

var get_relative_time = function (from) {
    var date = new Date;
    date.setTime(Date.parse(from));
    var distance_in_seconds = ((NOW - date) / 1000);
    var distance_in_minutes = Math.floor(distance_in_seconds / 60);

    if (distance_in_minutes == 0) { return 'less than a minute ago'; }
    if (distance_in_minutes == 1) { return 'a minute ago'; }
    if (distance_in_minutes < 45) { return distance_in_minutes + ' minutes ago'; }
    if (distance_in_minutes < 90) { return 'about 1 hour ago'; }
    if (distance_in_minutes < 1440) { return 'about ' + Math.floor(distance_in_minutes / 60) + ' hours ago'; }
    if (distance_in_minutes < 2880) { return '1 day ago'; }
    if (distance_in_minutes < 43200) { return Math.floor(distance_in_minutes / 1440) + ' days ago'; }
    if (distance_in_minutes < 86400) { return 'about 1 month ago'; }
    if (distance_in_minutes < 525960) { return Math.floor(distance_in_minutes / 43200) + ' months ago'; }
    if (distance_in_minutes < 1051199) { return 'about 1 year ago'; }

    return 'over ' + Math.floor(distance_in_minutes / 525960) + ' years ago';
};

// http://friendfeed-api.com/v2/search?q=ampify+OR+espians+OR+group%3Aampify+OR+group%3Aespians# 
