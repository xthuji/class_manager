// ==================== Class Manager 前端主逻辑 ====================
// 依赖: js/utils.js, js/icons.js, js/components.js (须在 HTML 中先于本文件加载)
// 本文件包含: API 客户端、路由、图表渲染、各业务页面逻辑、日期选择器 Polyfill、初始化
// 注意: 保持 Safari 13 (macOS 10.15) 兼容，不使用可选链(?.)、空值合并(??)等 ES2020+ 语法

// ==================== API 客户端 ====================
const API = {
  async request(url, opts = {}) {
    const res = await fetch(url, {
      headers: { 'Content-Type': 'application/json', ...opts.headers },
      ...opts,
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({ error: res.statusText }));
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    return res.json();
  },
  dashboard: {
    stats: () => API.request('/api/dashboard/stats'),
    charts: () => API.request('/api/dashboard/charts'),
    studentChart: (studentIds, courseIds, startDate, endDate) => {
      const q = new URLSearchParams();
      studentIds.forEach(id => q.append('student_id', String(id)));
      courseIds.forEach(id => q.append('course_id', String(id)));
      if (startDate) q.set('start_date', startDate);
      if (endDate) q.set('end_date', endDate);
      return API.request(`/api/dashboard/student-chart?${q}`);
    },
  },
  student: {
    list: (params = {}) => {
      const q = new URLSearchParams();
      if (params.page) q.set('page', params.page);
      if (params.page_size) q.set('page_size', params.page_size);
      return API.request(`/api/students?${q}`);
    },
    get: (id) => API.request(`/api/students/${id}`),
    create: (data) => API.request('/api/students', { method: 'POST', body: JSON.stringify(data) }),
    update: (id, data) => API.request(`/api/students/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id) => API.request('/api/students/batch-delete', { method: 'POST', body: JSON.stringify({ ids: [id] }) }),
    generateID: () => API.request('/api/students/generate-id'),
    search: (data) => API.request('/api/students/search', { method: 'POST', body: JSON.stringify(data) }),
    batchDelete: (ids) => API.request('/api/students/batch-delete', { method: 'POST', body: JSON.stringify({ ids }) }),
  },
  course: {
    list: (params = {}) => {
      const q = new URLSearchParams();
      if (params.page) q.set('page', params.page);
      if (params.page_size) q.set('page_size', params.page_size);
      return API.request(`/api/courses?${q}`);
    },
    get: (id) => API.request(`/api/courses/${id}`),
    create: (data) => API.request('/api/courses', { method: 'POST', body: JSON.stringify(data) }),
    update: (id, data) => API.request(`/api/courses/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id) => API.request('/api/courses/batch-delete', { method: 'POST', body: JSON.stringify({ ids: [id] }) }),
    batchDelete: (ids) => API.request('/api/courses/batch-delete', { method: 'POST', body: JSON.stringify({ ids }) }),
    getStudents: (id) => API.request(`/api/courses/${id}/students`),
    getByStudent: (sid) => API.request(`/api/students/${sid}/courses`),
    enroll: (data) => API.request('/api/courses/enroll', { method: 'POST', body: JSON.stringify(data) }),
    unenroll: (data) => API.request('/api/courses/unenroll', { method: 'POST', body: JSON.stringify(data) }),
    addHours: (id, hours) => API.request(`/api/courses/${id}/add-hours`, { method: 'POST', body: JSON.stringify({ hours }) }),
  },
  hourRecord: {
    list: (params = {}) => {
      const q = new URLSearchParams();
      if (params.student_ids) params.student_ids.forEach(id => q.append('student_ids', id));
      if (params.course_ids) params.course_ids.forEach(id => q.append('course_ids', id));
      if (params.start_date) q.set('start_date', params.start_date);
      if (params.end_date) q.set('end_date', params.end_date);
      if (params.student_status) q.set('student_status', params.student_status);
      if (params.page) q.set('page', params.page);
      if (params.page_size) q.set('page_size', params.page_size);
      return API.request(`/api/hour-records?${q}`);
    },
    batchCreate: (data) => API.request('/api/hour-records/batch', { method: 'POST', body: JSON.stringify(data) }),
    update: (id, data) => API.request(`/api/hour-records/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id) => API.request('/api/hour-records/batch-delete', { method: 'POST', body: JSON.stringify({ ids: [id] }) }),
    batchDelete: (ids) => API.request('/api/hour-records/batch-delete', { method: 'POST', body: JSON.stringify({ ids }) }),
    export: (data) => API.request('/api/hour-records/export', { method: 'POST', body: JSON.stringify(data) }),
  },
  hourRecharge: {
    list: (params = {}) => {
      const q = new URLSearchParams({ page: params.page || 1, page_size: params.page_size || 100 });
      if (params.student_id) q.set('student_id', params.student_id);
      if (params.start_date) q.set('start_date', params.start_date);
      if (params.end_date) q.set('end_date', params.end_date);
      return API.request(`/api/hour-recharges?${q}`);
    },
    create: (data) => API.request('/api/hour-recharges', { method: 'POST', body: JSON.stringify(data) }),
    update: (id, data) => API.request(`/api/hour-recharges/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    getByStudent: (sid) => API.request(`/api/hour-recharges/by-student/${sid}`),
    delete: (id) => API.request('/api/hour-recharges/batch-delete', { method: 'POST', body: JSON.stringify({ ids: [id] }) }),
    batchDelete: (ids) => API.request('/api/hour-recharges/batch-delete', { method: 'POST', body: JSON.stringify({ ids }) }),
  },
  schedule: {
    getByCourse: (cid) => API.request(`/api/schedules/by-course/${cid}`),
    create: (data) => API.request('/api/schedules', { method: 'POST', body: JSON.stringify(data) }),
    update: (id, data) => API.request(`/api/schedules/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    delete: (id) => API.request(`/api/schedules/${id}`, { method: 'DELETE' }),
  },
  notification: {
    list: (status = '', page = 1, pageSize = 10) => {
      const q = new URLSearchParams();
      if (status) q.set('status', status);
      q.set('page', page);
      q.set('page_size', pageSize);
      return API.request(`/api/notifications?${q}`);
    },
    unreadCount: () => API.request('/api/notifications/unread-count'),
    markRead: (id) => API.request(`/api/notifications/${id}/read`, { method: 'POST' }),
    batchRead: (ids) => API.request('/api/notifications/batch-read', { method: 'POST', body: JSON.stringify({ ids }) }),
    delete: (id) => API.request('/api/notifications/batch-delete', { method: 'POST', body: JSON.stringify({ ids: [id] }) }),
    batchDelete: (ids) => API.request('/api/notifications/batch-delete', { method: 'POST', body: JSON.stringify({ ids }) }),
    check: () => API.request('/api/notifications/check', { method: 'POST' }),
  },
  logs: {
    list: (params = {}) => {
      const q = new URLSearchParams({ page: params.page || 1, page_size: params.page_size || 200 });
      if (params.entity_types) params.entity_types.forEach(v => q.append('entity_types', v));
      if (params.operation_types) params.operation_types.forEach(v => q.append('operation_types', v));
      if (params.start_date) q.set('start_date', params.start_date);
      if (params.end_date) q.set('end_date', params.end_date);
      return API.request(`/api/logs?${q}`);
    },
    entities: () => API.request('/api/logs/entities'),
    delete: (id) => API.request('/api/logs/batch-delete', { method: 'POST', body: JSON.stringify({ ids: [id] }) }),
    batchDelete: (ids) => API.request('/api/logs/batch-delete', { method: 'POST', body: JSON.stringify({ ids }) }),
  },
  data: {
    exportAll: () => API.request('/api/data/export'),
    importAll: (data) => API.request('/api/data/import', { method: 'POST', body: JSON.stringify({ data }) }),
    saveToFile: (path, content) => API.request('/api/data/save-file', { method: 'POST', body: JSON.stringify({ path, content }) }),
  },
  calendar: {
    list: (month) => API.request(`/api/calendar?month=${month}`),
    holidays: (month) => API.request(`/api/holidays?month=${month}`),
  },
};

// ==================== 路由 ====================
const ROUTES = {
  '/': { title: '仪表盘', icon: 'dashboard', render: renderDashboard },
  '/students': { title: '学生管理', icon: 'students', render: renderStudents },
  '/courses': { title: '课程管理', icon: 'book', render: renderCourses },
  '/calendar': { title: '课程日历', icon: 'calendar', render: renderCourseCalendar },
  '/hour-records': { title: '课时记录', icon: 'clock', render: renderHourRecords },
  '/recharges': { title: '充值记录', icon: 'wallet', render: renderRecharges },
  '/notifications': { title: '通知中心', icon: 'bell', render: renderNotifications },
  '/logs': { title: '操作日志', icon: 'journal', render: renderLogs },
  '/backup': { title: '数据管理', icon: 'database', render: renderBackup },
};

let unreadCount = 0;

function router() {
  try {
    const hash = location.hash.slice(1) || '/';
    const route = ROUTES[hash];
    if (!route) { location.hash = '/'; return; }
    $('#pageTitle').textContent = route.title;
    document.querySelectorAll('.nav-item').forEach(el => {
      el.classList.toggle('active', el.dataset.path === hash);
    });
    // 异步渲染页面，捕获同步错误防止 WKWebView 崩溃
    try {
      var result = route.render();
      if (result && typeof result.catch === 'function') {
        result.catch(function(err) {
          console.error('页面渲染失败:', err);
          var body = $('#pageBody');
          if (body) body.innerHTML = '<div style="padding:20px;color:#dc2626;">页面加载失败: ' + escapeHtml(getErrorMessage(err)) + '</div>';
        });
      }
    } catch(e) {
      console.error('页面渲染同步错误:', e);
      var body = $('#pageBody');
      if (body) body.innerHTML = '<div style="padding:20px;color:#dc2626;">页面加载失败: ' + escapeHtml(String(e && e.message || e)) + '</div>';
    }
    initDatePickerPolyfill(document.getElementById('pageBody'));
  } catch(e) {
    console.error('router 错误:', e);
  }
}

function navigate(path) { location.hash = path; }

// ==================== 图表渲染 ====================
// 简易 SVG 柱状图生成器
function renderBarChart(data, opts = {}) {
  const w = opts.width || 600, h = opts.height || 200, pad = 30;
  const maxVal = Math.max(...data.map(d => d.value), 1);
  const barW = data.length > 0 ? (w - pad * 2) / data.length : 0;
  const barInnerW = Math.min(barW * 0.7, 30);
  const bars = data.map((d, i) => {
    const bh = (d.value / maxVal) * (h - pad * 2);
    const x = pad + i * barW + (barW - barInnerW) / 2;
    const y = h - pad - bh;
    const color = opts.color || '#3b82f6';
    const tipText = opts.tooltip ? opts.tooltip(d) : `${d.label || ''}: ${formatHours(d.value)}`;
    return `<rect x="${x}" y="${y}" width="${barInnerW}" height="${bh}" fill="${color}" rx="3"><title>${escapeHtml(tipText)}</title></rect>
            ${d.value > 0 ? `<text x="${x + barInnerW/2}" y="${y - 4}" text-anchor="middle" font-size="10" fill="#666">${formatHours(d.value)}</text>` : ''}`;
  }).join('');
  const labelStep = Math.ceil(data.length / 12);
  const xLabels = data.map((d, i) => {
    const x = pad + i * barW + barW / 2;
    return i % labelStep === 0 ? `<text x="${x}" y="${h - pad + 14}" text-anchor="middle" font-size="10" fill="#999">${escapeHtml(d.label || '')}</text>` : '';
  }).join('');
  return `<svg viewBox="0 0 ${w} ${h}" style="width:100%;height:auto;">
    <line x1="${pad}" y1="${h-pad}" x2="${w-pad}" y2="${h-pad}" stroke="#e5e7eb"/>
    <line x1="${pad}" y1="${pad}" x2="${pad}" y2="${h-pad}" stroke="#e5e7eb"/>
    ${bars}${xLabels}
  </svg>`;
}

// 多系列折线图生成器
function renderLineChart(rawData, opts = {}) {
  const w = opts.width || 800, h = opts.height || 300, pad = 40, legendH = 24;
  const dates = [...new Set(rawData.map(d => d.date))].sort();
  const courses = [...new Set(rawData.map(d => d.course_name))];
  const colors = ['#3b82f6', '#8b5cf6', '#10b981', '#f59e0b', '#ef4444', '#06b6d4', '#ec4899'];
  const series = courses.map((course, i) => ({
    name: course,
    color: colors[i % colors.length],
    values: dates.map(date => {
      const e = rawData.find(d => d.date === date && d.course_name === course);
      return e ? e.hours : 0;
    }),
  }));
  if (dates.length === 0) return '<p class="text-gray text-sm">暂无数据</p>';
  const maxVal = Math.max.apply(null, [].concat.apply([], series.map(s => s.values)).concat([1]));
  const chartH = h - pad - legendH;
  const xStep = dates.length > 1 ? (w - pad * 2) / (dates.length - 1) : 0;
  const labelStep = Math.ceil(dates.length / 8);
  const xLabels = dates.map((date, i) =>
    i % labelStep === 0 ? `<text x="${pad + i * xStep}" y="${h - pad + 14}" text-anchor="middle" font-size="9" fill="#999">${date.slice(5)}</text>` : ''
  ).join('');
  const yVals = [0, maxVal / 2, maxVal];
  const yLabels = yVals.map(v => `<text x="${pad - 6}" y="${chartH + pad - (v / maxVal) * chartH + 3}" text-anchor="end" font-size="9" fill="#999">${formatHours(v)}</text>`).join('');
  const lines = series.map(s => {
    const pts = s.values.map((v, i) => `${pad + i * xStep},${chartH + pad - (v / maxVal) * chartH}`).join(' ');
    const dots = s.values.map((v, i) => v > 0 ? `<circle cx="${pad + i * xStep}" cy="${chartH + pad - (v / maxVal) * chartH}" r="3" fill="${s.color}"><title>${dates[i]} ${s.name}: ${formatHours(v)} 课时</title></circle>` : '').join('');
    return `<polyline points="${pts}" fill="none" stroke="${s.color}" stroke-width="2"/>${dots}`;
  }).join('');
  const legend = series.map((s, i) =>
    `<g transform="translate(${pad + i * 90}, 8)"><rect width="10" height="10" fill="${s.color}" rx="2"/><text x="14" y="9" font-size="11" fill="#666">${escapeHtml(s.name)}</text></g>`
  ).join('');
  return `<svg viewBox="0 0 ${w} ${h}" style="width:100%;height:auto;">
    ${legend}
    <line x1="${pad}" y1="${chartH + pad}" x2="${w - pad}" y2="${chartH + pad}" stroke="#e5e7eb"/>
    <line x1="${pad}" y1="${pad + legendH}" x2="${pad}" y2="${chartH + pad}" stroke="#e5e7eb"/>
    ${yLabels}${xLabels}${lines}
  </svg>`;
}

// 分组柱状图生成器（多课程并列）
function renderGroupedBarChart(groups, opts = {}) {
  const w = opts.width || 800, h = opts.height || 280, pad = 30, legendH = 24;
  const colors = ['#3b82f6', '#8b5cf6', '#10b981', '#f59e0b', '#ef4444', '#06b6d4', '#ec4899'];
  if (groups.length === 0) return '<p class="text-gray text-sm">暂无数据</p>';
  const allNames = [];
  var _nameSet = {};
  [].concat.apply([], groups.map(function(g) { return g.values.map(function(v) { return v.name; }); })).forEach(function(n) {
    if (!_nameSet[n]) { _nameSet[n] = true; allNames.push(n); }
  });
  const nameColor = {};
  allNames.forEach((n, i) => nameColor[n] = colors[i % colors.length]);
  const maxVal = Math.max.apply(null, [].concat.apply([], groups.map(function(g) { return g.values.map(function(v) { return v.value; }); })).concat([1]));
  const chartH = h - pad - legendH;
  const groupW = (w - pad * 2) / groups.length;
  const barW = allNames.length > 0 ? Math.min(groupW / (allNames.length + 0.5), 24) : 0;
  const bars = groups.map((g, i) => {
    const gx = pad + i * groupW;
    return g.values.map((v, j) => {
      const x = gx + (groupW - barW * allNames.length) / 2 + j * barW;
      const bh = (v.value / maxVal) * chartH;
      const y = chartH + pad - bh;
      const tipText = g.tooltip ? g.tooltip(v) : `${g.label} ${v.name}: ${formatHours(v.value)} 课时`;
      return `<rect x="${x}" y="${y}" width="${barW}" height="${bh}" fill="${nameColor[v.name]}" rx="2"><title>${escapeHtml(tipText)}</title></rect>${v.value > 0 ? `<text x="${x + barW/2}" y="${y - 3}" text-anchor="middle" font-size="8" fill="#666">${formatHours(v.value)}</text>` : ''}`;
    }).join('');
  }).join('');
  const labelStep = Math.ceil(groups.length / 12);
  const xLabels = groups.map((g, i) => i % labelStep === 0 ? `<text x="${pad + i * groupW + groupW/2}" y="${h - pad + 14}" text-anchor="middle" font-size="9" fill="#999">${escapeHtml(g.label)}</text>` : '').join('');
  const legend = allNames.map((n, i) => `<g transform="translate(${pad + i * 90}, 8)"><rect width="10" height="10" fill="${nameColor[n]}" rx="2"/><text x="14" y="9" font-size="11" fill="#666">${escapeHtml(n)}</text></g>`).join('');
  return `<svg viewBox="0 0 ${w} ${h}" style="width:100%;height:auto;">
    ${legend}
    <line x1="${pad}" y1="${chartH + pad}" x2="${w - pad}" y2="${chartH + pad}" stroke="#e5e7eb"/>
    <line x1="${pad}" y1="${pad + legendH}" x2="${pad}" y2="${chartH + pad}" stroke="#e5e7eb"/>
    ${bars}${xLabels}
  </svg>`;
}

// ==================== 页面: 仪表盘 ====================
async function renderDashboard() {
  const body = $('#pageBody');
  body.innerHTML = renderLoading();
  try {
    const [stats, chartsData, notifCount] = await Promise.all([API.dashboard.stats(), API.dashboard.charts(), API.notification.unreadCount()]);
    const courseStats = stats.course_stats || [];
    const dailyByCourse = stats.daily_by_course || [];
    const charts = chartsData.charts || [];

    body.innerHTML = `
      <div class="stat-cards">
        <div class="stat-card stat-blue">
          <div class="stat-icon">${ICONS.students}</div>
          <div class="stat-info"><div class="stat-value">${stats.total_students || 0}</div><div class="stat-label">学生总数</div></div>
        </div>
        <div class="stat-card stat-purple">
          <div class="stat-icon">${ICONS.book}</div>
          <div class="stat-info"><div class="stat-value">${stats.total_courses || 0}</div><div class="stat-label">课程总数</div></div>
        </div>
        <div class="stat-card stat-orange">
          <div class="stat-icon">${ICONS.clock}</div>
          <div class="stat-info"><div class="stat-value">${formatHours(stats.total_hours_consumed || 0)}/${formatHours(stats.total_hours || 0)}</div><div class="stat-label">总消耗课时/总课时</div></div>
        </div>
      </div>
      <div class="card mb-6">
        <div class="card-header">快捷操作</div>
        <div class="card-body flex gap-3" style="flex-wrap:wrap;">
          <button class="btn btn-success" onclick="navigate('/hour-records');setTimeout(()=>openBatchHourModal(),200)">${ICONS.clock}记录课时</button>
          <button class="btn btn-primary" onclick="navigate('/students');setTimeout(()=>openStudentModal(),200)">${ICONS.plus}添加学生</button>
          <button class="btn btn-primary" onclick="navigate('/courses');setTimeout(()=>openCourseModal(0),200)">${ICONS.plus}添加课程</button>
          <button class="btn btn-chart" onclick="navigate('/calendar')">${ICONS.calendar}课程日历</button>
          <button class="btn btn-secondary" onclick="navigate('/notifications')">${ICONS.bell}通知中心${(notifCount && notifCount.count > 0) ? `<span class="btn-badge">${notifCount.count > 9 ? '9+' : notifCount.count}</span>` : ''}</button>
          <button class="btn btn-secondary" onclick="navigate('/logs')">${ICONS.journal}操作日志</button>
        </div>
      </div>
      <div class="card mb-6">
        <div class="card-header">各课程课时统计</div>
        <div class="card-body">
          <table class="data-table">
            <thead>
              <tr>
                <th>课程名称</th>
                <th>总消耗课时</th>
                <th>本周消耗</th>
                <th>本月消耗</th>
                <th>近半年消耗</th>
                <th>近一年消耗</th>
              </tr>
            </thead>
            <tbody>
              ${courseStats.map((cs, i) => `
              <tr>
                <td>${escapeHtml(cs.course_name)}</td>
                <td>${formatHours(cs.consumed_hours || 0)}</td>
                <td>${formatHours(cs.this_week_hours || 0)}</td>
                <td>${formatHours(cs.this_month_hours || 0)}</td>
                <td>${formatHours(cs.last_six_months_hours || 0)}</td>
                <td>${formatHours(cs.last_year_hours || 0)}</td>
              </tr>`).join('')}
              ${courseStats.length > 0 ? `
              <tr style="font-weight:bold;border-top:2px solid #e5e7eb;">
                <td>总计</td>
                <td>${formatHours(courseStats.reduce((sum, cs) => sum + (cs.consumed_hours || 0), 0))}</td>
                <td>${formatHours(courseStats.reduce((sum, cs) => sum + (cs.this_week_hours || 0), 0))}</td>
                <td>${formatHours(courseStats.reduce((sum, cs) => sum + (cs.this_month_hours || 0), 0))}</td>
                <td>${formatHours(courseStats.reduce((sum, cs) => sum + (cs.last_six_months_hours || 0), 0))}</td>
                <td>${formatHours(courseStats.reduce((sum, cs) => sum + (cs.last_year_hours || 0), 0))}</td>
              </tr>` : ''}
            </tbody>
          </table>
        </div>
      </div>
      <div class="card mb-6">
        <div class="card-header">近30天各课程每日消耗课时</div>
        <div class="card-body">
          ${dailyByCourse.length > 0 ? renderLineChart(dailyByCourse, { width: 900, height: 300 }) : '<p class="text-gray text-sm">暂无数据</p>'}
        </div>
      </div>
      <div class="card mb-6">
        <div class="card-header">各课程近半年消耗趋势（按周统计）</div>
        <div class="card-body">
          ${charts.length > 0 ? charts.map(c => `
            <div style="margin-bottom:20px;">
              <div class="text-sm" style="font-weight:600;margin-bottom:4px;">${escapeHtml(c.course_name)}</div>
              ${renderBarChart((c.weeks || []).filter(w => w.week).map(w => ({ label: w.week.replace(/^\d{4}-W/, 'W'), value: w.hours, week: w.week })), { color: '#10b981', width: 800, height: 150, tooltip: (d) => `${weekRangeLabel(d.week)} ${c.course_name}: ${formatHours(d.value || 0)} 课时` })}
            </div>
          `).join('') : '<p class="text-gray text-sm">暂无数据</p>'}
        </div>
      </div>
      <div class="card">
        <div class="card-header">各课程近一年消耗趋势（按月统计）</div>
        <div class="card-body">
          ${charts.length > 0 ? charts.map(c => `
            <div style="margin-bottom:20px;">
              <div class="text-sm" style="font-weight:600;margin-bottom:4px;">${escapeHtml(c.course_name)}</div>
              ${renderBarChart((c.months || []).filter(m => m.month).map(m => ({ label: m.month.replace(/^\d{4}-/, ''), value: m.hours, month: m.month })), { color: '#3b82f6', width: 800, height: 150, tooltip: (d) => `${d.month} ${c.course_name}: ${formatHours(d.value || 0)} 课时` })}
            </div>
          `).join('') : '<p class="text-gray text-sm">暂无数据</p>'}
        </div>
      </div>`;
  } catch (err) {
    body.innerHTML = `<div class="empty-state"><p>加载失败: ${escapeHtml(getErrorMessage(err))}</p></div>`;
  }
}

// ==================== 页面: 学生管理 ====================
async function renderStudents() {
  managingCourse = null;
  const body = $('#pageBody');
  body.innerHTML = `
    <div class="flex justify-between mb-6">
      <div></div>
      <button class="btn btn-primary" onclick="openStudentModal()">${ICONS.plus}添加学生</button>
    </div>
    <div class="filter-bar">
      <div class="form-group"><label class="form-label">搜索姓名</label><input class="form-input" id="searchName" placeholder="输入姓名..."></div>
      <div class="form-group"><label class="form-label">搜索学号</label><input class="form-input" id="searchStudentID" placeholder="输入学号..."></div>
      <div class="form-group"><label class="form-label">状态</label>${createSingleSelectHTML('searchStudentStatus', '全部')}</div>
      <button class="btn btn-secondary" onclick="searchStudents()">${ICONS.search}搜索</button>
    </div>
    <div class="card"><div id="studentTable">${renderLoading()}</div></div>`;
  populateSingleSelect('searchStudentStatus', [
    { value: '', label: '全部' },
    { value: '0', label: '在读' },
    { value: '1', label: '已退学' }
  ], '');
  await loadStudents(1);
}

let allCourses = [];
async function loadCourses() {
  try { const result = await API.course.list({ page: 1, page_size: 100 }); allCourses = result.data || []; } catch { allCourses = []; }
}

let studentPage = 1;
let studentTotalPages = 1;
let studentTotal = 0;

// 生成学生表格行的 HTML（loadStudents 和 searchStudents 共用）
function renderStudentRow(s, batchKey, withProgress) {
  const pct = s.total_hours > 0 ? Math.min(100, (s.completed_hours / s.total_hours) * 100) : 0;
  const remaining = s.total_hours - s.completed_hours;
  const cls = remaining < 5 ? 'danger' : remaining < 10 ? 'warning' : 'success';
  const statusBadge = s.is_dropped ? '<span class="badge badge-danger">已退学</span>' : '<span class="badge badge-success">在读</span>';
  var progressCell = withProgress ? '<td style="min-width:100px;"><div class="progress-bar"><div class="progress-bar-fill ' + cls + '" style="width:' + pct + '%"></div></div></td>' : '';
  return '<tr' + (s.is_dropped && !withProgress ? ' style="opacity:0.6;"' : '') + '>' +
    renderRowCheckbox(batchKey, s.id) +
    '<td>' + escapeHtml(s.name) + '</td>' +
    '<td>' + escapeHtml(s.student_id) + '</td>' +
    '<td>' + formatHours(s.completed_hours) + '/' + formatHours(s.total_hours) + '</td>' +
    '<td>' + formatHours(remaining) + '</td>' +
    '<td>' + statusBadge + '</td>' +
    progressCell +
    '<td><div class="flex gap-2">' +
      '<button class="btn-icon blue" onclick="openStudentModal(' + s.id + ')" title="编辑">' + ICONS.edit + '</button>' +
      '<button class="btn-icon red" onclick="deleteStudent(' + s.id + ', \'' + escapeHtml(s.name) + '\')" title="删除">' + ICONS.trash + '</button>' +
      '<button class="btn-icon purple" onclick="openRechargeModal(' + s.id + ', \'' + escapeHtml(s.name) + '\', ' + remaining + ')" title="充值">' + ICONS.wallet + '</button>' +
      '<button class="btn-icon green" onclick="viewStudentRecharges(' + s.id + ', \'' + escapeHtml(s.name) + '\')" title="充值记录">' + ICONS.journal + '</button>' +
    '</div></td>' +
  '</tr>';
}

async function loadStudents(page = 1) {
  studentPage = page;
  registerBatchConfig('students', API.student.batchDelete, loadStudents, '学生');
  const container = $('#studentTable');
  try {
    const result = await API.student.list({ page, page_size: pageSizes.loadStudents });
    const students = result.data || [];
    studentTotal = result.total || 0;
    studentTotalPages = result.total_pages || 1;

    if (!students || students.length === 0) {
      container.innerHTML = renderEmptyState(ICONS.students, '暂无学生数据');
      return;
    }
    var rowsHtml = students.map(function(s) { return renderStudentRow(s, 'students', true); }).join('');
    container.innerHTML = renderDataTable({
      batchKey: 'students',
      headers: ['姓名', '学号', '课时消耗', '剩余', '状态', '进度', '操作'],
      rowsHtml: rowsHtml,
      page: studentPage,
      totalPages: studentTotalPages,
      callbackName: 'loadStudents',
      pageSize: pageSizes.loadStudents,
      total: studentTotal,
    });
  } catch (err) {
    container.innerHTML = renderErrorState(getErrorMessage(err));
  }
}

async function searchStudents(page = 1) {
  studentPage = page;
  registerBatchConfig('searchStudents', API.student.batchDelete, searchStudents, '学生');
  const name = $('#searchName').value.trim();
  const sid = $('#searchStudentID').value.trim();
  const status = getSingleSelectValue('searchStudentStatus');
  const container = $('#studentTable');
  container.innerHTML = renderLoading();
  try {
    const params = { name, student_id: sid, page, page_size: pageSizes.searchStudents };
    if (status === '0' || status === '1') {
      params.is_dropped = status === '1';
    }
    const result = await API.student.search(params);
    const students = result.data || [];
    studentTotal = result.total || 0;
    studentTotalPages = result.total_pages || 1;

    if (!students || students.length === 0) {
      container.innerHTML = renderEmptyState('', '未找到匹配的学生');
      return;
    }
    var rowsHtml = students.map(function(s) { return renderStudentRow(s, 'searchStudents', false); }).join('');
    container.innerHTML = renderDataTable({
      batchKey: 'searchStudents',
      headers: ['姓名', '学号', '课时消耗', '剩余', '状态', '操作'],
      rowsHtml: rowsHtml,
      page: studentPage,
      totalPages: studentTotalPages,
      callbackName: 'searchStudents',
      pageSize: pageSizes.searchStudents,
      total: studentTotal,
    });
  } catch (err) {
    container.innerHTML = '<div class="empty-state"><p>搜索失败: ' + escapeHtml(getErrorMessage(err)) + '</p></div>';
  }
}

async function openStudentModal(id) {
  await loadCourses();
  let student = null, selectedCourses = [];
  if (id) {
    student = await API.student.get(id);
    try { const sc = await API.course.getByStudent(id); selectedCourses = (sc || []).map(c => c.id); } catch {}
  }
  const form = student || { name: '', student_id: '', contact: '', total_hours: 0, is_dropped: false };
  showModal(id ? '编辑学生' : '添加学生', `
    <div id="studentFormError"></div>
    <div class="form-group"><label class="form-label">姓名 *</label><input class="form-input" id="f_name" value="${escapeHtml(form.name)}"></div>
    <div class="form-group"><label class="form-label">学号</label>
      <div class="flex gap-2">
        <input class="form-input" id="f_student_id" value="${escapeHtml(form.student_id)}" placeholder="留空自动生成" ${id ? 'disabled' : ''}>
        ${!id ? `<button class="btn btn-secondary btn-sm" onclick="generateStudentID()">${ICONS.refresh}</button>` : ''}
      </div>
    </div>
    <div class="form-group"><label class="form-label">联系方式</label><input class="form-input" id="f_contact" value="${escapeHtml(form.contact)}"></div>
    <div class="form-group"><label class="form-label">总课时</label><input class="form-input" type="number" id="f_total_hours" value="${form.total_hours}" min="0" step="0.5"></div>
    <div class="form-group"><label class="form-label">关联课程</label>
      ${createMultiSelectHTML('studentCourseSelect', '选择课程')}
    </div>
    ${id ? `<div class="form-group"><label class="form-label"><input type="checkbox" id="f_is_dropped" ${form.is_dropped ? 'checked' : ''}> 退学</label></div>` : ''}
    ${renderFormFooter('saveStudent(' + (id || 0) + ')', id ? '保存' : '创建')}
  `);

  // 填充课程多选下拉框
  if (allCourses.length > 0) {
    populateMultiSelect('studentCourseSelect', allCourses.map(c => ({ value: c.id, label: c.name })), selectedCourses);
  }
}

async function generateStudentID() {
  try { const res = await API.student.generateID(); $('#f_student_id').value = res.student_id; } catch (err) { showToast('生成失败: ' + getErrorMessage(err), 'error'); }
}

async function saveStudent(id) {
  const name = $('#f_name').value.trim();
  if (!name) { $('#studentFormError').innerHTML = '<div class="form-error">请输入学生姓名</div>'; return; }
  const data = {
    name,
    student_id: $('#f_student_id').value.trim(),
    contact: $('#f_contact').value.trim(),
    total_hours: parseFloat($('#f_total_hours').value) || 0,
    is_dropped: $('#f_is_dropped') ? $('#f_is_dropped').checked : false,
  };
  // 获取选中的课程ID
  const selectedCourseIds = getMultiSelectValues('studentCourseSelect').map(Number);
  try {
    if (id) {
      await API.student.update(id, data);
      const currentCourses = await API.course.getByStudent(id);
      const currentIds = (currentCourses || []).map(c => c.id);
      for (const cid of currentIds) { if (!selectedCourseIds.includes(cid)) await API.course.unenroll({ student_id: id, course_id: cid }); }
      for (const cid of selectedCourseIds) { if (!currentIds.includes(cid)) await API.course.enroll({ student_id: id, course_id: cid }); }
    } else {
      const created = await API.student.create(data);
      for (const cid of selectedCourseIds) await API.course.enroll({ student_id: created.id, course_id: cid });
    }
    closeModal();
    showToast('保存成功', 'success');
    if (managingCourse) await viewCourseStudents(managingCourse.id, managingCourse.name);
    else await loadStudents(1);
  } catch (err) {
    $('#studentFormError').innerHTML = `<div class="form-error">${escapeHtml(getErrorMessage(err))}</div>`;
  }
}

function deleteStudent(id, name) {
  confirmModal('确认删除', `确定要删除学生 "${name}" 吗？此操作无法撤销。`, async () => {
    try { await API.student.delete(id); showToast('删除成功', 'success'); await loadStudents(1); }
    catch (err) { showToast('删除失败: ' + getErrorMessage(err), 'error'); }
  }, '删除');
}

function openRechargeModal(id, name, remaining) {
  showModal(`给 "${name}" 充值课时`, `
    <div class="card-body" style="background:var(--gray-50);border-radius:8px;margin-bottom:16px;">
      <p class="text-sm text-gray">当前剩余课时</p>
      <p style="font-size:20px;font-weight:600;">${formatHours(remaining)} 课时</p>
    </div>
    <div class="form-group"><label class="form-label">充值课时数 *</label><input class="form-input" type="number" id="rechargeAmount" min="0.5" step="0.5" placeholder="请输入充值课时数"></div>
    <div class="form-group"><label class="form-label">备注</label><input class="form-input" id="rechargeDesc" value="给 ${name} 充值课时"></div>
    ${renderFormFooter('doRecharge(' + id + ')', '确认充值').replace('btn-primary', 'btn-primary" style="background:var(--purple);"')}
  `);
}

async function doRecharge(id) {
  const hours = parseFloat($('#rechargeAmount').value);
  if (!hours || hours <= 0) { showToast('请输入有效的课时数', 'error'); return; }
  const desc = $('#rechargeDesc').value.trim();
  try {
    await API.hourRecharge.create({ student_id: id, hours, description: desc });
    closeModal();
    showToast('充值成功', 'success');
    // 根据当前页面决定刷新哪个列表
    var currentPage = location.hash.slice(1) || '/';
    if (managingCourse) await viewCourseStudents(managingCourse.id, managingCourse.name);
    else if (currentPage === '/recharges') await loadRecharges(1);
    else await loadStudents(1);
  } catch (err) { showToast('充值失败: ' + getErrorMessage(err), 'error'); }
}

// 学生管理 -> 查看某学生的充值记录入口
function viewStudentRecharges(studentId, studentName) {
  rechargeFilterState.studentId = studentId;
  rechargeFilterState.studentName = studentName;
  navigate('/recharges');
}

// ==================== 页面: 课程管理 ====================
const courseSchedulesMap = {};
let managingCourse = null;

async function renderCourses() {
  const body = $('#pageBody');
  body.innerHTML = `
    <div class="flex justify-between mb-6">
      <div></div>
      <button class="btn btn-primary" onclick="openCourseModal()">${ICONS.plus}添加课程</button>
    </div>
    <div class="card">
      <div class="card-header">课程列表</div>
      <div id="courseTable">${renderLoading()}</div>
    </div>`;
  await loadCoursesTable(1);
}

let coursePage = 1;
let courseTotalPages = 1;
let courseTotal = 0;

async function loadCoursesTable(page = 1) {
  coursePage = page;
  registerBatchConfig('courses', API.course.batchDelete, loadCoursesTable, '课程');
  const container = $('#courseTable');
  try {
    const result = await API.course.list({ page, page_size: pageSizes.loadCoursesTable });
    const courses = result.data || [];
    courseTotal = result.total || 0;
    courseTotalPages = result.total_pages || 1;

    if (!courses || courses.length === 0) {
      container.innerHTML = renderEmptyState(ICONS.book, '暂无课程数据');
      return;
    }
    for (const c of courses) {
      if (!courseSchedulesMap[c.id]) {
        try { courseSchedulesMap[c.id] = await API.schedule.getByCourse(c.id) || []; } catch { courseSchedulesMap[c.id] = []; }
      }
    }
    var rowsHtml = courses.map(function(c) {
      const scheds = courseSchedulesMap[c.id] || [];
      const schedStr = scheds.length > 0 ? scheds.map(s => `${weekdayLabel(s.day_of_week)} ${s.start_time}-${s.end_time}`).join(', ') : '-';
      return '<tr>' + renderRowCheckbox('courses', c.id) +
        '<td>' + escapeHtml(c.name) + '</td>' +
        '<td>' + (c.student_count || 0) + '</td>' +
        '<td>' + formatHours(c.remaining_hours || 0) + '</td>' +
        '<td class="text-sm text-gray">' + escapeHtml(schedStr) + '</td>' +
        '<td><div class="flex gap-2">' +
          '<button class="btn-icon blue" onclick="openCourseModal(' + c.id + ')" title="编辑">' + ICONS.edit + '</button>' +
          '<button class="btn-icon red" onclick="deleteCourse(' + c.id + ', \'' + escapeHtml(c.name) + '\')" title="删除">' + ICONS.trash + '</button>' +
          '<button class="btn-icon green" onclick="viewCourseStudents(' + c.id + ', \'' + escapeHtml(c.name) + '\')" title="管理学生">' + ICONS.users + '</button>' +
          '<button class="btn-icon purple" onclick="openAddHoursModal(' + c.id + ', \'' + escapeHtml(c.name) + '\')" title="赠送课时">' + ICONS.gift + '</button>' +
        '</div></td>' +
      '</tr>';
    }).join('');
    container.innerHTML = renderDataTable({
      batchKey: 'courses',
      headers: ['课程名称', '学生数', '剩余课时', '上课时间', '操作'],
      rowsHtml: rowsHtml,
      page: coursePage,
      totalPages: courseTotalPages,
      callbackName: 'loadCoursesTable',
      pageSize: pageSizes.loadCoursesTable,
      total: courseTotal,
    });
  } catch (err) {
    container.innerHTML = renderErrorState(getErrorMessage(err));
  }
}

async function openCourseModal(id) {
  let course = null, existingSchedules = [];
  if (id) {
    course = await API.course.get(id);
    try { existingSchedules = await API.schedule.getByCourse(id) || []; } catch {}
  }
  const form = course || { name: '', description: '', threshold: 2 };
  showModal(id ? '编辑课程' : '添加课程', `
    <div id="courseFormError"></div>
    <div class="form-group"><label class="form-label">课程名称 *</label><input class="form-input" id="c_name" value="${escapeHtml(form.name)}"></div>
    <div class="form-group"><label class="form-label">课时预警阈值</label><input class="form-input" type="number" id="c_threshold" value="${form.threshold || 2}" min="0" step="0.5"><p class="text-sm text-gray mt-4">学生剩余课时低于此值时触发预警，默认 2</p></div>
    <div class="form-group"><label class="form-label">描述</label><textarea class="form-textarea" id="c_description">${escapeHtml(form.description)}</textarea></div>
    <div class="form-group">
      <label class="form-label"><span class="icon-inline">${ICONS.clock}</span> 周期性上课时间</label>
      <div id="existingSchedules" style="margin-bottom:8px;">
        ${existingSchedules.map(s => `<div class="schedule-entry existing"><span>${weekdayLabel(s.day_of_week)} ${s.period} ${s.start_time}-${s.end_time} (${s.hours_consumed}课时)</span><button class="del-btn" onclick="deleteSchedule(${s.id}, ${id})">${ICONS.close}</button></div>`).join('')}
      </div>
      <div id="newSchedules"></div>
      <div class="schedule-input-row">
        <div class="form-group"><label class="form-label text-sm">星期</label>${createSingleSelectHTML('newDay', '请选择')}</div>
        <div class="form-group"><label class="form-label text-sm">开始</label><input class="form-input" type="time" id="newStart" value="19:00"></div>
        <div class="form-group"><label class="form-label text-sm">结束</label><input class="form-input" type="time" id="newEnd" value="20:00"></div>
        <button class="btn btn-primary btn-sm" onclick="addScheduleEntry()">${ICONS.plus}添加</button>
      </div>
    </div>
    ${renderFormFooter('saveCourse(' + (id || 0) + ')', id ? '保存' : '创建')}
  `, 'lg');
  window._newSchedules = [];
  populateSingleSelect('newDay', WEEKDAYS.map(w => ({ value: w.v, label: w.l })), '1');
}

function addScheduleEntry() {
  const day = parseInt(getSingleSelectValue('newDay'));
  const start = $('#newStart').value;
  const end = $('#newEnd').value;
  const hours = calcHours(start, end);
  if (hours <= 0) { showToast('结束时间需晚于开始时间', 'error'); return; }
  const entry = { day_of_week: day, start_time: start, end_time: end, period: getPeriod(start), hours_consumed: hours };
  window._newSchedules = window._newSchedules || [];
  window._newSchedules.push(entry);
  renderNewSchedules();
}

function renderNewSchedules() {
  $('#newSchedules').innerHTML = (window._newSchedules || []).map((e, i) =>
    `<div class="schedule-entry new"><span>${weekdayLabel(e.day_of_week)} ${e.period} ${e.start_time}-${e.end_time} (${e.hours_consumed}课时)</span><button class="del-btn" onclick="removeScheduleEntry(${i})">${ICONS.close}</button></div>`
  ).join('');
}

function removeScheduleEntry(idx) {
  window._newSchedules.splice(idx, 1);
  renderNewSchedules();
}

async function deleteSchedule(scheduleId, courseId) {
  try {
    await API.schedule.delete(scheduleId);
    const schedules = await API.schedule.getByCourse(courseId);
    courseSchedulesMap[courseId] = schedules || [];
    closeModal();
    await openCourseModal(courseId);
    showToast('已删除排课', 'success');
  } catch (err) { showToast('删除失败: ' + getErrorMessage(err), 'error'); }
}

async function saveCourse(id) {
  const name = $('#c_name').value.trim();
  if (!name) { $('#courseFormError').innerHTML = '<div class="form-error">请输入课程名称</div>'; return; }
  const data = {
    name,
    description: $('#c_description').value.trim(),
    threshold: parseFloat($('#c_threshold').value) || 2,
  };
  try {
    let courseId;
    if (id) { await API.course.update(id, data); courseId = id; }
    else { const created = await API.course.create(data); courseId = created.id; }
    for (const e of (window._newSchedules || [])) {
      await API.schedule.create({ course_id: courseId, ...e });
    }
    courseSchedulesMap[courseId] = await API.schedule.getByCourse(courseId) || [];
    closeModal();
    showToast('保存成功', 'success');
    await loadCoursesTable(1);
  } catch (err) {
    $('#courseFormError').innerHTML = `<div class="form-error">${escapeHtml(getErrorMessage(err))}</div>`;
  }
}

function deleteCourse(id, name) {
  confirmModal('确认删除', `确定要删除课程 "${name}" 吗？此操作将同时删除相关的学生选课记录和课时安排。`, async () => {
    try { await API.course.delete(id); showToast('删除成功', 'success'); await loadCoursesTable(1); }
    catch (err) { showToast('删除失败: ' + getErrorMessage(err), 'error'); }
  }, '删除');
}

async function viewCourseStudents(id, name) {
  managingCourse = { id, name };
  showModal(`管理学生 - ${name}`, `<div id="courseStudentModalBody">${renderLoading()}</div>`, 'lg');
  try {
    const students = await API.course.getStudents(id);
    const body = $('#courseStudentModalBody');
    if (!students || students.length === 0) {
      body.innerHTML = `
        <div class="empty-state"><p>该课程暂无学生</p></div>
        <div class="flex justify-between mt-4">
          <button class="btn btn-secondary" onclick="closeModal()">关闭</button>
          <button class="btn btn-primary btn-sm" onclick="openEnrollModal(${id})">${ICONS.plus}添加学生</button>
        </div>`;
      return;
    }
    body.innerHTML = `
      <div class="flex justify-between mb-4">
        <span class="text-sm text-gray">共 ${students.length} 名学生</span>
        <button class="btn btn-primary btn-sm" onclick="openEnrollModal(${id})">${ICONS.plus}添加学生</button>
      </div>
      <table class="data-table">
        <thead><tr><th>姓名</th><th>学号</th><th>联系方式</th><th>剩余课时</th><th>操作</th></tr></thead>
        <tbody>${students.map(s => `<tr>
          <td>${escapeHtml(s.name)}</td>
          <td>${escapeHtml(s.student_id)}</td>
          <td>${escapeHtml(s.contact)}</td>
          <td>${formatHours(s.total_hours - s.completed_hours)}</td>
          <td><div class="flex gap-2">
            <button class="btn-icon blue" onclick="openStudentModal(${s.id})" title="编辑">${ICONS.edit}</button>
            <button class="btn-icon red" onclick="unenrollStudent(${id}, ${s.id}, '${escapeHtml(s.name)}')" title="删除选课">${ICONS.trash}</button>
          </div></td>
        </tr>`).join('')}</tbody>
      </table>
      <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`;
  } catch (err) {
    $('#courseStudentModalBody').innerHTML = renderErrorState(getErrorMessage(err));
  }
}

async function unenrollStudent(courseId, studentId, studentName) {
  confirmModal('确认删除', `确定要将 "${studentName}" 从该课程中移除吗？`, async () => {
    try {
      await API.course.unenroll({ student_id: studentId, course_id: courseId });
      showToast('删除成功', 'success');
      await viewCourseStudents(managingCourse.id, managingCourse.name);
    } catch (err) { showToast('删除失败: ' + getErrorMessage(err), 'error'); }
  }, '删除');
}

function openAddHoursModal(id, name) {
  showModal(`为 ${name} 赠送课时`, `
    <div class="form-group"><label class="form-label">赠送课时数</label><input class="form-input" type="number" id="hoursToAdd" min="0.5" step="0.5" value="1"></div>
    <p class="text-sm text-gray">此操作将为课程中的所有学生增加课时</p>
    ${renderFormFooter('doAddHours(' + id + ')', '确认赠送').replace('btn-primary', 'btn-success')}
  `);
}

async function doAddHours(id) {
  const hours = parseFloat($('#hoursToAdd').value);
  if (!hours || hours <= 0) { showToast('请输入有效的课时数', 'error'); return; }
  try {
    await API.course.addHours(id, hours);
    closeModal();
    showToast('赠送成功', 'success');
    await loadCoursesTable(1);
    if (managingCourse) await viewCourseStudents(managingCourse.id, managingCourse.name);
  } catch (err) { showToast('充值失败: ' + getErrorMessage(err), 'error'); }
}

async function openEnrollModal(courseId) {
  const studentsResult = await API.student.list({ page: 1, page_size: 1000 });
  const students = studentsResult.data || [];
  const courseStudents = await API.course.getStudents(courseId);
  const enrolledIds = (courseStudents || []).map(s => s.id);
  showModal('添加学生到课程', `
    <div style="max-height:400px;overflow-y:auto;">
      ${students.map(s => {
        const enrolled = enrolledIds.includes(s.id);
        return `<div style="padding:8px;border:1px solid var(--border);border-radius:8px;margin-bottom:4px;display:flex;justify-content:space-between;align-items:center;${enrolled ? 'opacity:0.5;' : ''}">
          <span>${escapeHtml(s.name)} <span class="text-gray">${escapeHtml(s.student_id)}</span></span>
          ${enrolled ? '<span class="text-gray">已加入</span>' : `<button class="btn btn-primary btn-sm" onclick="doEnroll(${courseId}, ${s.id})">添加</button>`}
        </div>`;
      }).join('')}
    </div>
    <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`);
}

async function doEnroll(courseId, studentId) {
  try {
    await API.course.enroll({ student_id: studentId, course_id: courseId });
    showToast('添加成功', 'success');
    closeModal();
    if (managingCourse) await viewCourseStudents(managingCourse.id, managingCourse.name);
  } catch (err) { showToast('添加失败: ' + getErrorMessage(err), 'error'); }
}

// ==================== 页面: 充值记录 ====================
const rechargeFilterState = { studentId: null, studentName: null };

async function renderRecharges() {
  const body = $('#pageBody');
  const presetStudentId = rechargeFilterState.studentId;
  const presetStudentName = rechargeFilterState.studentName;
  // 渲染一次后清除预设，避免下次进入仍带筛选
  rechargeFilterState.studentId = null;
  rechargeFilterState.studentName = null;

  body.innerHTML = `
    <div class="flex justify-between mb-6">
      <div></div>
      <button class="btn btn-primary" onclick="openRechargeSelectStudent()">${ICONS.wallet}充值</button>
    </div>
    <div class="filter-bar">
      <div class="form-group"><label class="form-label">学生姓名</label><input class="form-input" id="rechargeSearchName" placeholder="输入姓名..." value="${escapeHtml(presetStudentName || '')}"></div>
      <div class="form-group"><label class="form-label">状态</label>${createSingleSelectHTML('rechargeStudentStatus', '全部')}</div>
      <div class="form-group"><label class="form-label">开始日期</label><input class="form-input" type="date" id="rechargeStartDate"></div>
      <div class="form-group"><label class="form-label">结束日期</label><input class="form-input" type="date" id="rechargeEndDate"></div>
      <button class="btn btn-secondary" onclick="loadRecharges()">${ICONS.search}筛选</button>
    </div>
    <div class="card">
      <div class="card-header">充值记录列表</div>
      <div id="rechargeTable">${renderLoading()}</div>
    </div>`;

  // 初始化状态筛选
  populateSingleSelect('rechargeStudentStatus', [
    { value: '', label: '全部' },
    { value: '0', label: '在读' },
    { value: '1', label: '已退学' }
  ], '');

  // 若由学生管理跳转而来，立即按 student_id 加载
  if (presetStudentId) {
    await loadRecharges(1, { student_id: presetStudentId });
  } else {
    await loadRecharges(1);
  }
}

let rechargePage = 1;
let rechargeTotalPages = 1;
let rechargeTotal = 0;

async function loadRecharges(page = 1, extra = {}) {
  rechargePage = page;
  registerBatchConfig('recharges', API.hourRecharge.batchDelete, loadRecharges, '充值记录');
  const container = $('#rechargeTable');
  if (!container) return;
  container.innerHTML = renderLoading();
  const name = ($('#rechargeSearchName') || {}).value ? ($('#rechargeSearchName').value.trim()) : '';
  const startDate = ($('#rechargeStartDate') || {}).value || '';
  const endDate = ($('#rechargeEndDate') || {}).value || '';

  // 获取状态筛选
  const statusEl = document.querySelector('input[name="rechargeStudentStatus"]:checked');
  const status = statusEl ? statusEl.value : '';

  try {
    // 先获取全部学生用于按姓名解析 student_id 和状态筛选
    let studentId = extra.student_id || null;
    let filteredStudentIds = [];
    const studentsResult = await API.student.list({ page: 1, page_size: 1000 });
    const students = studentsResult.data || [];

    // 根据状态筛选学生
    if (status !== '') {
      filteredStudentIds = students.filter(s => {
        const isDropped = s.is_dropped ? '1' : '0';
        return isDropped === status;
      }).map(s => s.id);
    }

    if (!studentId && name) {
      const matched = students.find(s => s.name === name);
      if (matched) studentId = matched.id;
      else {
        container.innerHTML = renderEmptyState('', '未找到匹配的学生');
        return;
      }
    }

    const params = { page, page_size: pageSizes.loadRecharges };
    if (studentId) params.student_id = studentId;
    if (startDate) params.start_date = startDate;
    if (endDate) params.end_date = endDate;
    const result = await API.hourRecharge.list(params);
    let records = result.data || [];
    rechargeTotal = result.total || 0;
    rechargeTotalPages = result.total_pages || 1;

    // 根据状态筛选充值记录
    if (status !== '' && filteredStudentIds.length > 0) {
      records = records.filter(r => filteredStudentIds.includes(r.student_id));
    }
    if (!records || records.length === 0) {
      container.innerHTML = renderEmptyState(ICONS.wallet, '暂无充值记录');
      return;
    }
    const totalHours = records.reduce((sum, r) => sum + (r.hours || 0), 0);
    var rowsHtml = records.map(function(r) {
      return '<tr>' + renderRowCheckbox('recharges', r.id) +
        '<td>' + escapeHtml(r.student_name) + '</td>' +
        '<td>' + escapeHtml(r.student_no) + '</td>' +
        '<td><span class="badge blue">' + formatHours(r.hours || 0) + '</span></td>' +
        '<td>' + escapeHtml(r.recharge_date) + '</td>' +
        '<td class="text-sm text-gray">' + escapeHtml(r.description || '-') + '</td>' +
        '<td class="text-sm text-gray">' + formatDate(r.created_at) + '</td>' +
        '<td><div class="flex gap-2">' +
          '<button class="btn-icon blue" onclick="openEditRechargeModal(' + r.id + ')" title="编辑">' + ICONS.edit + '</button>' +
          '<button class="btn-icon red" onclick="deleteRecharge(' + r.id + ')" title="删除">' + ICONS.trash + '</button>' +
        '</div></td>' +
      '</tr>';
    }).join('');
    container.innerHTML =
      '<div class="card">' +
        '<div style="padding:12px 16px;background:var(--gray-50);border-bottom:1px solid var(--border);font-size:13px;color:var(--gray-600);">' +
          '共 ' + rechargeTotal + ' 条记录 · 累计充值 ' + formatHours(totalHours) + ' 课时' +
        '</div>' +
        renderBatchBar('recharges') +
        '<table class="data-table">' +
          '<thead><tr><th style="width:36px;"><input type="checkbox" onchange="toggleSelectAll(\'recharges\', this.checked)"></th><th>学生姓名</th><th>学号</th><th>充值课时</th><th>充值日期</th><th>备注</th><th>创建时间</th><th>操作</th></tr></thead>' +
          '<tbody>' + rowsHtml + '</tbody>' +
        '</table>' +
        renderPagination(rechargePage, rechargeTotalPages, 'loadRecharges', pageSizes.loadRecharges, rechargeTotal) +
      '</div>';
  } catch (err) {
    container.innerHTML = renderErrorState(getErrorMessage(err));
  }
}

function deleteRecharge(id) {
  confirmModal('确认删除', '确定要删除这条充值记录吗？此操作不会自动扣减学生课时，需手动调整。', async () => {
    try { await API.hourRecharge.delete(id); showToast('删除成功', 'success'); await loadRecharges(1); }
    catch (err) { showToast('删除失败: ' + getErrorMessage(err), 'error'); }
  }, '删除');
}

async function openRechargeSelectStudent() {
  const studentsResult = await API.student.list({ page: 1, page_size: 1000 });
  const students = studentsResult.data || [];
  showModal('选择学生充值', `
    <div style="max-height:400px;overflow-y:auto;">
      ${students.map(s => {
        const remaining = s.total_hours - s.completed_hours;
        return `<div style="padding:8px;border:1px solid var(--border);border-radius:8px;margin-bottom:4px;display:flex;justify-content:space-between;align-items:center;">
          <span>${escapeHtml(s.name)} <span class="text-gray">剩余: ${formatHours(remaining)}</span></span>
          <button class="btn btn-primary btn-sm" onclick="openRechargeModal(${s.id}, '${escapeHtml(s.name)}', ${remaining})">充值</button>
        </div>`;
      }).join('')}
    </div>
    <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`);
}

async function openEditRechargeModal(id) {
  try {
    const allRecords = await API.hourRecharge.list({ page: 1, page_size: 1000 });
    const record = (allRecords.data || []).find(r => r.id === id);
    if (!record) { showToast('记录不存在', 'error'); return; }
    showModal('编辑充值记录', `
      <div id="editRechargeError"></div>
      <div class="form-group"><label class="form-label">学生</label><input class="form-input" value="${escapeHtml(record.student_name)} (${escapeHtml(record.student_no)})" disabled></div>
      <div class="form-group"><label class="form-label">充值课时 *</label><input class="form-input" type="number" id="editRechargeHours" value="${record.hours}" min="0.5" step="0.5"></div>
      <div class="form-group"><label class="form-label">充值日期</label><input class="form-input" type="date" id="editRechargeDate" value="${escapeHtml(record.recharge_date)}"></div>
      <div class="form-group"><label class="form-label">备注</label><input class="form-input" id="editRechargeDesc" value="${escapeHtml(record.description || '')}"></div>
      ${renderFormFooter('submitEditRecharge(' + id + ', ' + record.student_id + ')')}
    `);
  } catch (err) { showToast('加载失败: ' + getErrorMessage(err), 'error'); }
}

async function submitEditRecharge(id, studentId) {
  const hours = parseFloat($('#editRechargeHours').value);
  const rechargeDate = $('#editRechargeDate').value;
  const description = $('#editRechargeDesc').value.trim();
  if (!hours || hours <= 0) { $('#editRechargeError').innerHTML = '<div class="form-error">请输入有效的课时数</div>'; return; }
  if (!rechargeDate) { $('#editRechargeError').innerHTML = '<div class="form-error">请选择日期</div>'; return; }
  try {
    await API.hourRecharge.update(id, { student_id: studentId, hours, recharge_date: rechargeDate, description });
    closeModal();
    showToast('修改成功', 'success');
    await loadRecharges(1);
  } catch (err) {
    $('#editRechargeError').innerHTML = `<div class="form-error">${escapeHtml(getErrorMessage(err))}</div>`;
  }
}

// ==================== 页面: 课时记录 ====================
let hourRecordSchedules = {}; // courseId -> [schedules]

function getTimePeriod(dateStr, courseId) {
  const d = new Date(dateStr + 'T00:00:00');
  if (isNaN(d)) return '';
  const dayOfWeek = d.getDay();
  const wd = weekdayLabel(dayOfWeek);
  const schedules = hourRecordSchedules[courseId] || [];
  const matching = schedules.filter(s => s.day_of_week === dayOfWeek);
  if (matching.length > 0) {
    const period = matching[0].period || getPeriod(matching[0].start_time);
    return `${wd} ${period}`;
  }
  return wd;
}

// ==================== 课程日历 ====================
async function renderCourseCalendar() {
  try {
    var body = $('#pageBody');
    var now = new Date();
    var currentMonth = now.getFullYear() + '-' + (now.getMonth() < 9 ? '0' : '') + (now.getMonth() + 1);

    body.innerHTML = ''
      + '<div class="calendar-page">'
      + '  <div class="calendar-header flex gap-3 items-center" style="margin-bottom:16px;flex-wrap:wrap;">'
      + '    <button class="btn btn-secondary" id="calPrev">&lt; 上月</button>'
      + '    <h3 id="calMonthLabel" style="margin:0;min-width:120px;text-align:center;">' + currentMonth + '</h3>'
      + '    <button class="btn btn-secondary" id="calNext">下月 &gt;</button>'
      + '    <button class="btn btn-secondary" id="calToday">今天</button>'
      + '    <div style="margin-left:auto;min-width:200px;">'
      +      createMultiSelectHTML('calCourseFilter', '全部课程')
      + '    </div>'
      + '  </div>'
      + '  <div id="calGrid"></div>'
      + '</div>';

    var viewMonth = currentMonth;
    var cachedEvents = [];
    var cachedCourses = [];
    var cachedHolidays = {};

    function pad2(n) { return n < 10 ? '0' + n : '' + n; }

    // 获取课程列表并填充筛选下拉框
    async function loadCourses() {
      try {
        var res = await API.course.list({ page: 1, page_size: 100 });
        cachedCourses = res.data || [];
        populateMultiSelect('calCourseFilter', cachedCourses.map(function(c) { return { value: c.id, label: c.name }; }));
      } catch(e) {
        console.error('loadCourses error:', e);
      }
    }

    function getFilteredEvents() {
      var filterParams = getMultiSelectFilterParams('calCourseFilter');
      if (filterParams === null) return cachedEvents;
      if (filterParams.length === 0) return cachedEvents;
      var idSet = {};
      filterParams.forEach(function(id) { idSet[id] = true; });
      return cachedEvents.filter(function(e) { return !!idSet[String(e.course_id)]; });
    }

    function renderGrid(month) {
      var events = getFilteredEvents();

      // 按日期分组
      var eventsByDate = {};
      events.forEach(function(e) {
        if (!eventsByDate[e.date]) eventsByDate[e.date] = [];
        eventsByDate[e.date].push(e);
      });

      // 计算日历网格
      var parts = month.split('-');
      var year = parseInt(parts[0], 10);
      var mon = parseInt(parts[1], 10);
      var firstDay = new Date(year, mon - 1, 1);
      var lastDay = new Date(year, mon, 0);
      var daysInMonth = lastDay.getDate();
      var today = new Date().toISOString().split('T')[0];

      // 周一作为第一天
      var weekDayNames = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
      var leadingBlanks = (firstDay.getDay() + 6) % 7;

      var html = '<div class="calendar-grid" style="display:grid;grid-template-columns:repeat(7,1fr);gap:4px;">';
      weekDayNames.forEach(function(wd) {
        html += '<div style="text-align:center;font-weight:bold;padding:8px;background:var(--bg-alt,#f5f5f5);border-radius:4px;">' + wd + '</div>';
      });

      // 前置空格
      for (var i = 0; i < leadingBlanks; i++) {
        html += '<div style="min-height:100px;background:var(--bg-alt,#fafafa);border-radius:4px;opacity:0.5;"></div>';
      }

      for (var d = 1; d <= daysInMonth; d++) {
        var dateStr = month + '-' + pad2(d);
        var dayEvents = eventsByDate[dateStr] || [];
        var isToday = dateStr === today;
        var dateObj = new Date(year, mon - 1, d);
        var dayOfWeek = dateObj.getDay();
        var isWeekend = (dayOfWeek === 0 || dayOfWeek === 6);
        var holidayInfo = cachedHolidays[dateStr];
        var isHoliday = holidayInfo && holidayInfo.is_holiday;
        var isCompDay = holidayInfo && holidayInfo.is_comp_day;

        // 计算背景色
        var bgStyle = 'background:var(--bg,#fff);';
        if (isHoliday) {
          bgStyle = 'background:#fef2f2;';
        } else if (isCompDay) {
          bgStyle = 'background:var(--bg,#fff);';
        } else if (isWeekend) {
          bgStyle = 'background:#f0fdf4;';
        }

        html += '<div style="min-height:100px;padding:4px;border:1px solid var(--border,#e0e0e0);border-radius:4px;' + (isToday ? 'border-color:var(--primary,#3b82f6);border-width:2px;' : '') + bgStyle + '">';
        var dateColor = 'var(--text-muted,#999)';
        if (isToday) dateColor = 'var(--primary,#3b82f6)';
        else if (isHoliday) dateColor = '#dc2626'
        else if (isWeekend && !isCompDay) dateColor = '#16a34a'
        html += '<div style="font-size:12px;color:' + dateColor + ';font-weight:' + (isToday ? 'bold' : 'normal') + ';margin-bottom:2px;display:flex;justify-content:space-between;align-items:center;">';
        html += '<span>' + d + '</span>';
        if (isHoliday && holidayInfo.holiday_name) {
          html += '<span style="font-size:10px;color:#dc2626;font-weight:normal;">' + escapeHtml(holidayInfo.holiday_name) + '</span>';
        } else if (isCompDay) {
          html += '<span style="font-size:10px;color:#f59e0b;font-weight:normal;">补班</span>';
        }
        html += '</div>';
        dayEvents.forEach(function(e) {
          var isMakeup = e.is_makeup === true;
          // 法定节假日不显示课程，除非是补课
          if (isHoliday && !isMakeup) {
            return;
          }
          var periodLabel = isMakeup ? '补课' : (e.period || getPeriod(e.start_time));
          var isPastOrToday = dateStr <= today;
          var attendedList = (e.attended_students || []);
          var enrolledList = (e.students || []);
          var isAttended = isPastOrToday && attendedList.length > 0;
          var tooltip;
          if (isPastOrToday) {
            tooltip = attendedList.length > 0
              ? '已上课学生: ' + attendedList.join('、')
              : '暂无上课记录';
          } else {
            tooltip = enrolledList.length > 0
              ? '选课学生: ' + enrolledList.join('、')
              : '暂无学生选课';
          }
          if (isMakeup) {
            tooltip = '(补课) ' + tooltip;
          }
          var periodColor = isMakeup ? '#ef4444' : (periodLabel === '上午' ? '#3b82f6' : periodLabel === '下午' ? '#f59e0b' : '#8b5cf6');
          var borderStyle, bgStyle, textColor;
          if (isAttended) {
            // 已上课：实心背景 + 白色文字
            bgStyle = 'background:' + periodColor + ';';
            textColor = 'color:#fff;';
            if (isMakeup) {
              borderStyle = 'border:1px dashed #fff;';
            } else {
              borderStyle = 'border-left:3px solid rgba(255,255,255,0.5);';
            }
          } else {
            // 未上课：浅色背景 + 彩色文字
            bgStyle = isMakeup ? 'background:#fef2f2;' : 'background:' + periodColor + '20;';
            textColor = 'color:' + periodColor + ';';
            borderStyle = isMakeup ? 'border:1px dashed #ef4444;' : 'border-left:3px solid ' + periodColor + ';';
          }
          var originalBg = isAttended ? periodColor : (isMakeup ? '#fef2f2' : periodColor + '20');
          html += '<div class="cal-event" data-course-id="' + e.course_id + '" data-date="' + dateStr + '" data-is-attended="' + isAttended + '" data-pc="' + periodColor + '" data-original-bg="' + originalBg + '" title="' + escapeHtml(tooltip) + '" style="cursor:pointer;margin-bottom:2px;padding:2px 4px;border-radius:3px;font-size:11px;' + bgStyle + textColor + borderStyle + ';">';
          html += '<span style="font-weight:500;">' + escapeHtml(e.course_name) + '</span>';
          html += ' <span style="font-size:10px;">' + escapeHtml(periodLabel) + '</span>';
          if (isAttended) {
            html += ' <span style="font-size:10px;">✓</span>';
          }
          if (e.start_time && e.end_time) {
            html += ' <span style="font-size:10px;">' + escapeHtml(e.start_time) + '-' + escapeHtml(e.end_time) + '</span>';
          }
          html += '</div>';
        });
        html += '</div>';
      }

      html += '</div>';

      // 图例
      html += '<div class="flex gap-4" style="margin-top:16px;font-size:12px;color:var(--text-muted,#999);flex-wrap:wrap;">';
      // 时间段
      html += '<span><span style="display:inline-block;width:12px;height:12px;background:#3b82f620;border-left:3px solid #3b82f6;border-radius:2px;vertical-align:middle;"></span> 上午</span>';
      html += '<span><span style="display:inline-block;width:12px;height:12px;background:#f59e0b20;border-left:3px solid #f59e0b;border-radius:2px;vertical-align:middle;"></span> 下午</span>';
      html += '<span><span style="display:inline-block;width:12px;height:12px;background:#8b5cf620;border-left:3px solid #8b5cf6;border-radius:2px;vertical-align:middle;"></span> 晚上</span>';
      // 补课
      html += '<span><span style="display:inline-block;width:12px;height:12px;background:#ef4444;color:#fff;border:1px dashed #fff;border-radius:2px;vertical-align:middle;"></span> 补课</span>';
      // 日期类型
      html += '<span><span style="display:inline-block;width:12px;height:12px;background:#fef2f2;border:1px solid #fecaca;border-radius:2px;vertical-align:middle;"></span> 法定节假日</span>';
      html += '<span><span style="display:inline-block;width:12px;height:12px;background:#f0fdf4;border:1px solid #bbf7d0;border-radius:2px;vertical-align:middle;"></span> 周末</span>';
      html += '<span style="margin-left:auto;">点击课程可跳转到课程管理页面</span>';
      html += '</div>';

      $('#calGrid').innerHTML = html;

      // 绑定课程事件（使用 addEventListener 替代内联 handler，兼容 Safari 13）
      var calEvents = document.querySelectorAll('.cal-event');
      for (var j = 0; j < calEvents.length; j++) {
          (function(el) {
            var pc = el.getAttribute('data-pc');
            var originalBg = el.getAttribute('data-original-bg');
            var courseId = el.getAttribute('data-course-id');
            var recordDate = el.getAttribute('data-date');
            var isAttended = el.getAttribute('data-is-attended') === 'true';
            el.addEventListener('click', function() { 
              if (isAttended) {
                // 已上课：跳转到课时记录页并筛选
                navigate('/hour-records');
                setTimeout(function() {
                  var startDateInput = $('#filterStartDate');
                  var endDateInput = $('#filterEndDate');
                  var courseSelect = $('#filterCourses');
                  if (startDateInput) startDateInput.value = recordDate;
                  if (endDateInput) endDateInput.value = recordDate;
                  if (courseSelect) {
                    var checkboxes = courseSelect.querySelectorAll('input[type="checkbox"]');
                    for (var i = 0; i < checkboxes.length; i++) {
                      checkboxes[i].checked = (checkboxes[i].value === courseId);
                    }
                    updateMultiSelectText('filterCourses');
                  }
                  loadHourRecords(1);
                }, 100);
              } else {
                // 未上课：跳转到课程页面
                navigate('/courses');
              }
            });
            el.addEventListener('mouseover', function() { el.style.background = pc + '30'; });
            el.addEventListener('mouseout', function() { el.style.background = originalBg; });
          })(calEvents[j]);
      }
    }

    async function loadCalendar(month) {
      viewMonth = month;
      $('#calMonthLabel').textContent = month;
      // 先获取日历事件并渲染（不依赖节假日数据）
      try {
        var calRes = await API.calendar.list(month);
        cachedEvents = calRes.events || [];
      } catch(e) {
        console.error('loadCalendar error:', e);
        cachedEvents = [];
      }
      renderGrid(month);
      // 异步获取节假日数据，不阻塞日历渲染
      API.calendar.holidays(month).then(function(holRes) {
        cachedHolidays = (holRes && holRes.holidays) || {};
        renderGrid(month);
      }).catch(function(e) {
        cachedHolidays = {};
      });
    }

    // 先加载课程列表，再加载日历
    await loadCourses();
    await loadCalendar(currentMonth);

    // 课程筛选变化时重新渲染网格
    var calCourseFilterEl = $('#calCourseFilter');
    if (calCourseFilterEl) {
      calCourseFilterEl.addEventListener('change', function() { renderGrid(viewMonth); });
    }

    $('#calPrev').addEventListener('click', function() {
      var parts = viewMonth.split('-');
      var prev = new Date(parseInt(parts[0], 10), parseInt(parts[1], 10) - 2, 1);
      loadCalendar(prev.getFullYear() + '-' + pad2(prev.getMonth() + 1));
    });
    $('#calNext').addEventListener('click', function() {
      var parts = viewMonth.split('-');
      var next = new Date(parseInt(parts[0], 10), parseInt(parts[1], 10), 1);
      loadCalendar(next.getFullYear() + '-' + pad2(next.getMonth() + 1));
    });
    $('#calToday').addEventListener('click', function() {
      loadCalendar(currentMonth);
    });
  } catch(err) {
    console.error('renderCourseCalendar error:', err);
    var body2 = $('#pageBody');
    if (body2) body2.innerHTML = '<div style="padding:20px;color:#dc2626;">课程日历加载失败: ' + escapeHtml(String(err)) + '</div>';
  }
}

async function renderHourRecords() {
  const body = $('#pageBody');
  body.innerHTML = `
    <div class="flex justify-between mb-6">
      <div></div>
      <div class="flex gap-2">
        <button class="btn btn-export" onclick="openExportHourModal()">${ICONS.download}导出记录</button>
        <button class="btn btn-chart" onclick="openHourChartModal()">${ICONS.chart}课时图表</button>
        <button class="btn btn-primary" onclick="openBatchHourModal()">${ICONS.plus}记录课时</button>
      </div>
    </div>
    <div class="filter-bar">
      <div class="form-group"><label class="form-label">课程</label>${createMultiSelectHTML('filterCourses', '全部课程')}</div>
      <div class="form-group"><label class="form-label">学生</label>${createMultiSelectHTML('filterStudents', '全部学生')}</div>
      <div class="form-group"><label class="form-label">状态</label>${createSingleSelectHTML('filterStudentStatus', '全部')}</div>
      <div class="form-group"><label class="form-label">开始日期</label><input class="form-input" type="date" id="filterStartDate"></div>
      <div class="form-group"><label class="form-label">结束日期</label><input class="form-input" type="date" id="filterEndDate"></div>
      <button class="btn btn-secondary" onclick="loadHourRecords(1)">${ICONS.search}筛选</button>
    </div>
    <div id="hourRecordsBody">${renderLoading()}</div>`;
  try {
    const [studentsResult, coursesResult] = await Promise.all([
      API.student.list({ page: 1, page_size: 1000 }), API.course.list({ page: 1, page_size: 100 }),
    ]);
    window.allHourRecordStudents = (studentsResult.data || []);
    const courses = (coursesResult.data || []);
    populateMultiSelect('filterCourses', courses.map(c => ({ value: c.id, label: c.name })));
    populateSingleSelect('filterStudentStatus', [
      { value: '', label: '全部' },
      { value: '0', label: '在读' },
      { value: '1', label: '已退学' }
    ], '');
    // 添加课程联动学生
    $('#filterCourses')._onchange = async function() {
      await filterHourRecordStudentsByCourse();
    };
    // 初始化时显示所有学生（不区分状态）
    await filterHourRecordStudentsByCourse();
    // 加载所有课程的排课信息，用于时间段匹配
    hourRecordSchedules = {};
    await Promise.all(courses.map(async c => {
      try {
        const scheds = await API.schedule.getByCourse(c.id);
        hourRecordSchedules[c.id] = scheds || [];
      } catch { hourRecordSchedules[c.id] = []; }
    }));
    await loadHourRecords(1);
  } catch (err) {
    $('#hourRecordsBody').innerHTML = renderErrorState(getErrorMessage(err));
  }
}

// 使用统一抽象的 linkStudentsByCourse 实现（消除重复逻辑）
async function filterHourRecordStudentsByCourse() {
  await linkStudentsByCourse('filterCourses', 'filterStudents', window.allHourRecordStudents || [], { multi: true });
}

// 使用统一抽象的 filterStudentsByStatus 实现（消除重复逻辑）
function filterHourRecordStudentsByStatus() {
  var statusEl = document.querySelector('input[name="filterStudentStatus"]:checked');
  var status = statusEl ? statusEl.value : '';
  filterStudentsByStatus('filterStudents', window.allHourRecordStudents || [], status);
}

let hourRecordPage = 1;
let hourRecordTotalPages = 1;
let hourRecordTotal = 0;

async function loadHourRecords(page = 1) {
  hourRecordPage = page;
  registerBatchConfig('hourRecords', API.hourRecord.batchDelete, loadHourRecords, '课时记录');
  const studentIds = getMultiSelectValues('filterStudents').map(v => parseInt(v));
  const courseIds = getMultiSelectValues('filterCourses').map(v => parseInt(v));
  const startDate = ($('#filterStartDate') || {}).value || '';
  const endDate = ($('#filterEndDate') || {}).value || '';
  const statusEl = document.querySelector('input[name="filterStudentStatus"]:checked');
  const studentStatus = statusEl ? statusEl.value : '';
  const container = $('#hourRecordsBody');
  container.innerHTML = renderLoading();
  try {
    const result = await API.hourRecord.list({ student_ids: studentIds, course_ids: courseIds, start_date: startDate, end_date: endDate, student_status: studentStatus, page, page_size: pageSizes.loadHourRecords });
    const records = result.data || [];
    hourRecordTotal = result.total || 0;
    hourRecordTotalPages = result.total_pages || 1;
    const filtered = records || [];
    if (filtered.length === 0) {
      container.innerHTML = renderEmptyState('', '暂无课时记录');
      return;
    }
    // 按日期 -> 课程分组
    const grouped = {};
    const dateOrder = [];
    const courseOrderByDate = {};
    filtered.forEach(r => {
      const d = r.record_date;
      const ckey = `${r.course_id || 0}`;
      if (!grouped[d]) { grouped[d] = {}; dateOrder.push(d); courseOrderByDate[d] = []; }
      if (!grouped[d][ckey]) {
        grouped[d][ckey] = { courseName: r.course_name || '未分类课程', courseId: r.course_id, students: [], total: 0, count: 0 };
        courseOrderByDate[d].push(ckey);
      }
      grouped[d][ckey].students.push(r);
      grouped[d][ckey].total += r.hours || 0;
      grouped[d][ckey].count += 1;
    });

    // 统一表头 + 按日期分组行（日期合并到时间段开头，学生人数合并到学生汇总开头）
    let totalHours = 0, totalRecords = 0;
    const allRows = dateOrder.map((date) => {
      const courseKeys = courseOrderByDate[date];
      let dayTotal = 0;
      courseKeys.forEach(ck => { dayTotal += grouped[date][ck].total; });
      totalHours += dayTotal;
      totalRecords += courseKeys.reduce((sum, ck) => sum + grouped[date][ck].count, 0);
      const courseRows = courseKeys.map((ck, idx) => {
        const g = grouped[date][ck];
        const rowId = `hr_${date}_${idx}`.replace(/[^a-zA-Z0-9_]/g, '_');
        const period = getTimePeriod(date, g.courseId);
        // 日期合并到时间段开头：每个日期的第一行显示日期
        const periodCell = idx === 0 ? `${escapeHtml(date)} ${escapeHtml(period)}` : escapeHtml(period);
        const studentNames = g.students.map(s => s.student_name).join('、');
        // 学生人数合并到学生汇总开头
        const summaryCell = `${g.students.length}人: ${escapeHtml(studentNames)}`;
        const firstInDate = idx === 0 ? ' hr-first-in-date' : '';
        const studentRows = g.students.map(s => `<tr>
          <td>${escapeHtml(s.student_name)}</td>
          <td>${escapeHtml(s.student_no)}</td>
          <td><span class="hour-badge">${formatHours(s.hours || 0)}</span></td>
          <td class="text-sm text-gray">${escapeHtml(s.description || '-')}</td>
          <td><div class="flex gap-2">
            <button class="btn-icon blue" onclick="openEditHourModal(${s.id})" title="编辑">${ICONS.edit}</button>
            <button class="btn-icon red" onclick="deleteHourRecord(${s.id})" title="删除">${ICONS.trash}</button>
          </div></td>
        </tr>`).join('');
        return `<tr class="hr-summary-row${firstInDate}" onclick="toggleHrDetail('${rowId}', event)">
          <td><input type="checkbox" class="row-checkbox-hourRecords" value="${g.students.map(function(s){return s.id;}).join(',')}" onchange="toggleRowSelection('hourRecords')" onclick="event.stopPropagation()"></td>
          <td style="padding-left:24px;">
            <span class="hr-toggle" id="${rowId}_toggle">${ICONS.chevron}</span>
            ${escapeHtml(g.courseName)}
          </td>
          <td class="text-sm text-gray">${periodCell}</td>
          <td class="text-sm text-gray hr-student-summary">${summaryCell}</td>
          <td><strong>${formatHours(g.total)}</strong></td>
        </tr>
        <tr class="hr-detail-row" id="${rowId}" style="display:none;">
          <td colspan="5" style="padding:0;">
            <table class="data-table hr-detail-table">
              <thead><tr><th>学生姓名</th><th>学号</th><th>消耗课时</th><th>备注</th><th>操作</th></tr></thead>
              <tbody>${studentRows}</tbody>
            </table>
          </td>
        </tr>`;
      }).join('');
      return courseRows;
    }).join('');

    container.innerHTML = `
      <div class="card">
        <div style="padding:12px 16px;background:var(--gray-50);border-bottom:1px solid var(--border);font-size:13px;color:var(--gray-600);">
          共 ${hourRecordTotal} 条记录 · 总消耗 ${formatHours(totalHours)} 课时
        </div>
        ${renderBatchBar('hourRecords')}
        <table class="data-table hr-unified-table">
          <thead><tr><th style="width:36px;"><input type="checkbox" onchange="toggleSelectAll('hourRecords', this.checked)"></th><th>课程</th><th>时间段</th><th>学生</th><th>总课时</th></tr></thead>
          <tbody>${allRows}</tbody>
        </table>
        ${renderPagination(hourRecordPage, hourRecordTotalPages, 'loadHourRecords', pageSizes.loadHourRecords, hourRecordTotal)}
      </div>`;
  } catch (err) {
    container.innerHTML = renderErrorState(getErrorMessage(err));
  }
}

let chartRawData = [];
let chartTab = 'daily';

function openExportHourModal() {
  const filterStartDate = ($('#filterStartDate') || {}).value || '';
  const filterEndDate = ($('#filterEndDate') || {}).value || '';
  showModal('导出课时记录', `
    <div class="form-group"><label class="form-label">导出条数</label>
      <select class="form-input" id="exportLimit" style="width:200px;">
        <option value="0">全部数据</option>
        <option value="100">最近100条</option>
        <option value="500">最近500条</option>
        <option value="1000">最近1000条</option>
        <option value="5000">最近5000条</option>
      </select>
    </div>

    <div style="font-size:12px;color:#6b7280;padding:12px;background:#f3f4f6;border-radius:4px;">
      <p>当前筛选条件：</p>
      <p>- 日期范围：${filterStartDate || '全部'} 至 ${filterEndDate || '全部'}</p>
      <p>- 课程：${getMultiSelectText('filterCourses') || '全部'}</p>
      <p>- 学生：${getMultiSelectText('filterStudents') || '全部'}</p>
    </div>
    <div class="modal-footer">
      <button class="btn btn-secondary" onclick="closeModal()">取消</button>
      <button class="btn btn-primary" onclick="exportHourRecords()">${ICONS.download}确认导出</button>
    </div>`, 'md');
}

async function exportHourRecords() {
  try {
    const studentIds = getMultiSelectValues('filterStudents').map(v => parseInt(v));
    const courseIds = getMultiSelectValues('filterCourses').map(v => parseInt(v));
    const startDate = ($('#filterStartDate') || {}).value || '';
    const endDate = ($('#filterEndDate') || {}).value || '';
    const statusEl = document.querySelector('input[name="filterStudentStatus"]:checked');
    const studentStatus = statusEl ? statusEl.value : '';
    const limit = parseInt(($('#exportLimit') || {}).value || '0');

    const result = await API.hourRecord.export({
      student_ids: studentIds,
      course_ids: courseIds,
      student_status: studentStatus,
      start_date: startDate,
      end_date: endDate,
      limit: limit
    });

    const csvContent = result.csv || '';
    if (!csvContent) {
      showToast('没有数据可导出', 'warning');
      return;
    }

    const todayStr = formatDate(new Date());
    const defaultFilename = '课时记录_' + todayStr + '.csv';
    const content = '\ufeff' + csvContent;

    // 优先使用 Wails 文件服务打开保存对话框
    var wailsFS = (window.go && window.go.main && window.go.main.FileService) ? window.go.main.FileService : null;
    if (wailsFS && typeof wailsFS.ExportToFile === 'function') {
      var savePath = await wailsFS.ExportToFile(content, defaultFilename, 'csv');
      if (!savePath) {
        showToast('已取消导出', 'info');
        return;
      }
      showToast('已保存到: ' + savePath, 'success');
      closeModal();
      return;
    }

    // 降级 1: File System Access API（浏览器环境）
    if (window.showSaveFilePicker && typeof window.showSaveFilePicker === 'function') {
      try {
        const handle = await window.showSaveFilePicker({
          suggestedName: defaultFilename,
          types: [{ description: 'CSV 文件', accept: { 'text/csv': ['.csv'] } }],
        });
        const writable = await handle.createWritable();
        await writable.write(content);
        await writable.close();
        showToast('导出成功', 'success');
        closeModal();
        return;
      } catch (err) {
        if (err && err.name === 'AbortError') {
          showToast('已取消导出', 'info');
          return;
        }
      }
    }

    // 降级 2: Blob 下载
    const blob = new Blob([content], { type: 'text/csv;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = defaultFilename;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    showToast('导出成功', 'success');
    closeModal();
  } catch (err) {
    showToast(getErrorMessage(err), 'error');
  }
}

async function openHourChartModal() {
  const studentsResult = await API.student.list({ page: 1, page_size: 1000 });
  const coursesResult = await API.course.list({ page: 1, page_size: 100 });
  const students = studentsResult.data || [];
  const courses = coursesResult.data || [];
  const filterStartDate = ($('#filterStartDate') || {}).value || '';
  const filterEndDate = ($('#filterEndDate') || {}).value || '';
  chartTab = 'daily';
  chartRawData = [];
  showModal('学生课时消耗图表', `
    <div class="flex gap-4 mb-4">
      <div class="form-group" style="flex:2;"><label class="form-label">选择课程</label>${createMultiSelectHTML('chartCourseSelect', '全部课程')}</div>
      <div class="form-group" style="flex:2;"><label class="form-label">选择学生</label>${createMultiSelectHTML('chartStudentSelect', '全部学生')}</div>
      <div class="form-group" style="flex:1;"><label class="form-label">状态</label>${createSingleSelectHTML('chartStudentStatus', '全部')}</div>
      <div class="form-group" style="flex:1;"><label class="form-label">开始日期</label><input class="form-input" type="date" id="chartStartDate" value="${filterStartDate}"></div>
      <div class="form-group" style="flex:1;"><label class="form-label">结束日期</label><input class="form-input" type="date" id="chartEndDate" value="${filterEndDate}"></div>
    </div>
    <div class="flex justify-between items-center mb-4">
      <div id="chartTabs">
        <button class="btn btn-sm btn-primary" id="tabDaily" onclick="switchChartTab('daily')">按日维度</button>
        <button class="btn btn-sm btn-secondary" id="tabWeekly" onclick="switchChartTab('weekly')" style="margin-left:8px;">按周维度</button>
        <button class="btn btn-sm btn-secondary" id="tabMonthly" onclick="switchChartTab('monthly')" style="margin-left:8px;">按月维度</button>
      </div>
      <button class="btn btn-primary" onclick="loadStudentChart()">${ICONS.chart}显示图表</button>
    </div>
    <div id="chartArea">${renderLoading()}</div>
    <div class="modal-footer"><button class="btn btn-secondary" onclick="closeModal()">关闭</button></div>`, 'lg');
  setTimeout(async () => {
    populateMultiSelect('chartCourseSelect', courses.map(c => ({ value: c.id, label: c.name })));
    populateSingleSelect('chartStudentStatus', [
      { value: '', label: '全部' },
      { value: '0', label: '在读' },
      { value: '1', label: '已退学' }
    ], '');
    // 添加课程联动学生
    $('#chartCourseSelect')._onchange = async function() {
      await filterChartStudentsByCourse(students);
    };
    // 添加状态筛选事件
    const statusOptions = document.querySelectorAll('input[name="chartStudentStatus"]');
    statusOptions.forEach(opt => {
      opt.onchange = () => filterChartStudentsByStatus(students);
    });
    // 初始化时显示所有学生（包括已退学），默认选中所有学生
    const studentOptions = (students || []).map(s => ({ value: s.id, label: `${s.name} (${s.student_id})` }));
    const defaultSelected = (students || []).map(s => s.id.toString());
    populateMultiSelect('chartStudentSelect', studentOptions, defaultSelected);
    $('#chartStartDate').onchange = () => {};
    $('#chartEndDate').onchange = () => {};
    loadStudentChart();
  }, 50);
}

// 使用统一抽象的 linkStudentsByCourse 实现（消除重复逻辑）
async function filterChartStudentsByCourse(allStudents) {
  await linkStudentsByCourse('chartCourseSelect', 'chartStudentSelect', allStudents, { multi: true, autoSelect: true });
}

// 使用统一抽象的 filterStudentsByStatus 实现（消除重复逻辑）
function filterChartStudentsByStatus(allStudents) {
  var statusEl = document.querySelector('input[name="chartStudentStatus"]:checked');
  var status = statusEl ? statusEl.value : '';
  filterStudentsByStatus('chartStudentSelect', allStudents, status);
}

async function loadStudentChart() {
  // 确保学生选择框至少有一个选中项（如果有选项但没选中，自动全选）
  const studentContainer = $('#chartStudentSelect');
  if (studentContainer) {
    const studentItems = studentContainer.querySelectorAll('.multi-select-options input[type="checkbox"]');
    const checkedCount = Array.from(studentItems).filter(cb => cb.checked).length;
    if (studentItems.length > 0 && checkedCount === 0) {
      studentItems.forEach(cb => cb.checked = true);
      studentContainer.querySelector('.ms-all-cb').checked = true;
      updateMultiSelectText('chartStudentSelect');
    }
  }

  // 确保课程选择框至少有一个选中项（如果有选项但没选中，自动全选）
  const courseContainer = $('#chartCourseSelect');
  if (courseContainer) {
    const courseItems = courseContainer.querySelectorAll('.multi-select-options input[type="checkbox"]');
    const checkedCount = Array.from(courseItems).filter(cb => cb.checked).length;
    if (courseItems.length > 0 && checkedCount === 0) {
      courseItems.forEach(cb => cb.checked = true);
      courseContainer.querySelector('.ms-all-cb').checked = true;
      updateMultiSelectText('chartCourseSelect');
    }
  }

  const studentIds = getMultiSelectValues('chartStudentSelect').map(v => parseInt(v));
  const courseIds = getMultiSelectValues('chartCourseSelect').map(v => parseInt(v));
  if (studentIds.length === 0) {
    $('#chartArea').innerHTML = '<p class="text-gray text-sm">请选择学生后查看图表</p>';
    return;
  }
  const startDate = ($('#chartStartDate') || {}).value || '';
  const endDate = ($('#chartEndDate') || {}).value || '';
  $('#chartArea').innerHTML = renderLoading();
  try {
    const res = await API.dashboard.studentChart(studentIds, courseIds, startDate, endDate);
    chartRawData = res.data || [];
    if (chartRawData.length === 0) {
      $('#chartTabs').style.display = 'none';
      $('#chartArea').innerHTML = '<p class="text-gray text-sm">所选学生在此时间范围内无课程排课数据</p>';
      return;
    }
    $('#chartTabs').style.display = '';
    renderStudentChart();
  } catch (err) { $('#chartArea').innerHTML = `<p class="text-red">加载失败: ${escapeHtml(getErrorMessage(err))}</p>`; }
}

function switchChartTab(tab) {
  chartTab = tab;
  ['Daily', 'Weekly', 'Monthly'].forEach(t => {
    const btn = $('#tab' + t);
    if (btn) btn.className = 'btn btn-sm ' + (tab === t.toLowerCase() ? 'btn-primary' : 'btn-secondary');
  });
  renderStudentChart();
}

function renderStudentChart() {
  const groupBy = chartTab === 'daily' ? 'day' : chartTab === 'weekly' ? 'week' : 'month';
  const grouped = {};
  const groupedDates = {};
  chartRawData.forEach(d => {
    let key;
    if (groupBy === 'day') {
      key = d.date;
    } else if (groupBy === 'week') {
      key = mondayOf(d.date).toISOString().split('T')[0];
    } else {
      key = d.date.slice(0, 7);
    }
    if (!grouped[key]) { grouped[key] = {}; groupedDates[key] = {}; }
    grouped[key][d.course_name] = (grouped[key][d.course_name] || 0) + d.hours;
    if (!groupedDates[key][d.course_name]) groupedDates[key][d.course_name] = new Set();
    groupedDates[key][d.course_name].add(d.date);
  });
  const courses = [...new Set(chartRawData.map(d => d.course_name))];
  const groups = Object.keys(grouped).sort().map(key => {
    let label;
    if (groupBy === 'month') {
      label = key;
    } else if (groupBy === 'week') {
      const mon = new Date(key);
      const sun = new Date(mon);
      sun.setDate(mon.getDate() + 6);
      label = `${fmtMD(mon)}~${fmtMD(sun)}`;
    } else {
      label = key.slice(5);
    }
    return {
      label,
      values: courses.map(c => ({ name: c, value: grouped[key][c] || 0 })),
      tooltip: (v) => {
        const dates = [...(groupedDates[key][v.name] || [])].sort();
        const dateStr = dates.length > 0 ? dates.join(', ') : key;
        return `${dateStr} ${v.name}: ${formatHours(v.value)} 课时`;
      },
    };
  });
  $('#chartArea').innerHTML = renderGroupedBarChart(groups, { width: 750, height: 280 });
}

function toggleHrDetail(rowId, event) {
  if (event) event.stopPropagation();
  const row = document.getElementById(rowId);
  if (!row) return;
  const toggle = document.getElementById(rowId + '_toggle');
  const isHidden = row.style.display === 'none';
  row.style.display = isHidden ? '' : 'none';
  if (toggle) {
    toggle.classList.toggle('expanded', isHidden);
  }
}

async function openEditHourModal(id) {
  try {
    const [studentsResult, coursesResult, recordResult] = await Promise.all([API.student.list(), API.course.list(), API.hourRecord.list({})]);
    // API 返回的是 { data: [...], total: ... } 格式，需要使用 .data
    const students = studentsResult.data || [];
    const courses = coursesResult.data || [];
    const allRecords = recordResult.data || [];
    const record = allRecords.find(r => r.id === id);
    if (!record) { showToast('记录不存在', 'error'); return; }
    showModal('编辑课时记录', `
      <div id="editHourError"></div>
      <div class="form-group"><label class="form-label">学生</label>${createSingleSelectHTML('editHrStudent', '请选择')}</div>
      <div class="form-group"><label class="form-label">课程</label>${createSingleSelectHTML('editHrCourse', '请选择')}</div>
      <div class="form-group"><label class="form-label">日期</label><input class="form-input" type="date" id="editHrDate" value="${escapeHtml(record.record_date)}"></div>
      <div class="form-group"><label class="form-label">消耗课时</label><input class="form-input" type="number" id="editHrHours" value="${record.hours}" min="0.5" step="0.5"></div>
      <div class="form-group"><label class="form-label">备注</label><textarea class="form-textarea" id="editHrDesc">${escapeHtml(record.description || '')}</textarea></div>
      ${renderFormFooter('submitEditHour(' + id + ')')}
    `);
    populateSingleSelect('editHrStudent', students.map(s => ({ value: s.id, label: `${s.name} (${s.student_id})` })), String(record.student_id));
    populateSingleSelect('editHrCourse', courses.map(c => ({ value: c.id, label: c.name })), String(record.course_id));
  } catch (err) { showToast('加载失败: ' + getErrorMessage(err), 'error'); }
}

async function submitEditHour(id) {
  const data = {
    student_id: parseInt(getSingleSelectValue('editHrStudent')),
    course_id: parseInt(getSingleSelectValue('editHrCourse')),
    record_date: $('#editHrDate').value,
    hours: parseFloat($('#editHrHours').value),
    description: $('#editHrDesc').value.trim(),
  };
  if (!data.record_date) { $('#editHourError').innerHTML = '<div class="form-error">请选择日期</div>'; return; }
  if (!data.hours || data.hours <= 0) { $('#editHourError').innerHTML = '<div class="form-error">请输入有效的课时数</div>'; return; }
  try {
    await API.hourRecord.update(id, data);
    closeModal();
    showToast('修改成功', 'success');
    await loadHourRecords(1);
  } catch (err) {
    $('#editHourError').innerHTML = `<div class="form-error">${escapeHtml(getErrorMessage(err))}</div>`;
  }
}

function deleteHourRecord(id) {
  confirmModal('确认删除', '确定要删除这条课时记录吗？删除后会自动扣减学生的已完成课时。', async () => {
    try { await API.hourRecord.delete(id); showToast('删除成功', 'success'); await loadHourRecords(1); }
    catch (err) { showToast('删除失败: ' + getErrorMessage(err), 'error'); }
  }, '删除');
}

async function openBatchHourModal() {
  const [coursesResult, studentsResult] = await Promise.all([
    API.course.list({ page: 1, page_size: 100 }),
    API.student.list({ page: 1, page_size: 1000 }),
  ]);
  const courses = coursesResult.data || [];
  const students = studentsResult.data || [];
  showModal('记录课时', `
    <div id="batchError"></div>
    <div class="form-group"><label class="form-label">课程 *</label>${createSingleSelectHTML('batchCourse', '请选择课程')}</div>
    <div class="form-group"><label class="form-label">选择学生（可多选）</label>${createMultiSelectHTML('batchStudentSelect', '请先选择课程')}</div>
    <div class="form-group"><label class="form-label">日期</label><input class="form-input" type="date" id="batchDate" value="${todayStr()}"></div>
    <div class="form-group"><label class="form-label">消耗课时</label><input class="form-input" type="number" id="batchHours" value="1" min="0.5" step="0.5"></div>
    <div class="form-group"><label class="form-label">备注</label><textarea class="form-textarea" id="batchDesc"></textarea></div>
    ${renderFormFooter('submitBatchHour()', '确认记录')}
  `, 'lg');
  // 缓存所有学生数据
  window.allBatchStudents = students;
  // 默认选中第一个课程
  const defaultCourseId = courses.length > 0 ? courses[0].id : '';
  populateSingleSelect('batchCourse', courses.map(c => ({ value: c.id, label: c.name })), defaultCourseId);
  // 添加课程联动学生
  $('#batchCourse')._onchange = async function() {
    await filterBatchStudentsByCourse();
  };
  // 如果有默认课程，初始化时就加载学生
  if (defaultCourseId) {
    await filterBatchStudentsByCourse();
  } else {
    populateMultiSelect('batchStudentSelect', [], []);
  }
}

// 使用统一抽象的 linkStudentsByCourse 实现（消除重复逻辑）
// 单选课程、仅显示在读学生、默认全选、自定义标签显示剩余课时
async function filterBatchStudentsByCourse() {
  var allStudents = window.allBatchStudents || [];
  await linkStudentsByCourse('batchCourse', 'batchStudentSelect', allStudents, {
    multi: false,
    activeOnly: true,
    autoSelect: true,
    labelFn: function(s) {
      var remaining = (s.total_hours || 0) - (s.completed_hours || 0);
      return escapeHtml(s.name) + ' (' + escapeHtml(s.student_id) + ') - 剩余: ' + formatHours(remaining);
    },
  });
}

async function submitBatchHour() {
  const studentIds = getMultiSelectValues('batchStudentSelect').map(v => parseInt(v));
  const courseId = parseInt(getSingleSelectValue('batchCourse'));
  const date = $('#batchDate').value;
  const hours = parseFloat($('#batchHours').value);
  const desc = $('#batchDesc').value.trim();
  if (studentIds.length === 0) { $('#batchError').innerHTML = '<div class="form-error">请至少选择一名学生</div>'; return; }
  if (courseId === 0) { $('#batchError').innerHTML = '<div class="form-error">请选择课程</div>'; return; }
  if (!date) { $('#batchError').innerHTML = '<div class="form-error">请选择日期</div>'; return; }
  try {
    await API.hourRecord.batchCreate({ student_ids: studentIds, course_id: courseId, hours, record_date: date, description: desc });
    closeModal();
    showToast('记录成功', 'success');
    await loadHourRecords(1);
  } catch (err) {
    $('#batchError').innerHTML = `<div class="form-error">${escapeHtml(getErrorMessage(err))}</div>`;
  }
}

// ==================== 页面: 通知中心 ====================
let notifFilter = 'all';

async function renderNotifications() {
  const body = $('#pageBody');
  body.innerHTML = `
    <div class="flex justify-between mb-6">
      <div class="flex gap-2">
        ${['all', 'unread', 'read'].map(f => `<button class="btn ${notifFilter === f ? 'btn-primary' : 'btn-secondary'} btn-sm" onclick="setNotifFilter('${f}')">${f === 'all' ? '全部' : f === 'unread' ? '未读' : '已读'}</button>`).join('')}
      </div>
    </div>
    <div class="card" id="notifList">${renderLoading()}</div>`;
  await loadNotifications(1);
}

function setNotifFilter(f) { notifFilter = f; renderNotifications(); }

let notifPage = 1;
let notifTotalPages = 1;
let notifTotal = 0;

async function loadNotifications(page = 1) {
  notifPage = page;
  registerBatchConfig('notifications', API.notification.batchDelete, loadNotifications, '通知', [
    {
      label: '批量标记已读',
      btnClass: 'btn-success',
      action: function(ids) {
        API.notification.batchRead(ids).then(function() {
          showToast('已标记 ' + ids.length + ' 条为已读', 'success');
          loadNotifications(1);
        }).catch(function(e) {
          showToast('操作失败: ' + e.message, 'error');
        });
      }
    }
  ]);
  const container = $('#notifList');
  try {
    const status = notifFilter === 'all' ? '' : notifFilter;
    const result = await API.notification.list(status, page, pageSizes.loadNotifications);
    const notifications = result.data || [];
    notifTotal = result.total || 0;
    notifTotalPages = result.total_pages || 1;

    if (!notifications || notifications.length === 0) {
      container.innerHTML = renderEmptyState(ICONS.bell, '暂无通知');
      return;
    }
    const statusColors = { unread: 'red', read: 'blue' };
    const statusLabels = { unread: '未读', read: '已读' };
    const typeLabels = { threshold: '课时预警', student_no_course: '未绑定课程', course_no_students: '课程无学生' };
    container.innerHTML = `
      <div class="card">
        ${renderBatchBar('notifications')}
        <div style="padding:8px 16px;border-bottom:1px solid var(--border);background:var(--gray-50);">
          <label style="font-size:13px;color:var(--gray-600);cursor:pointer;display:inline-flex;align-items:center;">
            <input type="checkbox" style="margin-right:6px;" onchange="toggleSelectAll('notifications', this.checked)"> 全选当前页
          </label>
        </div>
        ${notifications.map(n => {
          const notifType = n.type || 'threshold';
          const typeLabel = typeLabels[notifType] || '通知';
          let desc = '';
          if (notifType === 'threshold') {
            desc = `当前剩余 ${formatHours(n.current_hours || 0)} 课时，已低于预警阈值`;
          } else if (notifType === 'student_no_course') {
            desc = `学生 ${escapeHtml(n.student_name || '未知')} 尚未绑定任何课程`;
          } else if (notifType === 'course_no_students') {
            desc = `课程 ${escapeHtml(n.course_name)} 还没有学生申报`;
          }
          const title = notifType === 'course_no_students' ? typeLabel : (escapeHtml(n.student_name) || typeLabel);
          return `
          <div class="notif-item">
            <div style="display:flex;align-items:center;"><input type="checkbox" class="row-checkbox-notifications" value="${n.id}" onchange="toggleRowSelection('notifications')"></div>
            <div class="notif-icon ${statusColors[n.status] || 'gray'}" style="background:${n.status === 'unread' ? '#fee2e2' : '#dbeafe'};color:${n.status === 'unread' ? 'var(--red)' : 'var(--blue)'};">${ICONS.bell}</div>
            <div class="notif-content">
              <div class="flex gap-2 items-center"><span class="notif-title">${title}</span><span class="badge ${statusColors[n.status] || 'gray'}">${typeLabel}</span><span class="badge ${statusColors[n.status] || 'gray'}">${statusLabels[n.status]}</span></div>
              <div class="notif-desc">${desc}</div>
              <div class="notif-time">${formatDate(n.created_at)}</div>
            </div>
            <div class="flex gap-2">
              ${n.status === 'unread' ? `<button class="btn-icon green" onclick="markNotifRead(${n.id})" title="标记已读">${ICONS.check}</button>` : ''}
              <button class="btn btn-danger btn-sm" onclick="deleteNotif(${n.id})" title="删除">${ICONS.trash}</button>
            </div>
          </div>`;
        }).join('')}
        ${renderPagination(notifPage, notifTotalPages, 'loadNotifications', pageSizes.loadNotifications, notifTotal)}
      </div>`;
  } catch (err) {
    container.innerHTML = renderErrorState(getErrorMessage(err));
  }
}

async function deleteNotif(id) {
  confirmModal('删除通知', '确定要删除该通知吗？此操作无法撤销。', async () => {
    try {
      await API.notification.delete(id);
      showToast('删除成功', 'success');
      await loadNotifications(1);
      updateUnreadCount();
    } catch (err) { showToast(getErrorMessage(err), 'error'); }
  }, '删除');
}

async function markNotifRead(id) { try { await API.notification.markRead(id); showToast('已标记已读', 'success'); await loadNotifications(1); updateUnreadCount(); } catch (err) { showToast(getErrorMessage(err), 'error'); } }

// ==================== 页面: 操作日志 ====================
const entityLabels = { student: '学生', course: '课程', hour_record: '课时记录', hour_recharge: '充值记录', schedule: '课程', notification: '通知' };
const opLabels = { create: '创建', update: '修改', delete: '删除' };
const opColors = { create: 'green', update: 'blue', delete: 'red' };

async function renderLogs() {
  const body = $('#pageBody');
  body.innerHTML = `
    <div class="filter-bar">
      <div class="form-group"><label class="form-label">操作类型</label>${createMultiSelectHTML('logFilterOp', '全部')}</div>
      <div class="form-group"><label class="form-label">对象类型</label>${createMultiSelectHTML('logFilterEntity', '全部')}</div>
      <div class="form-group"><label class="form-label">开始日期</label><input class="form-input" type="date" id="logStartDate"></div>
      <div class="form-group"><label class="form-label">结束日期</label><input class="form-input" type="date" id="logEndDate"></div>
      <button class="btn btn-secondary" onclick="loadLogs(1)">${ICONS.search}筛选</button>
    </div>
    <div class="card"><div id="logsList">${renderLoading()}</div></div>`;
  populateMultiSelect('logFilterOp', Object.entries(opLabels).map(([v, l]) => ({ value: v, label: l })));
  populateMultiSelect('logFilterEntity', Object.entries(entityLabels).map(([v, l]) => ({ value: v, label: l })));
  await loadLogs(1);
}

let logPage = 1;
let logTotalPages = 1;
let logTotal = 0;

async function loadLogs(page = 1) {
  logPage = page;
  registerBatchConfig('logs', API.logs.batchDelete, loadLogs, '操作日志');
  const container = $('#logsList');
  if (!container) return;
  container.innerHTML = renderLoading();
  try {
    const params = { page, page_size: pageSizes.loadLogs };
    const ops = getMultiSelectValues('logFilterOp');
    const entities = getMultiSelectValues('logFilterEntity');
    var startDateEl = $('#logStartDate');
    var endDateEl = $('#logEndDate');
    var startDate = startDateEl ? startDateEl.value : '';
    var endDate = endDateEl ? endDateEl.value : '';
    if (ops.length) params.operation_types = ops;
    if (entities.length) params.entity_types = entities;
    if (startDate) params.start_date = startDate;
    if (endDate) params.end_date = endDate;

    const result = await API.logs.list(params);
    const logs = result.data || [];
    logTotal = result.total || 0;
    logTotalPages = result.total_pages || 1;

    if (!logs || logs.length === 0) {
      container.innerHTML = renderEmptyState(ICONS.journal, '暂无操作日志');
      return;
    }
    var rowsHtml = logs.map(function(l) {
      return '<tr>' + renderRowCheckbox('logs', l.id) +
        '<td class="text-sm text-gray">' + formatDate(l.created_at) + '</td>' +
        '<td><span class="badge ' + (opColors[l.operation_type] || 'gray') + '">' + (opLabels[l.operation_type] || l.operation_type) + '</span></td>' +
        '<td>' + escapeHtml(entityLabels[l.entity_type] || l.entity_type) + '</td>' +
        '<td>' + escapeHtml(l.description || '-') + '</td>' +
        '<td><button class="btn-icon red" title="删除" onclick="deleteLog(' + l.id + ')">' + ICONS.trash + '</button></td>' +
      '</tr>';
    }).join('');
    container.innerHTML = renderDataTable({
      batchKey: 'logs',
      headers: ['时间', '操作', '对象', '描述', '操作'],
      rowsHtml: rowsHtml,
      page: logPage,
      totalPages: logTotalPages,
      callbackName: 'loadLogs',
      pageSize: pageSizes.loadLogs,
      total: logTotal,
    });
  } catch (err) {
    container.innerHTML = renderErrorState(getErrorMessage(err));
  }
}

async function deleteLog(id) {
  confirmModal('删除日志', '确定要删除此条日志吗？', async () => {
    try {
      await API.logs.delete(id);
      showToast('日志已删除', 'success');
      await loadLogs(1);
    } catch (err) { showToast(getErrorMessage(err), 'error'); }
  }, '删除');
}

// ==================== 页面: 数据管理 ====================
function renderBackup() {
  $('#pageBody').innerHTML = `
    <div class="grid-2">
      <div class="card">
        <div class="card-header">${ICONS.download} 数据导出</div>
        <div class="card-body">
          <p class="text-sm text-gray mb-4">导出所有数据为 JSON 文件，包括学生、课程、课时记录、充值记录等。</p>
          <button class="btn btn-primary" style="width:100%;justify-content:center;" onclick="exportData()">${ICONS.download}导出数据</button>
        </div>
      </div>
      <div class="card">
        <div class="card-header">${ICONS.upload} 数据导入</div>
        <div class="card-body">
          <div class="alert alert-warning">导入数据将覆盖当前数据库中的所有数据。此操作无法撤销，请确保已备份当前数据。</div>
          <input type="file" id="importFile" accept=".json,application/json" style="display:none;" onchange="onImportFileSelected(event)">
          <button class="btn btn-success" style="width:100%;justify-content:center;" onclick="triggerImportFile()">${ICONS.upload}选择文件并导入</button>
        </div>
      </div>
    </div>`;
}

function triggerImportFile() {
  // 优先使用 Wails 运行时绑定的原生导入函数
  var wailsFS = (window.go && window.go.main && window.go.main.FileService) ? window.go.main.FileService : null;
  if (wailsFS && typeof wailsFS.ImportDataFromFile === 'function') {
    doWailsImport();
    return;
  }
  // 降级：使用隐藏的 file input
  var input = $('#importFile');
  if (!input) return;
  input.value = '';
  input.click();
}

async function doWailsImport() {
  try {
    var wailsFS = (window.go && window.go.main && window.go.main.FileService) ? window.go.main.FileService : null;
    var content = await wailsFS.ImportDataFromFile();
    if (!content) return; // 用户取消
    confirmModal('确认导入', '确定要导入数据吗？此操作将覆盖当前所有数据，且无法撤销！', async () => {
      try {
        await API.data.importAll(content);
        showToast('导入成功', 'success');
        setTimeout(() => location.reload(), 1500);
      } catch (err) { showToast('导入失败: ' + getErrorMessage(err), 'error'); }
    }, '确认导入');
  } catch (err) { showToast('选择文件失败: ' + getErrorMessage(err), 'error'); }
}

function onImportFileSelected(event) {
  const file = event.target.files[0];
  if (!file) return;
  confirmModal('确认导入', `确定要导入文件 "${escapeHtml(file.name)}" 吗？此操作将覆盖当前所有数据，且无法撤销！`, async () => {
    try {
      const text = await file.text();
      await API.data.importAll(text);
      showToast('导入成功', 'success');
      setTimeout(() => location.reload(), 1500);
    } catch (err) { showToast('导入失败: ' + getErrorMessage(err), 'error'); }
  }, '确认导入');
}

async function exportData() {
  try {
    const res = await API.data.exportAll();
    const content = res.data;
    const defaultName = `class_manager_backup_${todayStr()}.json`;
    // 优先使用 Wails 运行时绑定的原生导出函数
    var wailsFS = (window.go && window.go.main && window.go.main.FileService) ? window.go.main.FileService : null;
    if (wailsFS && typeof wailsFS.ExportDataToFile === 'function') {
      var savePath = await wailsFS.ExportDataToFile(content, defaultName);
      if (!savePath) return; // 用户取消
      showToast('已保存到: ' + savePath, 'success');
      return;
    }
    // 降级 1: File System Access API（Chrome/Edge 浏览器环境）
    if (window.showSaveFilePicker && typeof window.showSaveFilePicker === 'function') {
      try {
        const handle = await window.showSaveFilePicker({
          suggestedName: defaultName,
          types: [{ description: 'JSON 文件', accept: { 'application/json': ['.json'] } }],
        });
        const writable = await handle.createWritable();
        await writable.write(content);
        await writable.close();
        showToast('导出成功', 'success');
        return;
      } catch (err) {
        if (err && err.name === 'AbortError') return; // 用户取消
        // 其他错误降级为下载
      }
    }
    // 降级 2: Blob 下载
    const blob = new Blob([content], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = defaultName;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    showToast('导出成功，文件已下载到浏览器默认下载目录', 'success');
  } catch (err) { showToast('导出失败: ' + getErrorMessage(err), 'error'); }
}

// ==================== 初始化 ====================
async function updateUnreadCount() {
  try { const res = await API.notification.unreadCount(); unreadCount = res.count || 0; } catch { unreadCount = 0; }
  renderNav();
}

function renderNav() {
  const navList = $('#navList');
  navList.innerHTML = Object.entries(ROUTES).map(([path, route]) => `
    <li><button class="nav-item" data-path="${path}" data-tooltip="${route.title}" title="${route.title}" onclick="navigate('${path}')">
      ${ICONS[route.icon]} <span>${route.title}</span>
      ${path === '/notifications' && unreadCount > 0 ? `<span class="nav-badge">${unreadCount > 9 ? '9+' : unreadCount}</span>` : ''}
    </button></li>`).join('');
  // 高亮当前路由
  const hash = location.hash.slice(1) || '/';
  document.querySelectorAll('.nav-item').forEach(el => el.classList.toggle('active', el.dataset.path === hash));
}

function init() {
  renderNav();
  updateUnreadCount();
  setInterval(updateUnreadCount, 60000);
  window.addEventListener('hashchange', router);
  router();
  // 侧边栏折叠
  $('#collapseBtn').onclick = () => {
    const sidebar = $('#sidebar');
    const main = $('#mainContent');
    sidebar.classList.toggle('collapsed');
    main.classList.toggle('sidebar-collapsed');
    const icon = $('#collapseIcon');
    icon.innerHTML = sidebar.classList.contains('collapsed')
      ? '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>'
      : '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/>';
  };
}

// ==================== 日期选择器 Polyfill (Safari 13 / macOS 10.15 兼容) ====================
let _dpPopup = null;
let _dpNeedsPolyfill = null;

function needsDatePickerPolyfill() {
  if (_dpNeedsPolyfill !== null) return _dpNeedsPolyfill;
  // 通过功能检测而非 UserAgent，更可靠
  _dpNeedsPolyfill = !String.prototype.padStart || !('aspectRatio' in document.documentElement.style);
  return _dpNeedsPolyfill;
}

function getDpPopup() {
  if (_dpPopup) return _dpPopup;
  _dpPopup = document.createElement('div');
  _dpPopup.className = 'dp-popup';
  _dpPopup.style.display = 'none';
  document.body.appendChild(_dpPopup);
  return _dpPopup;
}

function showDatePicker(input) {
  const popup = getDpPopup();
  const val = input.value ? new Date(input.value + 'T00:00:00') : new Date();
  renderCalendar(popup, val.getFullYear(), val.getMonth(), input);
  const rect = input.getBoundingClientRect();
  popup.style.display = 'block';
  popup.style.left = rect.left + 'px';
  popup.style.top = (rect.bottom + 4) + 'px';
  if (popup.offsetTop + popup.offsetHeight > window.innerHeight) {
    popup.style.top = (rect.top - popup.offsetHeight - 4) + 'px';
  }
  setTimeout(function() {
    var handler = function(e) {
      if (!popup.contains(e.target) && e.target !== input) {
        popup.style.display = 'none';
        document.removeEventListener('click', handler, true);
      }
    };
    document.addEventListener('click', handler, true);
  }, 0);
}

function renderCalendar(popup, year, month, input) {
  var monthNames = ['1月','2月','3月','4月','5月','6月','7月','8月','9月','10月','11月','12月'];
  var today = new Date();
  var firstDay = new Date(year, month, 1);
  var startWeekday = firstDay.getDay();
  var daysInMonth = new Date(year, month + 1, 0).getDate();
  var currentValue = input.value;

  var html = '<div class="dp-header">' +
    '<button type="button" class="dp-nav" data-action="prev">&lsaquo;</button>' +
    '<span class="dp-title">' + year + '年' + monthNames[month] + '</span>' +
    '<button type="button" class="dp-nav" data-action="next">&rsaquo;</button>' +
    '</div><div class="dp-grid">';
  ['日','一','二','三','四','五','六'].forEach(function(d) { html += '<span class="dp-weekday">' + d + '</span>'; });
  for (var i = 0; i < startWeekday; i++) { html += '<span class="dp-empty"></span>'; }
  for (var d = 1; d <= daysInMonth; d++) {
    var dateStr = year + '-' + padStart(month + 1, 2, '0') + '-' + padStart(d, 2, '0');
    var cls = 'dp-day';
    if (today.getFullYear() === year && today.getMonth() === month && today.getDate() === d) cls += ' dp-today';
    if (currentValue === dateStr) cls += ' dp-selected';
    html += '<span class="' + cls + '" data-date="' + dateStr + '">' + d + '</span>';
  }
  html += '</div>';
  popup.innerHTML = html;

  popup.querySelectorAll('.dp-nav').forEach(function(btn) {
    btn.onclick = function(e) {
      e.stopPropagation();
      var ny = year, nm = month;
      if (btn.dataset.action === 'prev') { nm--; if (nm < 0) { nm = 11; ny--; } }
      else { nm++; if (nm > 11) { nm = 0; ny++; } }
      renderCalendar(popup, ny, nm, input);
    };
  });
  popup.querySelectorAll('.dp-day').forEach(function(day) {
    day.onclick = function(e) {
      e.stopPropagation();
      input.value = day.dataset.date;
      popup.style.display = 'none';
      input.dispatchEvent(new Event('change', { bubbles: true }));
    };
  });
}

function initDatePickerPolyfill(container) {
  if (!needsDatePickerPolyfill()) return;
  var root = container || document;
  root.querySelectorAll('input[type="date"]').forEach(function(input) {
    if (input.dataset.dpInit) return;
    input.dataset.dpInit = '1';
    input.style.cursor = 'pointer';
    input.readOnly = true;
    input.addEventListener('click', function(e) {
      e.preventDefault();
      showDatePicker(input);
    });
  });
}

document.addEventListener('DOMContentLoaded', init);

// ==================== 全局错误处理 (macOS 10.15 / Safari 13 兼容) ====================
// 捕获未处理的 JavaScript 错误，防止 WKWebView 因未捕获异常而崩溃
window.addEventListener('error', function(e) {
  console.error('全局错误:', e.message || e.error || e);
  // 阻止错误冒泡，防止 WKWebView 崩溃
  if (e && e.preventDefault) e.preventDefault();
  return true;
});

window.addEventListener('unhandledrejection', function(e) {
  console.error('未处理的 Promise 拒绝:', e.reason || e);
  if (e && e.preventDefault) e.preventDefault();
  return true;
});
