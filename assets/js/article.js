function create_toc() {
    let id_counter = 1;
    // 自动为 content 中的每个 h 标签生成 id
    $('#article').find('h1, h2, h3, h4, h5, h6').each(function() {
        let id = $(this).attr('id');
        if (id && id.toLowerCase() === 'title')
            return;

        $(this).attr('id', id_counter++);
    });

    // 创建目录
    var toc = $('#toc'); // 获取目标 div
    var list = $('<ul></ul>'); // 创建一个无序列表

    // 用一个数组来保存每个标题级别的最新列表项
    var stack = [];

    $('#article').find('h1, h2, h3, h4, h5, h6').each(function() {
         let id = $(this).attr('id');
        if (id && id.toLowerCase() === 'title')
            return;
        var tag = $(this)[0].tagName; // 获取标签类型 (h1, h2, h3, ...)
        var text = $(this).text(); // 获取标签文本内容

        // 生成目录项
        var listItem = $('<li></li>');
        var link = $('<a></a>').text(text).attr('href', '#' + id);
        listItem.append(link);

        // 获取当前标签的级别
        var level = parseInt(tag[1]); // 获取 h1 -> 1, h2 -> 2, ..., h6 -> 6

        // 确保每个标题级别的子目录都添加到正确的父级
        while (stack.length > 0 && stack[stack.length-1][0] >= level) {
            stack.pop(); // 弹出比当前级别大的所有级别，找到正确的父级
        }

        // 如果栈为空，说明是一个顶级目录
        if (stack.length === 0) {
            list.append(listItem); // 直接添加到根目录
        } else {
            // 否则，找到父级并添加子目录
//            var parentList = stack[stack.length - 1][1].find('ul');
//            if (parentList.length === 0) {
                parentList = $('<ul></ul>');
                stack[stack.length - 1][1].append(parentList);
//            }
            parentList.append(listItem);
        }

        // 将当前目录项压入栈中
        stack.push([level, listItem]);
    });

    // 将生成的目录插入到 toc div 中
    toc.append(list);

    // 点击按钮切换目录的显示与隐藏
    $('#toggle-toc').click(toggle_menu);
}

function toggle_menu() {
    $('#toc').toggleClass('collapsed'); // 切换 "collapsed" 类
    var isCollapsed = $('#toc').hasClass('collapsed');

    // 根据目录是否收起更新按钮文本
    if (isCollapsed) {
        $(this).text('展开'); // 收起时按钮显示 "展开"
    } else {
        $(this).text('收起'); // 展开时按钮显示 "收起"
    }
}
