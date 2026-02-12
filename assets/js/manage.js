let ModelHome = "/";
let ModelEdit = "/edit";
let Logout = "/api/login";
let Article = "/api/edit";
let Manage = "/api/manage";
let Settings = "/manage/setting";
let Thread = "/article";

let currentPage = 1;
let pageSize = 10;
let allPosts = [];
let deleteTarget = { element: null, id: null };

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
    let isHidden = $(`#eye-${id}`).hasClass("btn-show");
    fetch(`${Article}/${id}?hide=${!isHidden}`, {method: 'PUT'}).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        if(isHidden) {
            $(`#eye-${id}`).removeClass("btn-show").addClass("btn-hide").text("Hide");
        }
        else {
            $(`#eye-${id}`).removeClass("btn-hide").addClass("btn-show").text("Show");
        }
    }).catch(function(e) {
        alert(e);
    });
}

function confirmDelete(element, id) {
    deleteTarget.element = element;
    deleteTarget.id = id;
    $('#deleteModal').modal('show');
}

function deleteArticle() {
    let id = deleteTarget.id;
    
    fetch(`${Article}/${id}`, {method: 'DELETE'}).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        allPosts = allPosts.filter(function(post) {
            return post.id !== id;
        });

        let maxPage = Math.max(1, Math.ceil(allPosts.length / pageSize));
        if(currentPage > maxPage) {
            currentPage = maxPage;
        }

        $('#deleteModal').modal('hide');
        loadPage(currentPage);
    }).catch(function(e) {
        alert(e);
        $('#deleteModal').modal('hide');
    });
}

function editArticle(id) {
    redirect(`${ModelEdit}/${id}`);
}

function buildTable(posts) {
    $('#posts').empty();
    for(const post of posts) {
        let pid = `<td style="width: 5%">${post.id}</td>`;
        let ptitle = `<td style="width: 20%"><a target="_blank" href="${Thread}/${post.id}">${post.title}</a></td>`;
        let ptime = `<td style="width: 20%">${post.create_time}<br>${post.last_modified}</td>`;
        let arr = [];
        for(const tag of post.tags) {
            if(tag) {
                arr.push(`<span class="badge">${tag}</span>`);
            }
        }
        let ptag = `<td style="width: 20%">${arr.join('\n')}</td>`;
        
        let eyeText = post.hide ? "Show" : "Hide";
        let eyeClass = post.hide ? "btn-show" : "btn-hide";
        
        let pop =
            `<td>
                <button id="eye-${post.id}" class="btn btn-default btn-operation ${eyeClass}" onclick="toggleEye(${post.id});">${eyeText}</button>
                <button class="btn btn-default btn-operation btn-danger" onclick="confirmDelete(this, ${post.id});">Delete</button>
                <button type="button" class="btn btn-default btn-operation" onclick="editArticle(${post.id});">Edit</button>
            </td>`;

        $('#posts').append(`<tr>${pid}${ptitle}${ptime}${ptag}${pop}</tr>`);
    }
}

function buildPagination(totalPosts) {
    let totalPages = Math.ceil(totalPosts / pageSize);
    let html = '';
    
    // Previous button
    html += `<button class="btn btn-default" ${currentPage === 1 ? 'disabled' : ''} onclick="changePage(${currentPage - 1})">Prev</button>`;
    
    // Page numbers
    for(let i = 1; i <= totalPages; i++) {
        if(i === currentPage) {
            html += `<button class="btn btn-primary" disabled>${i}</button>`;
        } else {
            html += `<button class="btn btn-default" onclick="changePage(${i})">${i}</button>`;
        }
    }
    
    // Next button
    html += `<button class="btn btn-default" ${currentPage === totalPages ? 'disabled' : ''} onclick="changePage(${currentPage + 1})">Next</button>`;
    
    $('#pagination').html(html);
}

function changePage(page) {
    currentPage = page;
    loadPage(currentPage);
}

function loadPage(page) {
    let start = (page - 1) * pageSize;
    let end = start + pageSize;
    let pagePosts = allPosts.slice(start, end);
    buildTable(pagePosts);
    buildPagination(allPosts.length);
}

function init() {
    fetch(Manage).then(function(j) {
        return j.json();
    }).then(function(res) {
        if(res.code !== 0) {
            throw res.error;
        }
        allPosts = res.posts || [];
        loadPage(currentPage);
    }).catch(function(e) {
        alert(e);
    });
}

$(document).ready(function() {
    init();
    
    // Bind confirm delete button
    $('#confirmDelete').click(function() {
        deleteArticle();
    });
});
