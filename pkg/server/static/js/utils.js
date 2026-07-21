// ==================== 工具函数模块 ====================
// 所有通用工具函数集中于此，供 app.js 和 components.js 共用
// 注意: 保持 Safari 13 (macOS 10.15) 兼容，不使用可选链(?.)、空值合并(??)等 ES2020+ 语法

// DOM 查询快捷函数
var $ = function(sel) { return document.querySelector(sel); };

// HTML 转义，防止 XSS
var escapeHtml = function(str) {
  if (str == null) return '';
  return String(str).replace(/[&<>"']/g, function(c) {
    return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
  });
};

// 格式化日期时间为本地字符串
var formatDate = function(dateStr) {
  if (!dateStr) return '-';
  var d = new Date(dateStr);
  return isNaN(d) ? dateStr : d.toLocaleString('zh-CN');
};

// 格式化课时数值，小数点后为0则显示整数
function formatHours(hours) {
  var h = parseFloat(hours) || 0;
  if (Math.floor(h) === h) {
    return Math.floor(h).toString();
  }
  return h.toFixed(1);
}

// 获取今天的日期字符串 (YYYY-MM-DD)
var todayStr = function() { return new Date().toISOString().split('T')[0]; };

// 星期常量与标签转换
var WEEKDAYS = [
  { v: 1, l: '周一' }, { v: 2, l: '周二' }, { v: 3, l: '周三' },
  { v: 4, l: '周四' }, { v: 5, l: '周五' }, { v: 6, l: '周六' },
  { v: 0, l: '周日' }
];
var weekdayLabel = function(d) {
  var w = WEEKDAYS.find(function(wk) { return wk.v === d; });
  return w ? w.l : '';
};

// 根据开始时间判断时段（上午/中午/下午/晚上）
var getPeriod = function(start) {
  var h = parseInt(start ? start.split(':')[0] : '0', 10);
  if (h < 12) return '上午';
  if (h < 14) return '中午';
  if (h < 18) return '下午';
  return '晚上';
};

// 根据开始/结束时间计算课时数
var calcHours = function(start, end) {
  var parts = start.split(':'), parts2 = end.split(':');
  var sh = parseInt(parts[0], 10), sm = parseInt(parts[1], 10);
  var eh = parseInt(parts2[0], 10), em = parseInt(parts2[1], 10);
  var d = (eh * 60 + em - sh * 60 - sm) / 60;
  return d > 0 ? Math.round(d * 10) / 10 : 0;
};

// 将 ISO 周字符串（如 "2026-W03"）转换为周一~周日的日期范围文本（如 "01-19~01-25"）
var weekRangeLabel = function(weekStr) {
  if (!weekStr) return '';
  var m = weekStr.match(/^(\d{4})-W(\d{2})$/);
  if (!m) return weekStr;
  var year = parseInt(m[1], 10), week = parseInt(m[2], 10);
  var jan1 = new Date(year, 0, 1);
  var jan1Day = jan1.getDay();
  var firstMonday = new Date(jan1);
  var offsetToMonday = jan1Day === 1 ? 0 : (8 - jan1Day);
  firstMonday.setDate(jan1.getDate() + offsetToMonday);
  if (jan1Day >= 1 && jan1Day <= 4) {
    firstMonday.setDate(jan1.getDate() - (jan1Day - 1));
  }
  var monday = new Date(firstMonday);
  monday.setDate(firstMonday.getDate() + (week - 1) * 7);
  var sunday = new Date(monday);
  sunday.setDate(monday.getDate() + 6);
  return fmtMD(monday) + '~' + fmtMD(sunday);
};

// 计算某日期所在周的周一日期对象
var mondayOf = function(dateStr) {
  var dt = new Date(dateStr);
  var mon = new Date(dt);
  mon.setDate(dt.getDate() - ((dt.getDay() + 6) % 7));
  return mon;
};

// 格式化日期为 MM-DD
var fmtMD = function(d) {
  return padStart(String(d.getMonth() + 1), 2, '0') + '-' + padStart(String(d.getDate()), 2, '0');
};

// Safari 13 兼容：安全获取错误消息（错误对象的 message 属性可能为 null）
function getErrorMessage(err) {
  if (err == null) return '未知错误';
  if (typeof err.message === 'string') return err.message;
  if (typeof err === 'string') return err;
  return String(err);
}

// Safari 13 不支持 String.prototype.padStart (ES2017)，提供兼容实现
function padStart(str, len, char) {
  str = String(str);
  char = char || '0';
  while (str.length < len) str = char + str;
  return str;
}
