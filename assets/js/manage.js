let ModelHome = "/";
let ModelEdit = "/edit";
let Logout = "/api/login";
let Article = "/api/edit";
let Manage = "/api/manage";
let Settings = "/manage/setting";
let Thread = "/article";

function redirect(dst) {
    window.location.href = dst;
}

function newPost() {
    redirect(`${ModelEdit}/0`);
}

function toHome() {
    redirect(ModelHome);
}

function toSettings() {
    redirect(Settings);
}

function logOut() {
    fetch(Logout, {method: 'DELETE'}).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        redirect(ModelHome);
    }).catch(function(e) {
        alert(e);
    });
}

function toggleEye(id) {
    let hide = $(`#eye-${id}`).hasClass("glyphicon-eye-close");
    fetch(`${Article}/${id}?hide=${!hide}`, {method: 'PUT'}).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        if(hide) {
            $(`#eye-${id}`).html("status: show");
            $(`#eye-${id}`).removeClass("glyphicon-eye-close").addClass("glyphicon-eye-open");
        }
        else {
            $(`#eye-${id}`).html("status: hide");
            $(`#eye-${id}`).removeClass("glyphicon-eye-open").addClass("glyphicon-eye-close");
        }
    }).catch(function(e) {
        alert(e);
    });
}

function deleteArticle(o, id) {
    fetch(`${Article}/${id}`, {method: 'DELETE'}).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        let p= o.parentNode.parentNode;
        p.parentNode.removeChild(p);
    }).catch(function(e) {
        alert(e);
    });
}

function editArticle(id) {
    redirect(`${ModelEdit}/${id}`);
}

function buildTable(data) {
    $('#posts').empty(); // remove all children
    for(const post of data.posts) {
        let pid = `<td style="width: 5%">${post.id}</td>`;
        let ptitle = `<td style="width: 20%"><a  target="_blank" href="${Thread}/${post.id}">${post.title}</a></td>`;
        let ptime = `<td style="width: 20%">${post.create_time}<br>${post.last_modified}</td>`;
        let arr = [];
        for(const tag of post.tags) {
            arr.push(`<span class="badge">${tag}</span>`);
        }
        let ptag = `<td style="width: 20%">${arr.join('\n')}</td>`;
        let on = "open";
        let status = "status: show";
        if(post.hide) {
            status = "status: hide";
            on = "close";
        }
        let pop =
            `<td>
                    <botton id="eye-${post.id}" class="btn btn-default glyphicon glyphicon-eye-${on}" onclick="toggleEye(${post.id});">${status}</botton>
                    <button class="btn btn-default glyphicon glyphicon-trash" onclick="deleteArticle(this, ${post.id});"></button>
                    <button type="button" class="btn btn-default glyphicon glyphicon-pencil" onclick="editArticle(${post.id});"></button>
                </td>`;

        $('#posts').append(`<tr>${pid}${ptitle}${ptime}${ptag}${pop}</tr>`);
    }
}

function init() {
    fetch(Manage).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        buildTable(res);
    }).catch(function(e) {
        alert(e);
    });
}

$(document).ready(function() {
    init();
});