package web

import "html/template"

func newAdminTemplate() *template.Template {
	return template.Must(template.New("admin").Parse(adminHTML))
}

const adminHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Yatori Admin</title>
<style>
:root{--bg:#f5f7fb;--panel:#fff;--line:#d8e0ea;--text:#172033;--muted:#667085;--primary:#2563eb;--ok:#15803d;--warn:#a16207;--bad:#b42318;--paused:#6d28d9}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","Microsoft YaHei",Arial,sans-serif}
button,input,select{font:inherit}
button{border:0;cursor:pointer}
button,input,select{min-width:0}
input,select{height:38px;border:1px solid #cfd8e3;border-radius:7px;padding:0 11px;outline:none;background:#fff}
input:focus{border-color:var(--primary);box-shadow:0 0 0 3px rgba(37,99,235,.13)}
.topbar{height:58px;display:flex;align-items:center;justify-content:space-between;padding:0 22px;background:#fff;border-bottom:1px solid var(--line)}
.brand{font-size:17px;font-weight:760}
.summary{font-size:13px;color:var(--muted)}
.nav{display:flex;gap:8px;padding:10px 14px;background:#fff;border-bottom:1px solid var(--line);overflow:auto}
.nav a{white-space:nowrap;text-decoration:none;color:#334155;background:#eef2f7;border-radius:7px;padding:8px 12px;font-size:13px;font-weight:720}
.nav a.active{background:var(--primary);color:#fff}
.layout{display:grid;grid-template-columns:minmax(760px,1fr) 390px;gap:14px;padding:14px;max-width:1500px;margin:0 auto}
.panel{background:var(--panel);border:1px solid var(--line);border-radius:8px;overflow:hidden}
.panel-head{min-height:48px;display:flex;align-items:center;justify-content:space-between;gap:10px;padding:8px 14px;border-bottom:1px solid var(--line);font-weight:720}
.tools{display:flex;gap:8px;align-items:center;flex-wrap:wrap}
.btn{min-height:32px;padding:0 10px;border-radius:7px;background:#e8eef8;color:#1e3a8a;font-weight:700}
.btn.primary{background:var(--primary);color:#fff}
.btn.danger{background:#fee4e2;color:#b42318}
.btn:disabled{opacity:.55;cursor:not-allowed}
.btn.mini{min-height:28px;padding:0 8px;font-size:12px}
.table-wrap{overflow:auto;max-height:calc(100vh - 104px)}
table{width:100%;border-collapse:collapse}
th,td{padding:10px 11px;border-bottom:1px solid var(--line);text-align:left;font-size:13px;vertical-align:top;white-space:nowrap}
th{position:sticky;top:0;background:#f8fafc;color:var(--muted);font-weight:720;z-index:1}
tr{cursor:pointer}
tr:hover{background:#f8fbff}
tr.active{background:#eef5ff}
.status{font-weight:760}
.status.queued{color:var(--warn)}
.status.running{color:var(--primary)}
.status.paused{color:var(--paused)}
.status.stopped,.status.failed{color:var(--bad)}
.status.succeeded{color:var(--ok)}
.muted{color:var(--muted)}
.message{max-width:320px;white-space:normal;color:var(--muted)}
.password{font-family:Consolas,"Cascadia Mono",monospace}
.log-box{height:calc(100vh - 188px);min-height:360px;overflow:auto;background:#0f172a;color:#dbeafe;padding:12px;font-family:Consolas,"Cascadia Mono",monospace;font-size:12px;line-height:1.55}
.log-line{padding:5px 0;border-bottom:1px solid rgba(148,163,184,.14);white-space:pre-wrap}
.log-time{color:#93c5fd}.log-info{color:#fbbf24}.log-success{color:#86efac}.log-error{color:#fb7185}
.detail{display:grid;gap:6px;padding:12px 14px;border-bottom:1px solid var(--line);font-size:13px}
.detail strong{font-weight:720}
.ops{display:grid;gap:14px;padding:14px}
.ops-panel{grid-column:1/-1}
.stats{display:grid;grid-template-columns:repeat(5,minmax(92px,1fr));gap:10px}
.stat{border:1px solid var(--line);border-radius:8px;padding:10px;background:#f8fafc}
.stat strong{display:block;font-size:20px;margin-bottom:3px}
.mini-form{display:grid;grid-template-columns:repeat(4,minmax(120px,1fr)) auto;gap:8px;padding:12px 14px;border-bottom:1px solid var(--line)}
.mini-form.license-form{grid-template-columns:repeat(5,minmax(120px,1fr)) auto}
.mini-form.logs{grid-template-columns:repeat(3,minmax(120px,1fr)) auto}
.delete-form{display:grid;gap:12px;padding:14px}
.subgrid{display:grid;grid-template-columns:1fr 1fr;gap:14px}
.mini-table{max-height:260px;overflow:auto}
.mini-table tr{cursor:default}
.auth{position:fixed;inset:0;display:grid;place-items:center;background:rgba(245,247,251,.92);z-index:10}
.auth-card{width:min(360px,calc(100vw - 32px));background:#fff;border:1px solid var(--line);border-radius:8px;padding:18px;box-shadow:0 18px 45px rgba(24,34,48,.12)}
.auth-card h1{font-size:18px;margin:0 0 12px}
.auth-card form{display:grid;gap:10px}
.auth-error{min-height:18px;color:var(--bad);font-size:13px}
.hidden{display:none!important}
@media (max-width:1100px){.layout{grid-template-columns:1fr}.log-box{height:360px}.table-wrap{max-height:none}.subgrid,.stats,.mini-form,.mini-form.license-form,.mini-form.logs{grid-template-columns:1fr}}
@media (max-width:680px){
  body{background:#fff}
  .topbar{height:auto;min-height:56px;align-items:flex-start;gap:8px;padding:10px 12px;flex-direction:column}
  .nav{padding:8px}
  .brand{font-size:16px}.summary{line-height:1.4}
  .layout{padding:8px;gap:10px}
  .panel{border-radius:8px}
  .panel-head{align-items:flex-start}
  .tools{width:100%}.tools .btn{flex:1;min-width:74px}
  input,select{height:42px}
  .table-wrap,.mini-table{max-height:none;overflow:visible}
  table,tbody,tr,td{display:block;width:100%}
  thead{display:none}
  td[data-label]{display:grid;grid-template-columns:82px minmax(0,1fr);gap:8px;padding:6px 0;border:0;white-space:normal}
  td[data-label]::before{content:attr(data-label);color:var(--muted);font-weight:650}
  tr[data-id],.mini-table tr{padding:10px 12px;border-bottom:1px solid var(--line)}
  tr.active{border-left:3px solid var(--primary)}
  .message{max-width:none}.password{word-break:break-all}
  .log-box{height:280px;min-height:280px}
  .ops{padding:10px;gap:10px}.stats{gap:8px}.mini-form{padding:10px}
}
</style>
</head>
<body>
<div class="auth" id="authPanel">
  <div class="auth-card">
    <h1>后台密钥</h1>
    <form id="loginForm">
      <input id="adminKey" type="password" autocomplete="current-password" placeholder="输入后台密钥">
      <button class="btn primary" type="submit">登录</button>
      <div class="auth-error" id="authError"></div>
    </form>
  </div>
</div>
<header class="topbar">
  <div class="brand">Yatori 管理后台</div>
  <div class="tools">
    <div class="summary" id="summary">加载中</div>
    <button class="btn" id="logoutBtn">退出</button>
  </div>
</header>
<nav class="nav">
  <a href="/admin" class="{{if eq .Page "tasks"}}active{{end}}">任务</a>
  <a href="/admin/users-page" class="{{if eq .Page "users"}}active{{end}}">用户</a>
  <a href="/admin/licenses-page" class="{{if eq .Page "licenses"}}active{{end}}">卡密</a>
  <a href="/admin/delete-tasks" class="{{if eq .Page "delete"}}active{{end}}">删除任务</a>
</nav>
<main class="layout">
  {{if eq .Page "tasks"}}
  <section class="panel">
    <div class="panel-head">
      <span>任务列表</span>
      <div class="tools"><button class="btn" id="refresh">刷新</button></div>
    </div>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>平台</th>
            <th>账号</th>
            <th>密码</th>
            <th>提交时间</th>
            <th>运行时间</th>
            <th>结束时间</th>
            <th>状态</th>
            <th>备注/日志摘要</th>
          </tr>
        </thead>
        <tbody id="tasks"><tr><td colspan="8" class="muted">加载中</td></tr></tbody>
      </table>
    </div>
  </section>
  <aside class="panel">
    <div class="panel-head">
      <span>日志</span>
      <div class="tools">
        <button class="btn primary" id="startBtn" disabled>启动</button>
        <button class="btn" id="pauseBtn" disabled>暂停</button>
        <button class="btn danger" id="stopBtn" disabled>停止</button>
        <button class="btn danger" id="clearLogsBtn" disabled>删除任务</button>
      </div>
    </div>
    <div class="detail" id="detail">
      <div class="muted">选择一个任务查看日志</div>
    </div>
    <div class="log-box" id="logs"><div class="log-line">等待任务日志...</div></div>
  </aside>
  {{end}}
  {{if eq .Page "users"}}
  <section class="panel ops-panel">
    <div class="panel-head">
      <span>用户管理</span>
      <div class="tools"><button class="btn" id="refreshOps">刷新</button></div>
    </div>
    <div class="ops">
      <div class="stats" id="stats"></div>
      <div class="panel">
        <div class="panel-head"><span>用户列表</span></div>
        <form class="mini-form" id="userForm">
          <input id="userId" type="hidden">
          <input id="userName" placeholder="用户名">
          <input id="userPassword" type="password" placeholder="密码/留空不改">
          <input id="userCreatedAt" placeholder="注册时间 2006-01-02">
          <select id="userDisabled"><option value="false">启用</option><option value="true">禁用</option></select>
          <button class="btn primary" type="submit">保存</button>
        </form>
        <div class="mini-table">
          <table>
            <thead><tr><th>用户名</th><th>注册时间</th><th>运行次数</th><th>状态</th><th>操作</th></tr></thead>
            <tbody id="users"><tr><td colspan="5" class="muted">加载中</td></tr></tbody>
          </table>
        </div>
      </div>
    </div>
  </section>
  {{end}}
  {{if eq .Page "licenses"}}
  <section class="panel ops-panel">
    <div class="panel-head">
      <span>卡密管理</span>
      <div class="tools"><button class="btn" id="refreshOps">刷新</button></div>
    </div>
    <div class="ops">
      <div class="panel">
        <div class="panel-head"><span>卡密</span></div>
        <form class="mini-form license-form" id="licenseForm">
          <input id="licenseKey" placeholder="留空自动生成">
          <input id="licensePrefix" placeholder="自定义前缀，默认 YATORI">
          <input id="licenseNote" placeholder="备注">
          <input id="licenseMaxUses" type="number" min="0" value="1" placeholder="可用次数，0不限">
          <select id="licenseActive"><option value="true">启用</option><option value="false">禁用</option></select>
          <button class="btn primary" type="submit">新增</button>
        </form>
        <div class="mini-table">
          <table>
            <thead><tr><th>卡密</th><th>备注</th><th>次数</th><th>最近用户</th><th>操作</th></tr></thead>
            <tbody id="licenses"><tr><td colspan="5" class="muted">加载中</td></tr></tbody>
          </table>
        </div>
      </div>
    </div>
  </section>
  {{end}}
  {{if eq .Page "delete"}}
  <section class="panel ops-panel">
    <div class="panel-head">
      <span>删除任务记录</span>
      <div class="tools"><button class="btn" id="refreshDeleteTasks">刷新</button></div>
    </div>
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>平台</th>
            <th>账号</th>
            <th>提交时间</th>
            <th>状态</th>
            <th>摘要</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody id="deleteTasks"><tr><td colspan="6" class="muted">加载中</td></tr></tbody>
      </table>
    </div>
    <div class="delete-form">
      <div class="muted">这里删除的是任务记录本身，包括状态和日志，不只是清空日志。</div>
    </div>
  </section>
  {{end}}
</main>
<script>
const $ = (id) => document.getElementById(id);
const page = "{{.Page}}";
const state = {tasks:[], selected:"", events:null};
const statusText = {queued:"排队中",running:"运行中",paused:"已暂停",stopped:"已停止",succeeded:"已完成",failed:"失败"};

function on(id, eventName, handler){
  const el = $(id);
  if(el) el.addEventListener(eventName, handler);
}

function escapeHTML(value){
  return String(value ?? "").replace(/[&<>"']/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;","'":"&#39;"}[c]));
}

function formatTime(value){
  if(!value) return "-";
  const date = new Date(value);
  if(Number.isNaN(date.getTime())) return "-";
  return date.toLocaleString();
}

function runtime(task){
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

async function postJSON(url, data){
  const res = await fetch(url,{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(data)});
  const text = await res.text();
  if(!res.ok) throw new Error(text || res.statusText);
  return text ? JSON.parse(text) : {};
}

function renderTasks(){
  const body = $("tasks");
  if(!body) return;
  if(!state.tasks.length){
    body.innerHTML = '<tr><td colspan="8" class="muted">暂无任务</td></tr>';
    return;
  }
  body.innerHTML = state.tasks.slice().reverse().map(t =>
    '<tr data-id="' + escapeHTML(t.id) + '" class="' + (state.selected === t.id ? "active" : "") + '">' +
      '<td data-label="平台">' + escapeHTML(t.platform) + '<br><span class="muted">' + escapeHTML(t.id) + '</span></td>' +
      '<td data-label="账号">' + escapeHTML(t.account) + '</td>' +
      '<td data-label="密码" class="password">' + escapeHTML(t.password || "-") + '</td>' +
      '<td data-label="提交时间">' + escapeHTML(formatTime(t.createdAt)) + '</td>' +
      '<td data-label="运行时间">' + escapeHTML(runtime(t)) + '</td>' +
      '<td data-label="结束时间">' + escapeHTML(formatTime(t.endedAt)) + '</td>' +
      '<td data-label="状态"><span class="status ' + escapeHTML(t.status) + '">' + escapeHTML(statusText[t.status] || t.status) + '</span></td>' +
      '<td data-label="摘要" class="message">' + escapeHTML(t.message || "") + '</td>' +
    '</tr>').join("");
  document.querySelectorAll("tr[data-id]").forEach(row => row.addEventListener("click", () => selectTask(row.dataset.id)));
}

function renderDetail(task){
  if(!$("detail")) return;
  if(!task){
    $("detail").innerHTML = '<div class="muted">选择一个任务查看日志</div>';
    setButtons(null);
    return;
  }
  $("detail").innerHTML =
    '<div><strong>' + escapeHTML(task.platform) + '</strong> · ' + escapeHTML(task.account) + '</div>' +
    '<div>提交：' + escapeHTML(formatTime(task.createdAt)) + '</div>' +
    '<div>运行：' + escapeHTML(runtime(task)) + ' · 结束：' + escapeHTML(formatTime(task.endedAt)) + '</div>' +
    '<div>状态：' + escapeHTML(statusText[task.status] || task.status) + '</div>';
  setButtons(task);
}

function setButtons(task){
  if(!$("startBtn")) return;
  $("startBtn").disabled = !task || !(task.status === "queued" || task.status === "paused");
  $("pauseBtn").disabled = !task || task.status !== "running";
  $("stopBtn").disabled = !task || ["succeeded","failed","stopped"].includes(task.status);
  $("clearLogsBtn").disabled = !task;
}

function appendLog(log){
  const box = $("logs");
  if(!box) return;
  const at = log.at ? new Date(log.at).toLocaleTimeString() : new Date().toLocaleTimeString();
  const level = escapeHTML(log.level || "info");
  const line = document.createElement("div");
  line.className = "log-line";
  line.innerHTML = '<span class="log-time">' + escapeHTML(at) + '</span> <span class="log-' + level + '">' + level + '</span> ' + escapeHTML(log.message || "");
  if(box.dataset.empty !== "false"){ box.innerHTML = ""; box.dataset.empty = "false"; }
  box.appendChild(line);
  box.scrollTop = box.scrollHeight;
}

async function loadTasks(){
  const res = await fetch("/admin/tasks");
  if(res.status === 401){
    showLogin();
    return;
  }
  if(!res.ok) throw new Error(await res.text() || res.statusText);
  state.tasks = await res.json();
  hideLogin();
  renderTasks();
  renderDeleteTasks();
  const running = state.tasks.filter(t => t.status === "running").length;
  if($("summary")) $("summary").textContent = "任务 " + state.tasks.length + " 个，运行中 " + running + " 个";
  renderDetail(state.tasks.find(t => t.id === state.selected));
}

async function loadOps(){
  const [statsRes, usersRes, licensesRes] = await Promise.all([
    fetch("/admin/stats"),
    fetch("/admin/users"),
    fetch("/admin/licenses")
  ]);
  if(statsRes.status === 401 || usersRes.status === 401 || licensesRes.status === 401){
    showLogin();
    return;
  }
  if(!statsRes.ok || !usersRes.ok || !licensesRes.ok) return;
  hideLogin();
  if($("stats")) renderStats(await statsRes.json());
  if($("users")) renderUsers(await usersRes.json());
  if($("licenses")) renderLicenses(await licensesRes.json());
}

function renderStats(stats){
  if(!$("stats")) return;
  $("stats").innerHTML =
    '<div class="stat"><strong>' + escapeHTML(stats.userCount || 0) + '</strong><span class="muted">用户数量</span></div>' +
    '<div class="stat"><strong>' + escapeHTML(stats.runCount || 0) + '</strong><span class="muted">运行次数</span></div>' +
    '<div class="stat"><strong>' + escapeHTML(stats.totalTasks || 0) + '</strong><span class="muted">总任务数</span></div>' +
    '<div class="stat"><strong>' + escapeHTML(stats.runningTasks || 0) + '</strong><span class="muted">运行中</span></div>' +
    '<div class="stat"><strong>' + escapeHTML(stats.totalLogs || 0) + '</strong><span class="muted">日志条数</span></div>';
}

function renderUsers(users){
  if(!$("users")) return;
  if(!users.length){
    $("users").innerHTML = '<tr><td colspan="5" class="muted">暂无用户</td></tr>';
    return;
  }
  $("users").innerHTML = users.map(u =>
    '<tr>' +
      '<td data-label="用户名">' + escapeHTML(u.username) + '<br><span class="muted">' + escapeHTML(u.id) + '</span></td>' +
      '<td data-label="注册时间">' + escapeHTML(formatTime(u.createdAt)) + '</td>' +
      '<td data-label="运行次数">' + escapeHTML(u.runCount || 0) + '</td>' +
      '<td data-label="状态">' + (u.disabled ? '<span class="status failed">禁用</span>' : '<span class="status succeeded">启用</span>') + '</td>' +
      '<td data-label="操作"><button class="btn mini" data-edit-user="' + escapeHTML(u.id) + '">编辑</button> <button class="btn mini danger" data-delete-user="' + escapeHTML(u.id) + '">删除</button></td>' +
    '</tr>').join("");
  document.querySelectorAll("[data-edit-user]").forEach(btn => btn.addEventListener("click", () => {
    const user = users.find(item => item.id === btn.dataset.editUser);
    if(!user) return;
    $("userId").value = user.id;
    $("userName").value = user.username;
    $("userPassword").value = "";
    $("userCreatedAt").value = user.createdAt ? new Date(user.createdAt).toISOString().slice(0,10) : "";
    $("userDisabled").value = user.disabled ? "true" : "false";
  }));
  document.querySelectorAll("[data-delete-user]").forEach(btn => btn.addEventListener("click", async () => {
    if(!confirm("确定删除该用户吗？")) return;
    await fetch("/admin/users/" + encodeURIComponent(btn.dataset.deleteUser), {method:"DELETE"});
    await loadOps();
  }));
}

function renderLicenses(items){
  if(!$("licenses")) return;
  if(!items.length){
    $("licenses").innerHTML = '<tr><td colspan="5" class="muted">暂无卡密</td></tr>';
    return;
  }
  $("licenses").innerHTML = items.map(item =>
    '<tr>' +
      '<td data-label="卡密" class="password">' + escapeHTML(item.key) + '<br>' + (item.active ? '<span class="status succeeded">启用</span>' : '<span class="status failed">禁用</span>') + '</td>' +
      '<td data-label="备注">' + escapeHTML(item.note || "-") + '</td>' +
      '<td data-label="次数">' + escapeHTML((item.uses || 0) + "/" + (item.maxUses > 0 ? item.maxUses : "不限") + "，剩余 " + (item.remaining >= 0 ? item.remaining : "不限")) + '</td>' +
      '<td data-label="最近用户">' + escapeHTML(item.usedBy || "-") + '</td>' +
      '<td data-label="操作"><button class="btn mini danger" data-delete-license="' + escapeHTML(item.key) + '">删除</button></td>' +
    '</tr>').join("");
  document.querySelectorAll("[data-delete-license]").forEach(btn => btn.addEventListener("click", async () => {
    if(!confirm("确定删除该卡密吗？")) return;
    await fetch("/admin/licenses/" + encodeURIComponent(btn.dataset.deleteLicense), {method:"DELETE"});
    await loadOps();
  }));
}

function renderDeleteTasks(){
  const body = $("deleteTasks");
  if(!body) return;
  if(!state.tasks.length){
    body.innerHTML = '<tr><td colspan="6" class="muted">暂无任务</td></tr>';
    return;
  }
  body.innerHTML = state.tasks.slice().reverse().map(t =>
    '<tr data-delete-id="' + escapeHTML(t.id) + '">' +
      '<td data-label="平台">' + escapeHTML(t.platform) + '<br><span class="muted">' + escapeHTML(t.id) + '</span></td>' +
      '<td data-label="账号">' + escapeHTML(t.account) + '</td>' +
      '<td data-label="提交时间">' + escapeHTML(formatTime(t.createdAt)) + '</td>' +
      '<td data-label="状态"><span class="status ' + escapeHTML(t.status) + '">' + escapeHTML(statusText[t.status] || t.status) + '</span></td>' +
      '<td data-label="摘要" class="message">' + escapeHTML(t.message || "") + '</td>' +
      '<td data-label="操作"><button class="btn mini danger" data-delete-task="' + escapeHTML(t.id) + '">删除</button></td>' +
    '</tr>').join("");
  document.querySelectorAll("[data-delete-task]").forEach(btn => btn.addEventListener("click", async (event) => {
    event.stopPropagation();
    await deleteTask(btn.dataset.deleteTask);
  }));
}

async function selectTask(id){
  if(!$("tasks")) return;
  state.selected = id;
  renderTasks();
  if(state.events){ state.events.close(); state.events = null; }
  const task = state.tasks.find(t => t.id === id);
  renderDetail(task);
  $("logs").innerHTML = "";
  $("logs").dataset.empty = "true";
  const logs = await fetch("/admin/tasks/" + encodeURIComponent(id) + "/logs").then(r=>r.json());
  logs.forEach(appendLog);
  state.events = new EventSource("/admin/tasks/" + encodeURIComponent(id) + "/events?after=" + logs.length);
  state.events.onmessage = (event) => appendLog(JSON.parse(event.data));
  state.events.addEventListener("done", () => {
    if(state.events){ state.events.close(); state.events = null; }
    loadTasks();
  });
}

async function control(action){
  if(!state.selected) return;
  try{
    await postJSON("/tasks/" + encodeURIComponent(state.selected) + "/control", {action});
    await loadTasks();
    await selectTask(state.selected);
  }catch(err){
    alert(err.message);
  }
}

function showLogin(){
  $("authPanel").classList.remove("hidden");
}

function hideLogin(){
  $("authPanel").classList.add("hidden");
  $("authError").textContent = "";
}

async function clearLogs(){
  if(!state.selected) return;
  if(!confirm("确定删除所选整条任务记录吗？")) return;
  try{
    const res = await fetch("/admin/tasks/" + encodeURIComponent(state.selected), {method:"DELETE"});
    const text = await res.text();
    if(!res.ok) throw new Error(text || res.statusText);
    if(state.events){ state.events.close(); state.events = null; }
    $("logs").innerHTML = '<div class="log-line">任务已删除</div>';
    $("logs").dataset.empty = "true";
    state.selected = "";
    await loadTasks();
  }catch(err){
    alert(err.message);
  }
}

async function login(key){
  const res = await fetch("/admin/login",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({key})});
  const text = await res.text();
  if(!res.ok) throw new Error(text || res.statusText);
  hideLogin();
  await loadCurrentPage();
}

async function logout(){
  await fetch("/admin/logout",{method:"POST"});
  if(state.events){ state.events.close(); state.events = null; }
  state.tasks = [];
  state.selected = "";
  renderTasks();
  showLogin();
}

async function loadCurrentPage(){
  if(page === "tasks" || page === "delete"){
    await loadTasks();
    return;
  }
  if(page === "users" || page === "licenses"){
    await loadOps();
  }
}

on("loginForm", "submit", async (event) => {
  event.preventDefault();
  try{
    await login($("adminKey").value);
  }catch(err){
    $("authError").textContent = err.message;
  }
});
on("refresh", "click", loadTasks);
on("refreshDeleteTasks", "click", loadTasks);
on("startBtn", "click", () => control("resume"));
on("pauseBtn", "click", () => control("pause"));
on("stopBtn", "click", () => control("stop"));
on("clearLogsBtn", "click", clearLogs);
on("logoutBtn", "click", logout);
on("refreshOps", "click", loadOps);
on("userForm", "submit", async (event) => {
  event.preventDefault();
  const id = $("userId").value;
  const payload = {
    username:$("userName").value.trim(),
    password:$("userPassword").value,
    createdAt:$("userCreatedAt").value.trim(),
    disabled:$("userDisabled").value === "true"
  };
  try{
    const url = id ? "/admin/users/" + encodeURIComponent(id) : "/admin/users";
    const res = await fetch(url,{method:id ? "PUT" : "POST",headers:{"Content-Type":"application/json"},body:JSON.stringify(payload)});
    const text = await res.text();
    if(!res.ok) throw new Error(text || res.statusText);
    $("userId").value = "";
    $("userName").value = "";
    $("userPassword").value = "";
    $("userCreatedAt").value = "";
    $("userDisabled").value = "false";
    await loadOps();
  }catch(err){
    alert(err.message);
  }
});
on("licenseForm", "submit", async (event) => {
  event.preventDefault();
  try{
    await postJSON("/admin/licenses", {
      key:$("licenseKey").value.trim(),
      prefix:$("licensePrefix").value.trim(),
      note:$("licenseNote").value.trim(),
      active:$("licenseActive").value === "true",
      maxUses:Number($("licenseMaxUses").value || 0)
    });
    $("licenseKey").value = "";
    $("licensePrefix").value = "";
    $("licenseNote").value = "";
    $("licenseMaxUses").value = "1";
    $("licenseActive").value = "true";
    await loadOps();
  }catch(err){
    alert(err.message);
  }
});
async function deleteTask(id){
  if(!id){
    alert("请选择要删除的任务");
    return;
  }
  if(!confirm("确定删除这整条任务记录吗？")) return;
  try{
    const res = await fetch("/admin/tasks/" + encodeURIComponent(id), {method:"DELETE"});
    const text = await res.text();
    if(!res.ok) throw new Error(text || res.statusText);
    await loadTasks();
  }catch(err){
    alert(err.message);
  }
}
loadCurrentPage();
if(page === "tasks"){
  setInterval(loadTasks, 3500);
}
</script>
</body>
</html>`
