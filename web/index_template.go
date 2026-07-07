package web

import "html/template"

func newIndexTemplate() *template.Template {
	return template.Must(template.New("index").Parse(indexHTML))
}

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Yatori Web</title>
<style>
:root{
  color-scheme:light;
  --bg:#f6f8fb;
  --panel:#ffffff;
  --panel-2:#f9fbfd;
  --line:#d9e1ea;
  --text:#182230;
  --muted:#637083;
  --primary:#2563eb;
  --primary-2:#1d4ed8;
  --ok:#16803c;
  --warn:#b45309;
  --bad:#c2410c;
  --shadow:0 18px 45px rgba(24,34,48,.08);
}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","Microsoft YaHei",Arial,sans-serif}
button,input,select,textarea{font:inherit}
button{border:0;cursor:pointer}
button,input,select,textarea{min-width:0}
.shell{min-height:100vh;display:grid;grid-template-rows:auto 1fr}
.topbar{height:64px;display:flex;align-items:center;justify-content:space-between;padding:0 28px;border-bottom:1px solid var(--line);background:rgba(255,255,255,.88);backdrop-filter:saturate(160%) blur(12px)}
.brand{display:flex;align-items:center;gap:12px;font-weight:750;font-size:18px}
.mark{width:30px;height:30px;border-radius:8px;background:linear-gradient(135deg,#2563eb,#0f766e);box-shadow:0 8px 18px rgba(37,99,235,.25)}
.statusline{font-size:13px;color:var(--muted)}
.workspace{display:grid;grid-template-columns:minmax(420px,520px) minmax(0,1fr);gap:18px;padding:18px;max-width:1440px;width:100%;margin:0 auto}
.panel{background:var(--panel);border:1px solid var(--line);border-radius:8px;box-shadow:var(--shadow)}
.panel-header{display:flex;align-items:center;justify-content:space-between;gap:12px;padding:16px 18px;border-bottom:1px solid var(--line);flex-wrap:wrap}
.panel-title{font-size:15px;font-weight:720}
.panel-body{padding:18px}
.grid{display:grid;grid-template-columns:1fr 1fr;gap:14px}
.field{display:flex;flex-direction:column;gap:7px;min-width:0}
.field.full{grid-column:1/-1}
label{font-size:13px;font-weight:650;color:#334155}
input,select,textarea{width:100%;border:1px solid #cfd8e3;border-radius:7px;background:#fff;color:var(--text);padding:10px 11px;outline:none;transition:border-color .15s,box-shadow .15s}
textarea{min-height:74px;resize:vertical}
input:focus,select:focus,textarea:focus{border-color:var(--primary);box-shadow:0 0 0 3px rgba(37,99,235,.13)}
.actions{display:flex;gap:10px;flex-wrap:wrap;margin-top:16px}
.btn{display:inline-flex;align-items:center;justify-content:center;gap:8px;min-height:38px;padding:0 14px;border-radius:7px;font-weight:700;background:#e8eef8;color:#1e3a8a}
.btn.primary{background:var(--primary);color:#fff}
.btn.primary:hover{background:var(--primary-2)}
.btn.ghost{background:#fff;border:1px solid var(--line);color:#334155}
.btn:disabled{opacity:.55;cursor:not-allowed}
.course-tools{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-top:18px}
.course-list{margin-top:10px;border:1px solid var(--line);border-radius:8px;max-height:260px;overflow:auto;background:var(--panel-2)}
.empty{padding:22px;color:var(--muted);text-align:center;font-size:13px}
.course-item{display:grid;grid-template-columns:26px 1fr auto;gap:10px;align-items:start;padding:11px 12px;border-bottom:1px solid var(--line)}
.course-item:last-child{border-bottom:0}
.course-item input{width:18px;height:18px;margin-top:2px}
.course-name{font-weight:680;line-height:1.35}
.course-meta{margin-top:4px;color:var(--muted);font-size:12px;line-height:1.45}
.badge{display:inline-flex;align-items:center;height:24px;padding:0 8px;border-radius:999px;font-size:12px;font-weight:720;background:#eef2ff;color:#3730a3}
.split{display:grid;grid-template-rows:minmax(280px,42vh) minmax(260px,1fr);gap:18px;min-width:0}
.task-table{width:100%;border-collapse:collapse}
.task-table th,.task-table td{padding:11px 12px;border-bottom:1px solid var(--line);text-align:left;font-size:13px;vertical-align:top}
.task-table th{color:var(--muted);font-weight:720;background:#f8fafc}
.task-row{cursor:pointer}
.task-row:hover{background:#f8fbff}
.task-row.active{background:#eef5ff}
.status{font-weight:760;text-transform:uppercase;letter-spacing:.02em}
.status.running{color:var(--primary)}
.status.succeeded{color:var(--ok)}
.status.failed{color:var(--bad)}
.status.queued{color:var(--warn)}
.status.paused{color:#7c3aed}
.status.stopped{color:#64748b}
.course-item.ended{opacity:.62}
.course-item.ended .course-name{text-decoration:line-through}
.task-actions{display:flex;gap:6px;flex-wrap:wrap}
.btn.mini{min-height:30px;padding:0 9px;font-size:12px}
.time-stack{display:grid;gap:3px;min-width:150px}
.log-window{height:100%;min-height:240px;overflow:auto;background:#0f172a;color:#dbeafe;border-radius:8px;padding:12px;font-family:Consolas,"Cascadia Mono",monospace;font-size:12px;line-height:1.55}
.log-line{white-space:pre-wrap;border-bottom:1px solid rgba(148,163,184,.14);padding:5px 0}
.log-time{color:#93c5fd}
.log-level{color:#fbbf24}
.log-level.error{color:#fb7185}
.log-level.success{color:#86efac}
.hint{color:var(--muted);font-size:12px}
.query-box{margin-top:18px;border-top:1px solid var(--line);padding-top:16px}
.query-form{display:grid;grid-template-columns:150px 1fr auto;gap:10px;align-items:end}
.query-results{margin-top:10px;border:1px solid var(--line);border-radius:8px;background:var(--panel-2);max-height:180px;overflow:auto}
.query-item{display:grid;grid-template-columns:1fr auto;gap:10px;padding:10px 12px;border-bottom:1px solid var(--line);font-size:13px}
.query-item:last-child{border-bottom:0}
.top-tools{display:flex;align-items:center;gap:10px}
.auth{position:fixed;inset:0;display:grid;place-items:center;background:rgba(246,248,251,.94);z-index:20}
.auth-card{width:min(380px,calc(100vw - 32px));background:#fff;border:1px solid var(--line);border-radius:8px;padding:18px;box-shadow:var(--shadow)}
.auth-card h1{font-size:18px;margin:0 0 12px}
.auth-card form{display:grid;gap:10px}
.auth-tabs{display:grid;grid-template-columns:1fr 1fr;gap:8px;margin-bottom:12px}
.auth-tabs button{height:34px;border-radius:7px;background:#eef2f7;color:#334155;font-weight:700}
.auth-tabs button.active{background:var(--primary);color:#fff}
.auth-error{min-height:18px;color:var(--bad);font-size:13px}
.license-row{display:grid;grid-template-columns:1fr auto;gap:10px}
.hidden{display:none!important}
@media (max-width:980px){
  .workspace{grid-template-columns:1fr;padding:12px}
  .topbar{padding:0 16px}
  .grid{grid-template-columns:1fr}
  .split{grid-template-rows:auto auto}
  .query-form{grid-template-columns:1fr}
  .license-row{grid-template-columns:1fr}
}
@media (max-width:680px){
  body{background:#fff}
  .topbar{height:auto;min-height:58px;align-items:flex-start;gap:8px;padding:10px 12px;flex-direction:column}
  .brand{font-size:16px}.mark{width:26px;height:26px}
  .top-tools{width:100%;justify-content:space-between;gap:8px}
  .statusline{line-height:1.4}
  .workspace{padding:8px;gap:10px}
  .panel{box-shadow:none;border-radius:8px}
  .panel-header{padding:12px}
  .panel-body{padding:12px}
  input,select,textarea{min-height:42px}
  .actions{display:grid;grid-template-columns:1fr 1fr;gap:8px}
  .actions .btn,.license-row .btn{width:100%}
  .course-tools{align-items:flex-start;flex-direction:column}
  .course-tools .actions{width:100%;grid-template-columns:1fr 1fr}
  .course-list{max-height:220px}
  .course-item{grid-template-columns:24px 1fr;gap:8px;padding:10px}
  .course-item input{width:18px;height:18px}
  .badge{grid-column:2;justify-self:start;margin-top:2px}
  .split{gap:10px}
  .task-table,.task-table tbody,.task-table tr,.task-table td{display:block;width:100%}
  .task-table thead{display:none}
  .task-table td{display:grid;grid-template-columns:72px minmax(0,1fr);gap:8px;padding:6px 0;border:0;white-space:normal}
  .task-table td::before{content:attr(data-label);color:var(--muted);font-weight:650}
  .task-row{padding:10px 12px;border-bottom:1px solid var(--line)}
  .task-row.active{border-left:3px solid var(--primary)}
  .time-stack{min-width:0}
  .task-actions{gap:8px}.task-actions .btn{flex:1;min-width:72px}
  .log-window{height:260px;min-height:260px}
  .query-results{max-height:240px}.query-item{grid-template-columns:1fr}
}
</style>
</head>
<body>
<div class="auth" id="authPanel">
  <div class="auth-card">
    <h1>Yatori Web</h1>
    <div class="auth-tabs">
      <button type="button" class="active" id="loginTab">登录</button>
      <button type="button" id="registerTab">注册</button>
    </div>
    <form id="authForm">
      <input id="authUsername" autocomplete="username" placeholder="用户名" required>
      <input id="authPassword" type="password" autocomplete="current-password" placeholder="密码" required>
      <button class="btn primary" type="submit" id="authSubmit">登录</button>
      <div class="auth-error" id="authError"></div>
    </form>
  </div>
</div>
<div class="shell">
  <header class="topbar">
    <div class="brand"><span class="mark"></span><span>Yatori Web</span></div>
    <div class="top-tools">
      <div class="statusline" id="summary">准备就绪</div>
      <button class="btn ghost" type="button" id="logoutBtn">退出</button>
    </div>
  </header>
  <main class="workspace">
    <section class="panel">
      <div class="panel-header">
        <div class="panel-title">任务配置</div>
        <span class="hint">先获取课程，再提交任务</span>
      </div>
      <div class="panel-body">
        <form id="taskForm">
          <div class="grid">
            <div class="field">
              <label for="platform">平台</label>
              <select id="platform" name="platform">
                <option value="haiqikeji">海奇科技</option>
                <option value="yinghua">英华</option>
                <option value="xuexitong">学习通</option>
              </select>
            </div>
            <div class="field">
              <label for="account">账号</label>
              <input id="account" name="account" autocomplete="username" required>
            </div>
            <div class="field">
              <label for="password">密码</label>
              <input id="password" name="password" type="password" autocomplete="current-password" required>
            </div>
            <div class="field">
              <label for="preUrl">平台地址</label>
              <input id="preUrl" name="preUrl" placeholder="海奇/英华填写站点地址">
            </div>
            <div class="field">
              <label for="aiUrl">AI 地址</label>
              <input id="aiUrl" name="aiUrl" placeholder="可选">
            </div>
            <div class="field">
              <label for="aiModel">AI 模型</label>
              <input id="aiModel" name="aiModel" placeholder="可选">
            </div>
            <div class="field">
              <label for="aiKey">AI 密钥</label>
              <input id="aiKey" name="aiKey" placeholder="可选">
            </div>
            <div class="field">
              <label for="aiType">AI 类型</label>
              <select id="aiType" name="aiType">
                <option value="">默认</option>
                <option value="OPENAI">OpenAI</option>
                <option value="DEEPSEEK">DeepSeek</option>
                <option value="CHATGLM">ChatGLM</option>
                <option value="TONGYI">通义</option>
                <option value="DOUBAO">豆包</option>
                <option value="OTHER">Other</option>
              </select>
            </div>
            <div class="field full">
              <label for="message">备注</label>
              <textarea id="message" name="message" placeholder="可选"></textarea>
            </div>
            <div class="field full">
              <label for="licenseKey">卡密</label>
              <div class="license-row">
                <input id="licenseKey" name="licenseKey" placeholder="输入卡密后验证">
                <button class="btn ghost" type="button" id="verifyLicense">验证卡密</button>
              </div>
              <span class="hint" id="licenseHint">提交任务前需要验证卡密</span>
            </div>
          </div>
          <div class="course-tools">
            <div>
              <div class="panel-title">课程</div>
              <div class="hint" id="courseHint">未选择时默认执行全部课程</div>
            </div>
            <div class="actions" style="margin-top:0">
              <button class="btn ghost" type="button" id="fetchCourses">获取课程</button>
              <button class="btn ghost" type="button" id="toggleCourses" disabled>全选</button>
            </div>
          </div>
          <div class="course-list" id="courses"><div class="empty">暂无课程</div></div>
          <div class="actions">
            <button class="btn primary" type="submit" id="submitTask">提交任务</button>
            <button class="btn ghost" type="button" id="refreshTasks">刷新任务</button>
          </div>
        </form>
        <div class="query-box">
          <div class="panel-title">查询状态</div>
          <div class="query-form">
            <div class="field">
              <label for="queryPlatform">平台</label>
              <select id="queryPlatform">
                <option value="haiqikeji">海奇科技</option>
                <option value="yinghua">英华</option>
                <option value="xuexitong">学习通</option>
              </select>
            </div>
            <div class="field">
              <label for="queryAccount">登录账号</label>
              <input id="queryAccount" placeholder="输入提交任务时的账号">
            </div>
            <button class="btn ghost" type="button" id="queryTasks">查询</button>
          </div>
          <div class="query-results" id="queryResults"><div class="empty">输入账号后查询运行状态</div></div>
        </div>
      </div>
    </section>
    <section class="split">
      <section class="panel">
        <div class="panel-header">
          <div class="panel-title">任务</div>
          <span class="hint" id="taskHint">自动刷新中</span>
        </div>
        <div class="panel-body" style="padding:0;overflow:auto">
          <table class="task-table">
            <thead><tr><th>任务</th><th>账号</th><th>时间</th><th>状态</th><th>课程</th><th>控制</th></tr></thead>
            <tbody id="tasks"><tr><td colspan="6" class="empty">加载中</td></tr></tbody>
          </table>
        </div>
      </section>
      <section class="panel">
        <div class="panel-header">
          <div class="panel-title">实时日志</div>
          <span class="hint" id="logHint">选择一个任务查看</span>
        </div>
        <div class="panel-body">
          <div class="log-window" id="logs"><div class="log-line">等待任务日志...</div></div>
        </div>
      </section>
    </section>
  </main>
</div>
<script>
const state = { courses: [], tasks: [], selectedTask: "", events: null, user: null, authMode: "login", verifiedLicense: "" };
const $ = (id) => document.getElementById(id);
const statusText = {queued:"排队中",running:"运行中",paused:"已暂停",stopped:"已停止",succeeded:"已完成",failed:"失败"};

function payloadFromForm(){
  return {
    platform:$("platform").value,
    account:$("account").value.trim(),
    password:$("password").value,
    preUrl:$("preUrl").value.trim(),
    aiUrl:$("aiUrl").value.trim(),
    aiModel:$("aiModel").value.trim(),
    aiKey:$("aiKey").value.trim(),
    aiType:$("aiType").value,
    message:$("message").value.trim(),
    licenseKey:$("licenseKey").value.trim(),
    courseIds:[...document.querySelectorAll("input[name=courseId]:checked")].map(i=>i.value)
  };
}

async function postJSON(url, data){
  const res = await fetch(url,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(data)});
  const text = await res.text();
  if(!res.ok) throw new Error(text || res.statusText);
  return text ? JSON.parse(text) : {};
}

function escapeHTML(value){
  return String(value ?? "").replace(/[&<>"']/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"}[c]));
}

function renderCourses(){
  const box = $("courses");
  $("toggleCourses").disabled = state.courses.length === 0;
  if(!state.courses.length){
    box.innerHTML = '<div class="empty">暂无课程</div>';
    $("courseHint").textContent = "未选择时默认执行全部课程";
    return;
  }
  box.innerHTML = state.courses.map(c =>
    '<label class="course-item ' + (c.ended ? "ended" : "") + '">' +
      '<input type="checkbox" name="courseId" value="' + escapeHTML(c.id) + '"' + (c.ended ? " disabled" : "") + '>' +
      '<span>' +
        '<span class="course-name">' + escapeHTML(c.name || c.id) + '</span>' +
        '<span class="course-meta">' + escapeHTML(c.meta || "课程ID: " + c.id) + '</span>' +
      '</span>' +
      '<span class="badge">' + escapeHTML(c.progress || (c.ended ? "不可选" : "可选")) + '</span>' +
    '</label>').join("");
  $("courseHint").textContent = "已获取 " + state.courses.length + " 门课程";
}

function renderTasks(){
  const body = $("tasks");
  if(!state.tasks.length){
    body.innerHTML = '<tr><td colspan="6" class="empty">暂无任务</td></tr>';
    return;
  }
  body.innerHTML = state.tasks.slice().reverse().map(t =>
    '<tr class="task-row ' + (t.id===state.selectedTask ? "active" : "") + '" data-id="' + escapeHTML(t.id) + '">' +
      '<td data-label="任务"><strong>' + escapeHTML(t.platform) + '</strong><br><span class="hint">' + escapeHTML(t.id) + '</span><br><span class="hint">' + escapeHTML(t.message || "") + '</span></td>' +
      '<td data-label="账号">' + escapeHTML(t.account) + '</td>' +
      '<td data-label="时间"><span class="time-stack">' +
        '<span>提交 ' + escapeHTML(formatTime(t.createdAt)) + '</span>' +
        '<span>开始 ' + escapeHTML(formatTime(t.startedAt)) + '</span>' +
        '<span>结束 ' + escapeHTML(formatTime(t.endedAt)) + '</span>' +
        '<span>运行 ' + escapeHTML(formatDuration(t)) + '</span>' +
      '</span></td>' +
      '<td data-label="状态"><span class="status ' + escapeHTML(t.status) + '">' + escapeHTML(statusText[t.status] || t.status) + '</span></td>' +
      '<td data-label="课程">' + ((t.courseIds && t.courseIds.length) ? escapeHTML(t.courseIds.length + " 门") : "全部") + '</td>' +
      '<td data-label="控制"><span class="task-actions">' + actionButtons(t) + '</span></td>' +
    '</tr>').join("");
  document.querySelectorAll(".task-row").forEach(row => row.addEventListener("click", () => selectTask(row.dataset.id)));
  document.querySelectorAll("[data-action]").forEach(btn => btn.addEventListener("click", (event) => {
    event.stopPropagation();
    controlTask(btn.dataset.id, btn.dataset.action);
  }));
}

function formatTime(value){
  if(!value) return "-";
  const date = new Date(value);
  if(Number.isNaN(date.getTime())) return "-";
  return date.toLocaleString();
}

function formatDuration(task){
  let seconds = Number(task.runtimeSeconds || 0);
  if(task.status === "running" && task.startedAt){
    seconds = Math.max(seconds, Math.floor((Date.now() - new Date(task.startedAt).getTime()) / 1000));
  }
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = seconds % 60;
  if(h) return h + "时" + m + "分";
  if(m) return m + "分" + s + "秒";
  return s + "秒";
}

function actionButtons(task){
  const id = escapeHTML(task.id);
  if(task.status === "running"){
    return '<button class="btn mini ghost" data-id="' + id + '" data-action="pause">暂停</button>' +
      '<button class="btn mini ghost" data-id="' + id + '" data-action="stop">停止</button>';
  }
  if(task.status === "paused" || task.status === "queued"){
    return '<button class="btn mini primary" data-id="' + id + '" data-action="resume">启动</button>' +
      '<button class="btn mini ghost" data-id="' + id + '" data-action="stop">停止</button>';
  }
  return '<span class="hint">不可操作</span>';
}

async function controlTask(id, action){
  try{
    await postJSON("/tasks/" + encodeURIComponent(id) + "/control", {action});
    await loadTasks();
    if(state.selectedTask === id) selectTask(id);
  }catch(err){
    alert(err.message);
  }
}

function appendLog(log){
  const box = $("logs");
  const at = log.at ? new Date(log.at).toLocaleTimeString() : new Date().toLocaleTimeString();
  const level = escapeHTML(log.level || "info");
  const line = document.createElement("div");
  line.className = "log-line";
  line.innerHTML = '<span class="log-time">' + escapeHTML(at) + '</span> <span class="log-level ' + level + '">' + level + '</span> ' + escapeHTML(log.message || "");
  if(box.dataset.empty !== "false"){ box.innerHTML = ""; box.dataset.empty = "false"; }
  box.appendChild(line);
  box.scrollTop = box.scrollHeight;
}

async function loadTasks(){
  const res = await fetch("/tasks");
  if(res.status === 401){
    showAuth();
    return;
  }
  state.tasks = await res.json();
  renderTasks();
  const running = state.tasks.filter(t => t.status === "running").length;
  const userText = state.user ? state.user.username + " · " : "";
  $("summary").textContent = userText + (state.tasks.length ? "任务 " + state.tasks.length + " 个，运行中 " + running + " 个" : "准备就绪");
}

async function selectTask(id){
  state.selectedTask = id;
  renderTasks();
  if(state.events){ state.events.close(); state.events = null; }
  $("logs").innerHTML = "";
  $("logs").dataset.empty = "true";
  $("logHint").textContent = id;
  const logs = await fetch("/tasks/" + encodeURIComponent(id) + "/logs").then(r=>r.json());
  logs.forEach(appendLog);
  state.events = new EventSource("/tasks/" + encodeURIComponent(id) + "/events?after=" + logs.length);
  state.events.onmessage = (event) => appendLog(JSON.parse(event.data));
  state.events.addEventListener("done", () => {
    if(state.events){ state.events.close(); state.events = null; }
  });
}

function renderQueryResults(items){
  const box = $("queryResults");
  if(!items.length){
    box.innerHTML = '<div class="empty">没有找到该账号的任务</div>';
    return;
  }
  box.innerHTML = items.slice().reverse().map(t =>
    '<div class="query-item">' +
      '<div>' +
        '<strong>' + escapeHTML(statusText[t.status] || t.status) + '</strong> · ' + escapeHTML(t.platform) + ' · ' + escapeHTML(t.id) +
        '<div class="hint">' + escapeHTML(t.message || "") + '</div>' +
        '<div class="hint">提交 ' + escapeHTML(formatTime(t.createdAt)) + ' · 运行 ' + escapeHTML(formatDuration(t)) + '</div>' +
      '</div>' +
      '<button class="btn mini ghost" type="button" data-query-id="' + escapeHTML(t.id) + '">查看</button>' +
    '</div>').join("");
  document.querySelectorAll("[data-query-id]").forEach(btn => btn.addEventListener("click", async () => {
    await loadTasks();
    selectTask(btn.dataset.queryId);
  }));
}

async function queryTaskStatus(){
  const platform = $("queryPlatform").value;
  const account = $("queryAccount").value.trim();
  if(!account){
    $("queryResults").innerHTML = '<div class="empty">请输入登录账号</div>';
    return;
  }
  const params = new URLSearchParams({platform, account});
  const res = await fetch("/task-query?" + params.toString());
  const text = await res.text();
  if(!res.ok) throw new Error(text || res.statusText);
  renderQueryResults(text ? JSON.parse(text) : []);
}

$("fetchCourses").addEventListener("click", async () => {
  const btn = $("fetchCourses");
  btn.disabled = true;
  btn.textContent = "获取中";
  try{
    state.courses = await postJSON("/courses", payloadFromForm());
    renderCourses();
  }catch(err){
    $("courses").innerHTML = '<div class="empty">' + escapeHTML(err.message) + '</div>';
  }finally{
    btn.disabled = false;
    btn.textContent = "获取课程";
  }
});

$("toggleCourses").addEventListener("click", () => {
  const items = [...document.querySelectorAll("input[name=courseId]:not(:disabled)")];
  const shouldCheck = items.some(i => !i.checked);
  items.forEach(i => i.checked = shouldCheck);
  $("toggleCourses").textContent = shouldCheck ? "取消全选" : "全选";
});

$("taskForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  const licenseKey = $("licenseKey").value.trim();
  if(!state.user){
    showAuth();
    return;
  }
  if(!licenseKey || licenseKey !== state.verifiedLicense){
    alert("请先验证卡密");
    return;
  }
  const btn = $("submitTask");
  btn.disabled = true;
  try{
    const result = await postJSON("/submit", payloadFromForm());
    await loadTasks();
    if(result.task && result.task.id) selectTask(result.task.id);
  }catch(err){
    alert(err.message);
  }finally{
    btn.disabled = false;
  }
});

$("refreshTasks").addEventListener("click", loadTasks);
$("queryTasks").addEventListener("click", async () => {
  const btn = $("queryTasks");
  btn.disabled = true;
  try{
    await queryTaskStatus();
  }catch(err){
    $("queryResults").innerHTML = '<div class="empty">' + escapeHTML(err.message) + '</div>';
  }finally{
    btn.disabled = false;
  }
});
$("queryAccount").addEventListener("keydown", (event) => {
  if(event.key === "Enter"){
    event.preventDefault();
    $("queryTasks").click();
  }
});

function showAuth(){
  $("authPanel").classList.remove("hidden");
}

function hideAuth(){
  $("authPanel").classList.add("hidden");
  $("authError").textContent = "";
}

function setAuthMode(mode){
  state.authMode = mode;
  $("loginTab").classList.toggle("active", mode === "login");
  $("registerTab").classList.toggle("active", mode === "register");
  $("authSubmit").textContent = mode === "login" ? "登录" : "注册";
}

async function refreshMe(){
  const res = await fetch("/auth/me");
  if(res.status === 401){
    state.user = null;
    showAuth();
    return false;
  }
  const data = await res.json();
  state.user = data.user;
  hideAuth();
  return true;
}

async function authSubmit(){
  const username = $("authUsername").value.trim();
  const password = $("authPassword").value;
  const endpoint = state.authMode === "login" ? "/auth/login" : "/auth/register";
  await postJSON(endpoint, {username, password});
  if(state.authMode === "register"){
    await postJSON("/auth/login", {username, password});
  }
  await refreshMe();
  await loadTasks();
}

async function logout(){
  await fetch("/auth/logout", {method:"POST"});
  state.user = null;
  state.tasks = [];
  state.verifiedLicense = "";
  $("licenseHint").textContent = "提交任务前需要验证卡密";
  renderTasks();
  showAuth();
}

$("authForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  try{
    await authSubmit();
  }catch(err){
    $("authError").textContent = err.message;
  }
});
$("loginTab").addEventListener("click", () => setAuthMode("login"));
$("registerTab").addEventListener("click", () => setAuthMode("register"));
$("logoutBtn").addEventListener("click", logout);
$("verifyLicense").addEventListener("click", async () => {
  const key = $("licenseKey").value.trim();
  if(!key){
    $("licenseHint").textContent = "请输入卡密";
    return;
  }
  try{
    await postJSON("/license/verify", {key});
    state.verifiedLicense = key;
    $("licenseHint").textContent = "卡密验证成功";
  }catch(err){
    state.verifiedLicense = "";
    $("licenseHint").textContent = err.message;
  }
});
$("licenseKey").addEventListener("input", () => {
  if($("licenseKey").value.trim() !== state.verifiedLicense){
    state.verifiedLicense = "";
    $("licenseHint").textContent = "提交任务前需要验证卡密";
  }
});

refreshMe().then(ok => {
  if(ok) loadTasks();
});
setInterval(loadTasks, 3500);
</script>
</body>
</html>`
