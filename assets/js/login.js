function showOk(msg) {
    $('#myStatus').text('Status');
    $('#myModal').html(`<span style="color: green">${msg}</span>`);
    $('#modal').modal('show');
    setTimeout(function() {
        $('#modal').modal('hide');
    }, 2000);
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
    let hostname = location.protocol+'//'+location.hostname+(location.port ? ':'+location.port: '');
    window.location.href = `${hostname}/${dst}`;
}

function login() {
    let user = $('#username').val();
    let pass = $('#password').val();
    if(user === '' || pass === '') {
        throw 'uername and password can NOT be empty';
    }
    fetch('/api/login', {
        method: 'POST',
        body: JSON.stringify({
            username: user,
            password: pass
        }),
        headers: {
            'Content-Type': 'application/json'
        }
    }).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        redirect('/manage');
    }).catch(function(e) {
        showError(e);
    });
}