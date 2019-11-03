let ModelHome = "/";
let ModelManage = "/manage";
let Navi = "/api/navi";
let User = "/api/user";
let Logout = "/api/login";
let hostname = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');

function showOk(msg, cb, arg) {
    $('#myStatus').text('Status');
    $('#myModal').html(`<span style="color: green">${msg}</span>`);
    $('#modal').modal('show');
    setTimeout(function () {
        $('#modal').modal('hide');
        cb(arg);
    }, 2000);
}

function showErr(msg) {
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

function toHome() {
    redirect(ModelHome);
}

function toManage() {
    redirect(ModelManage);
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


function saveProfile() {
    let l = $('#password').val();
    let r = $('#password').val();
    if(l === '' || r === '') {
        $('#password').val('');
        showErr("password can't be empty");
    }
    else if(l !== r) {
        $('#password').val('');
        showErr("password mismatch");
    }
    else {
        fetch(User, {
            method: 'PUT',
            body: l,
            headers: {
                'Content-Type': 'text/plain'
            }
        }).then(function(j) {
            return j.json();
        }).then(function(res) {
            if(res.code !== 0) {
                throw res.error;
            }

            showOk("password changed, logout in 2 seconds", logOut);
        }).catch(function(e) {
            showErr(e);
        });
    }
}

function saveNavi(id) {
    let method = id === '' || id === undefined ? 'POST' : 'PUT';
    if(id === undefined) {
        id = '';
    }
    let seq = $(`#seq-${id}`).val();
    let name = $(`#name-${id}`).val();
    let target = $(`#dst-${id}`).val();
    if(seq === '' || name === '' || target === '') {
        showErr('fields can NOT be empty');
        return;
    }

    id = Number(id);
    seq = Number(seq);
    if(Number.isNaN(id) || Number.isNaN(seq)) {
        showErr('sequence must be number');
        return;
    }

    let data = {
        id: id,
        sequence: seq,
        name: name,
        target: target
    };
    fetch(Navi, {
        method: method,
        body: JSON.stringify(data),
        headers: {
            'Content-Type': 'application/json'
        }
    }).then(function(j) {
        return j.json();
    }).then(function(res) {
        console.log(res);
        if(res.code !== 0) {
            throw res.error;
        }
        initNavi();
        showOk('navi saved');
    }).catch(function(e) {
        showErr(e);
    });
}

function deleteNavi(o, id) {
    fetch(`${Navi}?id=${id}`, {method: 'DELETE'}).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        var p=o.parentNode.parentNode;
        p.parentNode.removeChild(p);
    }).catch(function(e) {
        showErr(e);
    });
}

function newNavi() {
    buildTable({
            navis: [{
                id: '',
                sequence: '',
                name: '',
                target: '',
            }]
        }
    );
}

$(document).ready(function() {
    $('#modal').modal('hide');
    initNavi();
});

function initNavi() {
    fetch(Navi).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        $('#navis').empty();
        buildTable(res);
    }).catch(function(e) {
        showErr(e);
    });
}

function buildTable(data) {
    for(const nav of data.navis) {
        let pid = `<td style="width: 10%">${nav.id}</td>`;
        let pseq = `<td style="width: 10%"><input style="width: 100%" id="seq-${nav.id}" type="text" value="${nav.sequence}"></td>`;
        let pname = `<td style="width: 25%"><input style="width: 100%" id="name-${nav.id}" type="text" value="${nav.name}"></td>`;
        let ptarget = `<td style="width: 35%"><input style="width: 100%" id="dst-${nav.id}" type="text" value="${nav.target}"></td>`;
        let pop =
            `<td style="width: 20%">
                    <button class="btn btn-default glyphicon glyphicon-trash" onclick="deleteNavi(this,${nav.id});"></button>
                    <button class="btn btn-default glyphicon glyphicon-floppy-save" onclick="saveNavi(${nav.id});"></button>
                </td>`;

        $('#navis').append(`<tr>${pid}${pseq}${pname}${ptarget}${pop}</tr>`);
    }
}