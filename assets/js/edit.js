let id = "";
let Image = "/image/"; // end with slash `/`, since it's wildcard
let Edit = "/api/edit";
let Manage = "/manage";

let hostname = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');

let editor = new EasyMDE({
    element: $('#editor')[0],
    toolbar: ["side-by-side", "preview", "fullscreen"],
    indentWithTabs: false,
    spellChecker: false,
    renderingConfig: {
        hljs: hljs,
        codeSyntaxHighlighting: true,
    }
});

function showOk(msg, cb, arg) {
    $('#myStatus').text('Status');
    $('#myModal').html(`<span style="color: green">${msg}</span>`);
    $('#modal').modal('show');
    setTimeout(function() {
        $('#modal').modal('hide');
    }, 2000);

    if(cb !== undefined) {
        cb(arg);
    }
}

function showError(msg) {
    $('#myStatus').text('Error');
    $('#myModal').html(`<span style="color: red">${msg}</span>`);
    $('#modal').modal('show');
}

function redirect(dst) {
    if(dst.startsWith('/')) {
        dst = dst.substring(1);
    }
    window.location.href = `${hostname}/${dst}`;
}

function deleteArticle() {
    if(id === "") {
        return redirect('manage');
    }

    fetch(`${Edit}/${id}`, {method: 'DELETE'}).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        redirect(Manage);
    }).catch(showError);
}

function publishArticle() {
    let images = {};

    let rows = document.getElementById('table').rows;
    for(let i = 0; i < rows.length; ++i) {
        let cols = rows[i].cells;
        if(cols.length !== 3) {
            showError('invalid table');
            return;
        }
        images[cols[0].innerText] = cols[1].innerText;
    }

    let title = $('#title').val();
    if(title.length === 0) {
        showError("title can't be empty");
        return;
    }

    let data = {
        title: title,
        content: editor.value(),
        tags: $('#tags').val(),
        images: JSON.stringify(images)
    };

    let method = id === "" ? 'POST' : 'PUT';
    let rid = id === "" ? '0' : id;
    fetch(`${Edit}/${rid}`, {
        method: method,
        body: JSON.stringify(data),
        headers: {
            'Content-Type': 'application/json'
        }
    }).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }

        if(id === "") {
            showOk('post saved, redirect to manage page', redirect, Manage);
        }
        else {
            showOk('post saved');
        }
    }).catch(showError);
}

function back2Manage() {
    redirect(Manage);
}

function uploadImage() {
    let img = $('#image-file').prop('files')[0];
    let formData = new FormData();

    formData.append('image', img);
    fetch(`${Image}?filename=${img.name}`, {method: 'POST', body: formData}).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }

        showOk('Upload Ok');

        $('#img').append(`<tr><td><a href="${res.link}">${res.link}</a></td><td>${res.name}</td><td><span class="btnDel glyphicon glyphicon-trash"></span></td></tr>`);
        editor.value(editor.value() + `![${res.name}](${res.link})`);
    }).catch(showError);
}

function toggleImage() {
    $('#imageDialog').modal('show');
}

function toggleTag() {
    $('#tagDialog').modal('show');
}

$(window).bind('keydown', function(event) {
    if (event.ctrlKey || event.metaKey) {
        switch (String.fromCharCode(event.which).toLowerCase()) {
            case 's':
                event.preventDefault();
                publishArticle();
                break;
            default:
                break;
        }
    }
});

$(document).ready(function() {
    $('#status_no').hide();
    $('#status_ok').hide();

    let s = location.pathname.split('/');
    let q = s[s.length - 1];
    if(q.length === 0) {
        return;
    }
    let nq = Number(q);
    if(Number.isNaN(nq) || nq === 0) {
        return;
    }

    id = '' + nq;

    fetch(`${Edit}/${id}`).then(function(j) {
        return j.json();
    }).then(function(data) {
        $('#title').val(data.title);
        $('#tags').val(data.tags);
        editor.value(data.content);

        let images = [];
        let jImages = JSON.parse(data.images);
        for(const [name, link] of Object.entries(jImages)) {
            images.push(`<tr><td><a href="${link}">${link}</a></td><td>${name}</td><td><span class="btnDel glyphicon glyphicon-trash"></span></td></tr>`);
        }
        $('#img').html(images.join(''));

        // remove item when click trash icon
        $("#table").on('click', '.btnDel', function () {
            $(this).closest('tr').remove();
        });
    }).catch(showError);
});