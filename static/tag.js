const tagName = $('#name');
const rename = $('#rename');
const name_input = $('#name-input');
const rename_block = $('#rename-block');
const cancel = $('#cancel');
const ok = $('#ok');
const check_tags_btn = $('#check-tags-btn');

const tag_name = getUrlParam('name');
tagName.text(tag_name);

ajaxGet(`/api/tag/${tag_name}/notes`, null, that => {
  $('#tag-name').show();
  $('#count-block').show();
  const notes = that.response;
  if (!notes) {
    $('#count').text(0);
    return;
  }
  $('#count').text(notes.length);

  notes.sort((a, b) => {
    if (a.UpdatedAt > b.UpdatedAt) return 1;
    if (a.UpdatedAt < b.UpdatedAt) return -1;
    return 0;
  });
  notes.forEach(addNoteElem);
}, () => {
  // onloadend
  $('#loading').hide();
});

rename.click(event => {
  event.preventDefault();
  rename_toggle();
  name_input.focus();
});
cancel.click(event => {
  event.preventDefault();
  rename_toggle();
});

function rename_toggle() {
  rename.toggle();
  rename_block.toggle();
}

ok.click(() => {
  const new_name = getTag(name_input);
  if (new_name == '') {
    insertErrorAlert('标签名称不可空白');
    name_input.focus();
    return;
  }

  const form = new FormData();
  const old_name = decodeURIComponent(tag_name);
  form.append('old-name', old_name);
  form.append('new-name', new_name);

  ajaxPut(form, '/api/tag/', ok, () => {
    rename_toggle();
    tagName.val(new_name);
    $('#tag-name').hide();
    $('.alert').hide();
    $('ul').hide();
    insertSuccessAlert(`正在重命名: ${old_name} --> ${new_name}`);
    insertInfoAlert('重命名成功时会自动刷新页面');
    window.setTimeout(function(){
      window.location = '/html/tag/?name=' + encodeURIComponent(new_name)
    }, 5000);
  });  
});

const confirm_block = $('#confirm-block');
const delete_btn = $('#delete');
const yes_btn = $('#yes');
const no_btn = $('#no');

// 删除按钮
delete_btn.click(delete_toggle);

// 取消删除
no_btn.click(delete_toggle);

function delete_toggle(event) {
  event.preventDefault();
  delete_btn.toggle();
  confirm_block.toggle();
}

// 确认删除
yes_btn.click(event => {
  event.preventDefault();
  ajaxDelete('/api/tag/'+encodeURIComponent(tag_name), yes_btn, function() {
    $('.alert').hide();
    $('#head-buttons').hide();
    $('p').hide();
    $('ul').hide();
    insertSuccessAlert(`标签 ${tag_name} 已删除`);
  });
});

// 对标签文本框内的字符串进行处理，提取出一个标签。
function getTag(tagsElem) {
  let trimmed = tag_replace(tagsElem.val());
  if (trimmed.length == 0) {
    return '';
  }
  let arr = trimmed.split(/ +/);
  return arr[0];
}

name_input.focus(() => {
  ok.hide();
  check_tags_btn.show();
});
name_input.blur(() => {
  const tag = getTag(name_input);
  if (tag) {
    name_input.val('#' + tag);
    ok.show();  
    check_tags_btn.hide();
  }
});