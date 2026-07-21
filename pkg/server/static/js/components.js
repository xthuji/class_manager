// ==================== 通用 UI 组件模块 ====================
// 包含: 模态框、多选/单选下拉框、批量操作、分页、表格渲染等可复用组件
// 依赖: utils.js, icons.js
// 注意: 保持 Safari 13 (macOS 10.15) 兼容

// ==================== 模态框 ====================
function showModal(title, bodyHtml, size) {
  size = size || '';
  var container = $('#modalContainer');
  container.innerHTML =
    '<div class="modal-overlay" onclick="if(event.target===this)closeModal()">' +
      '<div class="modal ' + (size === 'lg' ? 'modal-lg' : '') + '">' +
        '<div class="modal-header">' +
          '<h3>' + escapeHtml(title) + '</h3>' +
          '<button class="modal-close" onclick="closeModal()">' + ICONS.close + '</button>' +
        '</div>' +
        '<div class="modal-body">' + bodyHtml + '</div>' +
      '</div>' +
    '</div>';
  initDatePickerPolyfill(container);
}

function closeModal() { $('#modalContainer').innerHTML = ''; }

function confirmModal(title, message, onConfirm, confirmText) {
  confirmText = confirmText || '确认';
  showModal(title,
    '<p style="margin-bottom:16px;">' + escapeHtml(message) + '</p>' +
    '<div class="flex gap-3" style="justify-content:flex-end;">' +
      '<button class="btn btn-secondary" onclick="closeModal()">取消</button>' +
      '<button class="btn btn-danger" id="confirmBtn">' + confirmText + '</button>' +
    '</div>');
  $('#confirmBtn').onclick = function() { closeModal(); onConfirm(); };
}

function showToast(msg, type) {
  type = type || 'info';
  var colors = { info: 'var(--blue)', success: 'var(--green)', error: 'var(--red)', warning: 'var(--amber)' };
  var toast = document.createElement('div');
  toast.style.cssText = 'position:fixed;top:20px;right:20px;z-index:200;padding:12px 20px;border-radius:8px;color:#fff;background:' + (colors[type] || colors.info) + ';box-shadow:0 4px 12px rgba(0,0,0,0.2);font-size:14px;';
  toast.textContent = msg;
  document.body.appendChild(toast);
  setTimeout(function() { toast.remove(); }, 3000);
}

// 生成表单模态框的底部按钮（取消 + 提交），减少重复代码
function renderFormFooter(saveHandler, saveText) {
  return '<div class="modal-footer">' +
    '<button class="btn btn-secondary" onclick="closeModal()">取消</button>' +
    '<button class="btn btn-primary" onclick="' + saveHandler + '">' + (saveText || '保存') + '</button>' +
  '</div>';
}

// ==================== 多选下拉框组件 ====================
function createMultiSelectHTML(id, placeholder) {
  return '<div class="multi-select" id="' + id + '" data-placeholder="' + escapeHtml(placeholder) + '">' +
    '<div class="multi-select-trigger" onclick="toggleMultiSelect(\'' + id + '\', event)">' +
      '<span class="multi-select-text">' + escapeHtml(placeholder) + '</span>' + ICONS.chevron +
    '</div>' +
    '<div class="multi-select-menu">' +
      '<label class="multi-select-item multi-select-all"><input type="checkbox" class="ms-all-cb" onchange="onMultiSelectAll(\'' + id + '\')"> <span>全部</span></label>' +
      '<div class="multi-select-options" id="' + id + '_options"></div>' +
    '</div>' +
  '</div>';
}

function populateMultiSelect(id, options, initialValues) {
  var container = $('#' + id);
  if (!container) return;
  var optsContainer = $('#' + id + '_options');
  var allCb = container.querySelector('.ms-all-cb');
  optsContainer.innerHTML = options.map(function(o) {
    return '<label class="multi-select-item"><input type="checkbox" value="' + escapeHtml(String(o.value)) + '" onchange="onMultiSelectItem(\'' + id + '\')"> <span>' + escapeHtml(o.label) + '</span></label>';
  }).join('');
  var items = optsContainer.querySelectorAll('input[type="checkbox"]');
  container._isDirty = false;
  if (initialValues && initialValues.length > 0) {
    var initialSet = {};
    initialValues.forEach(function(v) { initialSet[String(v)] = true; });
    items.forEach(function(cb) {
      cb.checked = !!initialSet[cb.value];
    });
    allCb.checked = items.length > 0 && Array.from(items).every(function(cb) { return cb.checked; });
    container._isDirty = true;
  } else {
    allCb.checked = true;
    items.forEach(function(cb) { cb.checked = true; });
  }
  updateMultiSelectText(id);
}

function toggleMultiSelect(id, event) {
  event.stopPropagation();
  var container = $('#' + id);
  container.classList.toggle('open');
  document.querySelectorAll('.multi-select.open').forEach(function(el) {
    if (el.id !== id) el.classList.remove('open');
  });
}

function onMultiSelectAll(id) {
  var container = $('#' + id);
  var checked = container.querySelector('.ms-all-cb').checked;
  container.querySelectorAll('.multi-select-options input[type="checkbox"]').forEach(function(cb) { cb.checked = checked; });
  updateMultiSelectText(id);
  container._isDirty = true;
  if (container && container._onchange) container._onchange();
}

function onMultiSelectItem(id) {
  var container = $('#' + id);
  var allCb = container.querySelector('.ms-all-cb');
  var items = container.querySelectorAll('.multi-select-options input[type="checkbox"]');
  var checkedCount = Array.from(items).filter(function(cb) { return cb.checked; }).length;
  allCb.checked = items.length > 0 && checkedCount === items.length;
  updateMultiSelectText(id);
  container._isDirty = true;
  if (container && container._onchange) container._onchange();
}

function updateMultiSelectText(id) {
  var container = $('#' + id);
  var placeholder = container.dataset.placeholder;
  var items = container.querySelectorAll('.multi-select-options input[type="checkbox"]');
  var checkedItems = Array.from(items).filter(function(cb) { return cb.checked; });
  var textEl = container.querySelector('.multi-select-text');
  if (checkedItems.length === 0) {
    textEl.textContent = placeholder;
  } else if (checkedItems.length === items.length) {
    textEl.textContent = '全部 (' + checkedItems.length + ')';
  } else if (checkedItems.length <= 2) {
    textEl.textContent = checkedItems.map(function(cb) { return cb.nextElementSibling.textContent; }).join(', ');
  } else {
    textEl.textContent = '已选 ' + checkedItems.length + ' 项';
  }
}

// 获取选中的值；全部选中时返回所有选中项
function getMultiSelectValues(id) {
  var container = $('#' + id);
  if (!container) return [];
  var items = container.querySelectorAll('.multi-select-options input[type="checkbox"]');
  var checkedItems = Array.from(items).filter(function(cb) { return cb.checked; });
  return checkedItems.map(function(cb) { return cb.value; });
}

// 获取多选下拉框的过滤参数
// 返回值:
// - null: 用户未手动操作，处于默认状态，不应用过滤
// - []: 用户手动操作后全选，不应用过滤（空列表参数）
// - [value1, value2, ...]: 用户手动操作后选择了部分选项，应用过滤
function getMultiSelectFilterParams(id) {
  var container = $('#' + id);
  if (!container) return null;
  if (!container._isDirty) return null;
  var items = container.querySelectorAll('.multi-select-options input[type="checkbox"]');
  var checkedItems = Array.from(items).filter(function(cb) { return cb.checked; });
  if (checkedItems.length === items.length) return [];
  return checkedItems.map(function(cb) { return cb.value; });
}

// 获取选中的值（不做"全部"归一化，用于需要精确知道哪些被选中的场景）
function getMultiSelectRawValues(id) {
  var container = $('#' + id);
  if (!container) return [];
  return Array.from(container.querySelectorAll('.multi-select-options input[type="checkbox"]:checked')).map(function(cb) { return cb.value; });
}

// 获取多选下拉框选中项的文本
function getMultiSelectText(id) {
  var container = $('#' + id);
  if (!container) return '';
  var items = container.querySelectorAll('.multi-select-options input[type="checkbox"]:checked');
  var labels = [];
  items.forEach(function(cb) {
    var label = cb.nextElementSibling;
    if (label) labels.push(label.textContent);
  });
  return labels.join('、') || '';
}

// ==================== 单选下拉框组件（视觉同多选，行为为单选） ====================
function createSingleSelectHTML(id, placeholder) {
  return '<div class="multi-select" id="' + id + '" data-placeholder="' + escapeHtml(placeholder) + '">' +
    '<div class="multi-select-trigger" onclick="toggleMultiSelect(\'' + id + '\', event)">' +
      '<span class="multi-select-text">' + escapeHtml(placeholder) + '</span>' + ICONS.chevron +
    '</div>' +
    '<div class="multi-select-menu">' +
      '<div class="multi-select-options" id="' + id + '_options"></div>' +
    '</div>' +
  '</div>';
}

function populateSingleSelect(id, options, initialValue) {
  var optsContainer = $('#' + id + '_options');
  if (!optsContainer) return;
  optsContainer.innerHTML = options.map(function(o) {
    return '<label class="multi-select-item"><input type="radio" name="' + id + '" value="' + escapeHtml(String(o.value)) + '" onchange="onSingleSelectChange(\'' + id + '\')" ' + (String(o.value) === String(initialValue) ? 'checked' : '') + '> <span>' + escapeHtml(o.label) + '</span></label>';
  }).join('');
  updateSingleSelectText(id);
}

function onSingleSelectChange(id) {
  updateSingleSelectText(id);
  var container = $('#' + id);
  if (container && container._onchange) container._onchange();
  container.classList.remove('open');
}

function updateSingleSelectText(id) {
  var container = $('#' + id);
  if (!container) return;
  var checked = container.querySelector('input[name="' + id + '"]:checked');
  var textEl = container.querySelector('.multi-select-text');
  textEl.textContent = checked ? checked.nextElementSibling.textContent : container.dataset.placeholder;
}

function getSingleSelectValue(id) {
  var container = $('#' + id);
  if (!container) return '';
  var checked = container.querySelector('input[name="' + id + '"]:checked');
  return checked ? checked.value : '';
}

// 点击外部关闭下拉
document.addEventListener('click', function(e) {
  document.querySelectorAll('.multi-select.open').forEach(function(el) {
    if (!el.contains(e.target)) el.classList.remove('open');
  });
});

// ==================== 分页 ====================
// 各分页页面每页条数（可由用户在页面上调整）
var pageSizes = {
  loadStudents: 10,
  searchStudents: 10,
  loadCoursesTable: 10,
  loadRecharges: 10,
  loadHourRecords: 20,
  loadNotifications: 10,
  loadLogs: 10,
};

// 切换每页条数：更新全局 pageSizes 并重新加载第一页
function changePageSize(callbackName, size) {
  pageSizes[callbackName] = parseInt(size, 10) || 10;
  if (typeof window[callbackName] === 'function') {
    window[callbackName](1);
  }
}

function renderPagination(page, totalPages, callbackName, pageSize, total) {
  if (totalPages <= 1 && (!total || total <= 0)) return '';
  var ps = pageSize || pageSizes[callbackName] || 10;
  var sizeOptions = [10, 20, 50, 100];
  if (sizeOptions.indexOf(ps) === -1) {
    sizeOptions.unshift(ps);
    sizeOptions.sort(function(a, b) { return a - b; });
  }

  var html = '<div class="pagination-bar">' +
    '<div class="pagination-info">' +
      (total != null ? '<span class="pagination-total">共 ' + total + ' 条</span>' : '') +
      '<span class="pagination-size"><span>每页</span>' +
      '<select class="form-select page-size-select" onchange="changePageSize(\'' + callbackName + '\', this.value)">';
  for (var i = 0; i < sizeOptions.length; i++) {
    html += '<option value="' + sizeOptions[i] + '"' + (sizeOptions[i] === ps ? ' selected' : '') + '>' + sizeOptions[i] + '</option>';
  }
  html += '</select><span>条</span></span></div>';

  html += '<div class="pagination-controls">';
  if (page > 1) {
    html += '<button class="btn btn-sm btn-secondary" onclick="' + callbackName + '(' + (page - 1) + ')">上一页</button>';
  } else {
    html += '<button class="btn btn-sm btn-secondary" disabled>上一页</button>';
  }
  html += '<span class="pagination-page-num">第 ' + page + ' / ' + totalPages + ' 页</span>';
  if (page < totalPages) {
    html += '<button class="btn btn-sm btn-primary" onclick="' + callbackName + '(' + (page + 1) + ')">下一页</button>';
  } else {
    html += '<button class="btn btn-sm btn-primary" disabled>下一页</button>';
  }
  html += '</div></div>';
  return html;
}

// ==================== 批量选择/删除 ====================
// 批量删除配置（每个页面注册自己的配置）
var batchConfigs = {};

function registerBatchConfig(key, apiCall, reloadFn, label, extraActions) {
  batchConfigs[key] = { apiCall: apiCall, reloadFn: reloadFn, label: label, extraActions: extraActions || [] };
}

// 全选/取消全选当前页的所有行
function toggleSelectAll(key, checked) {
  var checkboxes = document.querySelectorAll('.row-checkbox-' + key);
  for (var i = 0; i < checkboxes.length; i++) {
    checkboxes[i].checked = checked;
  }
  updateBatchBar(key);
}

// 单行选择变化时更新批量操作栏
function toggleRowSelection(key) {
  updateBatchBar(key);
}

// 获取当前选中的 ID 列表
function getSelectedIds(key) {
  var ids = [];
  var checkboxes = document.querySelectorAll('.row-checkbox-' + key + ':checked');
  for (var i = 0; i < checkboxes.length; i++) {
    var parts = (checkboxes[i].value || '').split(',');
    for (var j = 0; j < parts.length; j++) {
      var n = parseInt(parts[j], 10);
      if (!isNaN(n)) ids.push(n);
    }
  }
  return ids;
}

// 更新批量操作栏的显示状态
function updateBatchBar(key) {
  var bar = document.getElementById('batch-bar-' + key);
  if (!bar) return;
  var count = document.querySelectorAll('.row-checkbox-' + key + ':checked').length;
  if (count > 0) {
    bar.style.display = 'flex';
    var countEl = bar.querySelector('.batch-count');
    if (countEl) countEl.textContent = count;
  } else {
    bar.style.display = 'none';
  }
}

// 执行批量删除
function batchDelete(key) {
  var config = batchConfigs[key];
  if (!config) return;
  var ids = getSelectedIds(key);
  if (ids.length === 0) {
    showToast('请先选择要删除的记录', 'warning');
    return;
  }
  confirmModal('批量删除', '确定要删除选中的 ' + ids.length + ' 条' + config.label + '吗？此操作不可恢复。', function() {
    config.apiCall(ids).then(function(res) {
      showToast('成功删除 ' + (res.deleted || ids.length) + ' 条' + config.label, 'success');
      config.reloadFn(1);
    }).catch(function(e) {
      showToast('删除失败: ' + e.message, 'error');
    });
  }, '批量删除');
}

// 执行批量额外操作（如标记已读）
function batchAction(key, actionIndex) {
  var config = batchConfigs[key];
  if (!config || !config.extraActions || !config.extraActions[actionIndex]) return;
  var ids = getSelectedIds(key);
  if (ids.length === 0) {
    showToast('请先选择记录', 'warning');
    return;
  }
  config.extraActions[actionIndex].action(ids);
}

// 渲染批量操作栏（页面加载时调用 registerBatchConfig 注册配置）
function renderBatchBar(key) {
  var config = batchConfigs[key];
  var extraHtml = '';
  if (config && config.extraActions) {
    for (var i = 0; i < config.extraActions.length; i++) {
      var action = config.extraActions[i];
      extraHtml += '<button class="btn ' + action.btnClass + ' btn-sm" onclick="batchAction(\'' + key + '\', ' + i + ')">' + action.label + '</button>';
    }
  }
  return '<div id="batch-bar-' + key + '" class="batch-bar" style="display:none;">' +
    '<span class="batch-info">已选择 <span class="batch-count">0</span> 项</span>' +
    extraHtml +
    '<button class="btn btn-danger btn-sm" onclick="batchDelete(\'' + key + '\')">批量删除</button>' +
    '<button class="btn btn-secondary btn-sm" onclick="toggleSelectAll(\'' + key + '\', false)">取消选择</button>' +
  '</div>';
}

// ==================== 通用状态/表格渲染助手 ====================
// 加载中占位
function renderLoading() {
  return '<div class="loading"><div class="spinner"></div></div>';
}

// 空数据占位
function renderEmptyState(icon, text) {
  return '<div class="card"><div class="empty-state">' + (icon || '') + '<p>' + escapeHtml(text) + '</p></div></div>';
}

// 错误占位
function renderErrorState(msg) {
  return '<div class="empty-state"><p>加载失败: ' + escapeHtml(msg) + '</p></div>';
}

// 渲染带批量选择 + 分页的标准数据表格
// opts: { batchKey, headers, rowsHtml, page, totalPages, callbackName, pageSize, total, headerHtml }
// headers: 数组，每项为表头文本（首列为 checkbox 由本函数自动生成）
// headerHtml: 可选，自定义完整表头 HTML（覆盖 headers）
function renderDataTable(opts) {
  var batchKey = opts.batchKey;
  var cbHeader = batchKey
    ? '<th style="width:36px;"><input type="checkbox" onchange="toggleSelectAll(\'' + batchKey + '\', this.checked)"></th>'
    : '';
  var headerCells;
  if (opts.headerHtml) {
    headerCells = opts.headerHtml;
  } else {
    headerCells = (opts.headers || []).map(function(h) { return '<th>' + h + '</th>'; }).join('');
  }
  return '<div class="card">' +
    (batchKey ? renderBatchBar(batchKey) : '') +
    '<table class="data-table">' +
      '<thead><tr>' + cbHeader + headerCells + '</tr></thead>' +
      '<tbody>' + (opts.rowsHtml || '') + '</tbody>' +
    '</table>' +
    renderPagination(opts.page, opts.totalPages, opts.callbackName, opts.pageSize, opts.total) +
  '</div>';
}

// 生成带 checkbox 的表格行首单元格
function renderRowCheckbox(batchKey, value) {
  return '<td><input type="checkbox" class="row-checkbox-' + batchKey + '" value="' + escapeHtml(String(value)) + '" onchange="toggleRowSelection(\'' + batchKey + '\')"></td>';
}

// ==================== 学生筛选联动助手（统一抽象） ====================
// 统一处理"按课程联动学生"逻辑，消除三处重复实现
// 参数:
//   courseSelectId: 课程下拉框元素 ID（多选或单选）
//   studentSelectId: 待填充的学生下拉框元素 ID（多选）
//   allStudents: 全部学生数组
//   opts: { multi: true/false(课程是否多选), activeOnly: bool(仅显示在读), autoSelect: bool(是否默认全选学生) }
async function linkStudentsByCourse(courseSelectId, studentSelectId, allStudents, opts) {
  opts = opts || {};
  var container = $('#' + studentSelectId);
  if (!container) return;

  var labelFn = opts.labelFn || function(s) { return s.name + ' (' + s.student_id + ')'; };
  var courseIds;
  if (opts.multi) {
    // 多选课程: 使用过滤参数（区分"未操作"与"全选"）
    var courseParams = getMultiSelectFilterParams(courseSelectId);
    if (courseParams === null || courseParams.length === 0) {
      // 未操作或全选 -> 显示全部学生
      var allOpts = allStudents.map(function(s) {
        return { value: s.id, label: labelFn(s) };
      });
      var initVals = opts.autoSelect ? allStudents.map(function(s) { return s.id; }) : [];
      populateMultiSelect(studentSelectId, allOpts, initVals);
      return;
    }
    courseIds = courseParams.map(function(v) { return parseInt(v, 10); });
  } else {
    // 单选课程
    var cidStr = getSingleSelectValue(courseSelectId);
    var cid = cidStr ? parseInt(cidStr, 10) : 0;
    if (!cid) {
      populateMultiSelect(studentSelectId, [], []);
      return;
    }
    courseIds = [cid];
  }

  // 收集所选课程的学生 ID
  var studentIdSet = {};
  for (var i = 0; i < courseIds.length; i++) {
    try {
      var students = await API.course.getStudents(courseIds[i]);
      if (students && students.length > 0) {
        for (var j = 0; j < students.length; j++) {
          studentIdSet[students[j].id] = true;
        }
      }
    } catch (e) { /* 忽略单个课程查询失败 */ }
  }

  // 筛选出在所选课程中的学生
  var filtered = allStudents.filter(function(s) {
    if (!studentIdSet[s.id]) return false;
    if (opts.activeOnly && s.is_dropped) return false;
    return true;
  });

  var studentOpts = filtered.map(function(s) {
    return { value: s.id, label: labelFn(s) };
  });
  var initVals = opts.autoSelect ? filtered.map(function(s) { return s.id; }) : [];
  populateMultiSelect(studentSelectId, studentOpts, initVals);
}

// 按学生状态过滤多选下拉框中的选项（启用/禁用 checkbox）
// 参数:
//   studentSelectId: 学生多选下拉框 ID
//   allStudents: 全部学生数组
//   status: '' (全部) / '0' (在读) / '1' (已退学)
function filterStudentsByStatus(studentSelectId, allStudents, status) {
  var container = $('#' + studentSelectId);
  if (!container) return;
  var checkboxes = container.querySelectorAll('.multi-select-options input[type="checkbox"]');
  checkboxes.forEach(function(cb) {
    var studentId = parseInt(cb.value, 10);
    var student = null;
    for (var i = 0; i < allStudents.length; i++) {
      if (allStudents[i].id === studentId) { student = allStudents[i]; break; }
    }
    if (student) {
      if (status === '') {
        cb.disabled = false;
      } else {
        var isDropped = student.is_dropped ? '1' : '0';
        cb.disabled = isDropped !== status;
        if (cb.disabled) cb.checked = false;
      }
    }
  });
  updateMultiSelectText(studentSelectId);
}
