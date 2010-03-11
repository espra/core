// No Copyright (-) 2009-2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

/*jslint plusplus: false */
/*global $, alert, window */

(function () {

    function remove_item(array, item) {
        var i = 0;
        while (i < array.length) {
            if (array[i] === item) {
                array.splice(i, 1);
            } else {
                i += 1;
            }
        }
    }

    $(function () {

        var item,
            item_id,
            id,
            i,
            j,
            x,
            segment,
            tag_fragment_updated = false,
            tag_fragment_found = false,
            TAG2NAME = {},
            NAME2TAG = {},
            TAG2NORM = {},
            NORM2TAG = {},
            TAG2DISPLAYNAMES = {},
            ITEM2DEPS = {},
            IDS2TAGS = {},
            IDS2TAGSETS = {},
            ALL_IDS = [],
            PARENT_IDS = {},
            IDS2PARENTS = {},
            segments = $('.tag-segment').get(),
            plan_container = $('#plan-container'),
            plan = $('<div id="plan-tags"></div>'),
            all = $('<a class="button" href="" id="tag-all"><span>All</span></a>'),
            help = $('<div class="plan-help">↙ Use these buttons to filter for items with ALL the selected tags ↙</div><hr class="clear" />'),
            CURRENT_TAGS = [],
            VALID_CHARS = 'abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_0123456789',
            VALID_CHAR_DICT = {},
            last_prefix = '',
            key,
            requested_tags,
            tag,
            tagnames = [],
            tagname,
            tagid,
            button,
            new_prefix;

        // deps = $('<a class="button" href="" id="tag-deps"><span>Show Dependencies</span></a>'),

        // add <p /> wrapping to certain elements
        $('div.container > blockquote').each(function () {
            var p_elems = $('p', this);
            if (p_elems.length === 0) {
                $(this).wrapInner('<p />');
            }
        });

        function extract_metadata() {
            var tag = this.className,
                dep;
            if (tag === 'tag-link') {
                return true;
            }
            if (tag.indexOf('tag-type-dep') !== -1) {
                dep = tag.split(' ')[2].slice(12);
                ITEM2DEPS[id] = dep; // @/@ implement dependency analysis
                return true;
            }
            tag = tag.split(' ');
            remove_item(tag, 'tag');
            tag = tag.join('-');
            if (id in IDS2TAGS) {
                IDS2TAGS[id].push(tag);
            } else {
                IDS2TAGS[id] = [tag];
            }
            if (!(tag in TAG2NAME)) {
                TAG2NAME[tag] = this.getAttribute('tagname');
                NAME2TAG[this.getAttribute('tagname')] = tag;
            }
            if (!(tag in TAG2NORM)) {
                TAG2NORM[tag] = this.getAttribute('tagnorm');
                NORM2TAG[this.getAttribute('tagnorm')] = tag;
            }
            if (!(tag in TAG2DISPLAYNAMES)) {
                if (!this.innerText) {
                    TAG2DISPLAYNAMES[tag] = this.innerHTML; // this.textContent;
                } else {
                    TAG2DISPLAYNAMES[tag] = this.innerText;
                }
            }
            return true;
        }
        
        for (i = 0; i < segments.length; i++) {
            segment = segments[i];
            item_id = '#' + segment.id;
            id = item_id.slice(0, item_id.lastIndexOf('-tag'));
            $(item_id).children().each(extract_metadata);
        }

        $.each(IDS2TAGS, function (k, v) {
            var dict = {},
                m,
                parent_id;
            for (m = 0; m < v.length; m++) {
                dict[v[m]] = 1;
            }
            IDS2TAGSETS[k] = dict;
            ALL_IDS.push(k);
            parent_id = '#' + $(k).parent().attr('id');
            if (parent_id) {
                if (!(parent_id in PARENT_IDS)) {
                    PARENT_IDS[parent_id] = 1;
                }
                IDS2PARENTS[k] = parent_id;
            }
        });

        $('.section').each(function () {
            var parent_id = '#' + this.id;
            if (!(parent_id in PARENT_IDS)) {
                PARENT_IDS[parent_id] = 1;
            }
        });

        function tag_button_handler() {

            var hash,
                par,
                tag = this.id,
                self = $(this),
                SHOWN_PARENTS,
                j,
                m,
                citem_id,
                citem_tags,
                show,
                t,
                ct;

            for (par in PARENT_IDS) {
                if (PARENT_IDS.hasOwnProperty(par)) {
                    $(par).show();
                }
            }

            if (tag === 'tag-all') {
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
                if (CURRENT_TAGS.length === 0) {
                    $('.tag-content').show();
                    $('#plan-tags a').removeClass("buttondown");
                    $('#tag-all').addClass("buttondown");
                    hash = '#';
                } else {
                    SHOWN_PARENTS = {};
                    $('.tag-content').hide();
                    for (j = 0; j < ALL_IDS.length; j++) {
                        citem_id = ALL_IDS[j];
                        citem_tags = IDS2TAGSETS[citem_id];
                        show = true;
                        for (t = 0; t < CURRENT_TAGS.length; t++) {
                            ct = CURRENT_TAGS[t];
                            if (!(ct in citem_tags)) {
                                show = false;
                                break;
                            }
                        }
                        if (show === true) {
                            $(citem_id).show();
                            SHOWN_PARENTS[IDS2PARENTS[citem_id]] = true;
                        }
                    }
                    for (par in PARENT_IDS) {
                        if (PARENT_IDS.hasOwnProperty(par)) {
                            if (!(par in SHOWN_PARENTS)) {
                                $(par).hide();
                            }
                        }
                    }
                    hash = "#";
                    for (m = 0; m < CURRENT_TAGS.length; m++) {
                        if (m !== 0) {
                            hash += ',';
                        }
                        hash += TAG2NORM[CURRENT_TAGS[m]];
                    }
                }
            }

            if (hash) {
                if (hash === "#") {
                    if (tag_fragment_updated) {
                        window.location.hash = hash;
                    } else {
                        tag_fragment_updated = true;
                    }
                } else {
                    window.location.hash = hash;
                }
            }

            this.blur();
            return false;

        }

        all.click(tag_button_handler);
        plan.append(help);
        plan.append(all);
        // plan.append(deps);

        for (key in TAG2NAME) {
            if (TAG2NAME.hasOwnProperty(key)) {
                tagnames.push([key, TAG2NAME[key]]);
            }
        }

        // for (var key in TAG2DISPLAYNAMES)
        //     tagnames.push([key, TAG2DISPLAYNAMES[key]]);

        tagnames.sort();

        for (i = 0; i < tagnames.length; i++) {
            tagname = tagnames[i];
            tagid = tagname[0];
            button = $(
                '<a class="button" href="" id="' + tagid + '"><span>' + tagname[1] + '</span></a>'
            ).click(tag_button_handler);
            new_prefix = tagid.slice(0, tagid.indexOf('-tag-val-'));
            if (new_prefix !== last_prefix) {
                if (last_prefix) {
                    plan.append($('<hr class="clear" />'));
                }
                last_prefix = new_prefix;
            }
            plan.append(button);
        }

        plan_container.append(plan);

        for (j = 0; j < VALID_CHARS.length; j++) {
            VALID_CHAR_DICT[VALID_CHARS[j]] = true;
        }

        if (window.location.hash) {
            requested_tags = window.location.hash.substr(1).split(',');
            for (x = 0; x < requested_tags.length; x++) {
                tag = decodeURIComponent(requested_tags[x]).toLowerCase();
                if (NORM2TAG[tag]) {
                    tag_fragment_found = true;
                    $('#' + NORM2TAG[tag]).click();
                }
            }
        }

        if (!tag_fragment_found) {
            $('#tag-all').click();
        }

    });

}());