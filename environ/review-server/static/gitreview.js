// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

var NOW = new Date();

var get_relative_time = function (timestamp) {
    var delta = (NOW - timestamp) / 1000;
    if (delta < 60)
        return "less than " + Math.floor(delta) + " seconds ago";
    delta = Math.floor(delta / 60);
    if (delta === 0)
        return 'less than a minute ago';
    if (delta === 1)
        return 'a minute ago';
    if (delta < 60)
        return delta + ' minutes ago';
    if (delta < 62)
        return 'about 1 hour ago';
    if (delta < 120)
        return 'about 1 hour and ' + (delta - 60) + ' minutes ago';
    if (delta < 1440)
        return 'about ' + Math.floor(delta / 60) + ' hours ago';
    if (delta < 2880)
        return '1 day ago';
    if (delta < 43200)
        return Math.floor(delta / 1440) + ' days ago';
    if (delta < 86400)
        return 'about 1 month ago';
    if (delta < 525960)
        return Math.floor(delta / 43200) + ' months ago';
    if (delta < 1051199)
        return 'about 1 year ago';
    return 'over ' + Math.floor(delta / 525960) + ' years ago';
};

var get_container_for_line = function (line) {
    var $container = $('#' + line + '-notes');
    if (!$container.length) {
        var $tr = $('#' + line).parent();
        $container = $('<td class="note-list" id="' + line + '-notes" colspan="3"></td>');
        var $newTr = $('<tr />');
        $container.appendTo($newTr);
        $tr.after($newTr);
    }
    return $container;
};

var add_line_notes = function (data) {
    if (!data)
        return;
    var number_of_comments = data.length;
    for (var i=0; i < number_of_comments; i++) {
        var comment_info = data[i];
        var filename = comment_info[0];
        var line = comment_info[1];
        var line_id = 'L' + files.indexOf(filename) + line;
        var $container = get_container_for_line(line_id);
        var user = comment_info[2];
        var gravatar = comment_info[3];
        var timestamp = comment_info[4];
        var comment = comment_info[5];
        $container.append(
            '<div class="note-wrap"><div class="note-head">'
            + '<img class="gravatar-note" src="http://www.gravatar.com/avatar/'
            + gravatar
            + '?s=20&d=http%3A%2F%2Fgithub.com%2Fimages%2Fgravatars%2Fgravatar-20.png"'
            + '/> <a class="bold-link" href="/profile/' + encodeURIComponent(user) + '">'
            + user.replace('&', '&amp;').replace('<', '&lt;').replace('>', '&gt;').replace('"', '&quot;')
            + '</a> wrote a comment <span class="small-grey">'
            + get_relative_time(new Date(parseInt(timestamp) * 1000))
            + '</span></div><div class="note-text">' + comment
            + '</div></div>'
        );
    }
};

var add_line_note = function (line) {
    line = line.split('-');
    var file = line[0];
    line = line[1];
    var line_id = 'L' + file + line;
    var $container = get_container_for_line(line_id);
    var filename = files[parseInt(file, 10)];
    var base_id = '' + ((new Date()).getTime()) + '-' + line_id;
    var $new = $(
        '<div class="note-wrap"><div class="note-head-write">'
        + '<a href="" class="note-button pressed" id="write-' + base_id
        + '">Write</a> '
        + '<a href="" class="note-button" id="preview-' + base_id
        + '">Preview</a>'
        + '</div><form action="/comment" method="post" '
        + 'class="note-text-write" id="form-' + base_id + '">'
        + '<input type="hidden" name="key" value="' + key + '" />'
        + '<input type="hidden" name="xsrf_token" value="' + xsrf_token + '" />'
        + '<input type="hidden" name="file" value="' + filename + '" />'
        + '<input type="hidden" name="line" value="' + line + '" />'
        + '<textarea name="text" rows="6" id="textarea-' + base_id
        + '"></textarea><br />'
        + '</form><div class="note-text" id="text-' + base_id + '">'
        + '</div></div>'
        + '<a class="button" id="submit-' + base_id
        + '"><span>Add Comment</span></a>'
        + '<hr class="clear" />'
        );
    $container.append($new);
    $('#preview-' + base_id).click(function () {
        $('#preview-' + base_id).addClass('pressed');
        $('#write-' + base_id).removeClass('pressed');
        $('#form-' + base_id).hide();
        var $preview = $('#text-' + base_id);
        var text = $('#textarea-' + base_id).val();
        if (!text) {
            $preview.html("<p>Nothing to preview.</p>").show();
        } else {
            $.post('/preview', {text: text}, function(data) {
                $preview.html(data).show();
            });
        }
        return false;
    });
    $('#write-' + base_id).click(function () {
        $('#preview-' + base_id).removeClass('pressed');
        $('#write-' + base_id).addClass('pressed');
        $('#form-' + base_id).show();
        $('#text-' + base_id).hide();
        $('#textarea-' + base_id).focus();
        return false;
    });
    $('#submit-' + base_id).click(function () {
        $('#form-' + base_id).submit();
    });
    $('#textarea-' + base_id).focus();
};

$(function () {
    $('.timestamp').each(function () {
        var $span = $(this);
        $span.text(get_relative_time(new Date(parseInt($span.text()) * 1000)));
    });
});

function setup_comments () {
    $('#comment-preview').click(function () {
        $('#comment-preview').addClass('pressed');
        $('#comment-write').removeClass('pressed');
        $('#comment-form').hide();
        var $preview = $('#comment-text');
        var text = $('#comment-textarea').val();
        if (!text) {
            $preview.html("<p>Nothing to preview.</p>").show();
        } else {
            $.post('/preview', {text: text}, function(data) {
                $preview.html(data).show();
            });
        }
        return false;
    });
    $('#comment-write').click(function () {
        $('#comment-preview').removeClass('pressed');
        $('#comment-write').addClass('pressed');
        $('#comment-form').show();
        $('#comment-text').hide();
        $('#comment-textarea').focus();
        return false;
    });
    $('#comment-textarea').focus();
    $('#comment-submit').click(function () { $('#comment-form').submit(); });
    $('.add-bubble').each(function (elem) {
        var $elem = $(this);
        $elem.click(function () {
            add_line_note($elem.attr('rel'));
        });
    });
};