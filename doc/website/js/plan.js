// No Copyright (-) 2009-2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

var remove_item = function (array, item) {
  var i = 0;
  while (i < array.length) {
    if (array[i] == item) {
      array.splice(i, 1);
    } else {
      i++;
    }
  }
};

var dump_dict = function (ob) {
  for (var key in ob) {
	alert("key: " + key);
	alert("val: " + ob[key]);
  }
};

$(function () {

  var item;
  var i;
  var segment;

  var TAG2NAME = {};
  var NAME2TAG = {};
  var TAG2NORM = {};
  var NORM2TAG = {};
  var TAG2DISPLAYNAMES = {};
  var ITEM2DEPS = {};
  var IDS2TAGS = {};

  var segments = $('.tag-segment').get();

  for (i=0; i < segments.length; i++) {
	segment = segments[i];
	item = '#' + segment.id;

	$(item).children().each(function () {
      var tag = this.className;
	  if (tag == 'tag-link')
		return true;
	  if (tag.indexOf('tag-type-dep') != -1) {
		var dep = tag.split(' ')[2].slice(12);
		ITEM2DEPS[item] = dep; // @/@ implement dependency analysis
		return true;
	  }
	  tag = tag.split(' ');
	  remove_item(tag, 'tag');
	  tag = tag.join('-');
	  var id = item + '-main';
      if (id in IDS2TAGS) {
		IDS2TAGS[id].push(tag);
	  } else {
		IDS2TAGS[id] = [tag];
	  }
    if (!(tag in TAG2NAME))
	  TAG2NAME[tag] = this.getAttribute('tagname');
      NAME2TAG[this.getAttribute('tagname')] = tag;
    if (!(tag in TAG2NORM))
	  TAG2NORM[tag] = this.getAttribute('tagnorm');
      NORM2TAG[this.getAttribute('tagnorm')] = tag;
    if (!(tag in TAG2DISPLAYNAMES))
	  if (!this.innerText) {
	    TAG2DISPLAYNAMES[tag] = this.innerHTML; // this.textContent;
	  } else {
	    TAG2DISPLAYNAMES[tag] = this.innerText;
	  }
	return true;
	});

  }

  var IDS2TAGSETS = {};
  var ALL_IDS = [];
  var PARENT_IDS = {};
  var IDS2PARENTS = {};

  $.each(IDS2TAGS, function (k, v) {
	var dict = {};
	for (var m=0; m<v.length; m++)
	  dict[v[m]] = 1;
    IDS2TAGSETS[k] = dict;
    ALL_IDS.push(k);
    var parent_id = '#' + $(k).parent().attr('id');
	if (parent_id) {
	  if (!(parent_id in PARENT_IDS))
		PARENT_IDS[parent_id] = 1;
	  IDS2PARENTS[k] = parent_id;
	}
  });

  $('.section').each(function () {
    var parent_id = '#' + this.id;
    if (!(parent_id in PARENT_IDS))
      PARENT_IDS[parent_id] = 1;
  });

  var tag_button_handler = function () {

        var hash;
        var par;
		var tag = this.id;
		var self = $(this);

		for (par in PARENT_IDS)
		  $(par).show();

		if (tag == 'tag-all') {
		  $('.tag-content').show();
		  $('#plan-tags a').removeClass("buttondown");
		  self.addClass("buttondown");
		  CURRENT_TAGS.length = 0;
          if (window.location.hash) {
			hash = "#";
		  } else {
			hash = null;
		  }
		} else {
		  $('#tag-all').removeClass('buttondown');
		  self.toggleClass("buttondown");
		  if (self.hasClass('buttondown')) {
			CURRENT_TAGS.push(tag);
		  } else {
			remove_item(CURRENT_TAGS, tag);
		  }
		  if (CURRENT_TAGS.length == 0) {
			$('.tag-content').show();
			$('#plan-tags a').removeClass("buttondown");
			$('#tag-all').addClass("buttondown");
			hash = '#';
		  } else {
            SHOWN_PARENTS = {};
			$('.tag-content').hide();
			for (var j=0; j<ALL_IDS.length; j++) {
              var citem_id = ALL_IDS[j];
			  var citem_tags = IDS2TAGSETS[citem_id];
			  var show = true;
			  for (var t=0;	t<CURRENT_TAGS.length; t++) {
				var ct = CURRENT_TAGS[t];
				if (!(ct in citem_tags)) {
				  show = false;
				  break;
				}
			  }
			  if (show == true) {
				$(citem_id).show();
				SHOWN_PARENTS[IDS2PARENTS[citem_id]] = true;
			  }
			}
			for (par in PARENT_IDS) {
			  if (!(par in SHOWN_PARENTS)) {
				$(par).hide();
			  }
			}
            hash = "#";
			for (var m=0; m < CURRENT_TAGS.length; m++) {
			  if (m != 0)
				hash += ',';
			  hash += TAG2NORM[CURRENT_TAGS[m]];
			  }
			}
		}

		if (hash)
		  window.location.hash = hash;

		this.blur();
		return false;

  };

  var plan_container = $('#plan-container');
  var plan = $('<div id="plan-tags"></div>');
  var all = $('<a class="button" href="" id="tag-all"><span>All</span></a>');
  // var deps = $('<a class="button" href="" id="tag-deps"><span>Show Dependencies</span></a>');
  var help = $('<div class="plan-help">↙ Use these buttons to filter for items with ALL the selected tags ↙</div><hr class="clear" />');

  all.click(tag_button_handler);
  plan.append(help);
  plan.append(all);
  // plan.append(deps);

  var tagnames = [];

  for (var key in TAG2NAME)
    tagnames.push([key, TAG2NAME[key]]);

//   for (var key in TAG2DISPLAYNAMES)
// 	tagnames.push([key, TAG2DISPLAYNAMES[key]]);

  tagnames.sort();

  CURRENT_TAGS = [];
  var last_prefix = '';

  for (i=0; i<tagnames.length; i++) {
	var tagname =	tagnames[i];
	var tagid = tagname[0];
	var button = $(
	  '<a class="button" href="" id="'+tagid+'"><span>'+tagname[1]+'</span></a>'
	  ).click(tag_button_handler);
    var new_prefix = tagid.slice(0, tagid.indexOf('-tag-val-'));
	if (new_prefix != last_prefix) {
	  if (last_prefix)
        plan.append($('<hr class="clear" />'));
	  last_prefix = new_prefix;
	}
	plan.append(button);
  }

  plan_container.append(plan);

  VALID_CHARS = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_0123456789';
  VALID_CHAR_DICT = {};

  for (var j=0; j < VALID_CHARS.length; j++) {
    VALID_CHAR_DICT[VALID_CHARS[j]] = true;
  }

  if (window.location.hash) {
    var requested_tags = window.location.hash.substr(1).split(',');
    for (var x=0; x < requested_tags.length; x++) {
      var tag = requested_tags[x];
	  $('#' + NORM2TAG[tag.toLowerCase()]).click();
    }
  } else {
    $('#tag-all').click();
  }

});

