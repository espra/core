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
        var line = comment_info[0];
        var $container = get_container_for_line(line);
        var user = comment_info[1];
        var gravatar = comment_info[2];
        var timestamp = comment_info[3];
        var comment = comment_info[4];
        $container.append(
            '<div class="note-wrap"><div class="note-head">'
            + '<img class="gravatar-note" src="http://www.gravatar.com/avatar/'
            + gravatar
            + '?s=20&d=http%3A%2F%2Fgithub.com%2Fimages%2Fgravatars%2Fgravatar-20.png"'
            + '/> <a class="bold-link" href="/profile/' + user + '">' + user
            + '</a> wrote a comment <span class="small-grey">'
            + get_relative_time(new Date(parseInt(timestamp) * 1000))
            + '</span></div><div class="note-text">' + comment
            + '</div></div>'
        );
    }
};

var add_line_note = function (line) {
    var $container = get_container_for_line(line);
    var base_id = '' + ((new Date()).getTime()) + '-' + line;
    var $new = $(
        '<div class="note-wrap"><div class="note-head-write">'
        + '<a href="" class="note-button pressed" id="write-' + base_id
        + '">Write</a> '
        + '<a href="" class="note-button" id="preview-' + base_id
        + '">Preview</a>'
        + '</div><form class="note-text-write" id="form-' + base_id + '">'
        + '<textarea name="text" rows="6" id="textarea-' + base_id
        + '"></textarea><br />'
        + '</form><div class="note-text" id="text-' + base_id + '">'
        + '</div></div>'
        + '<a href="" class="button" id="submit-' + base_id
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
        $new.hide(); return false;
    });
    $('#textarea-' + base_id).focus();
};

$(function () {
    $('.timestamp').each(function () {
        var $span = $(this);
        $span.text(get_relative_time(new Date(parseInt($span.text()) * 1000)));
    });
});