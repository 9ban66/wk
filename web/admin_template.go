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
button,input{font:inherit}
button{border:0;cursor:pointer}
input{height:38px;border:1px solid #cfd8e3;border-radius:7px;padding:0 11px;outline:none}
input:focus{border-color:var(--primary);box-shadow:0 0 0 3px rgba(37,99,235,.13)}
.topbar{height:58px;display:flex;align-items:center;justify-content:space-between;padding:0 22px;background:#fff;border-bottom:1px solid var(--line)}
.brand{font-size:17px;font-weight:760}
.summary{font-size:13px;color:var(--muted)}
.layout{display:grid;grid-template-columns:minmax(760px,1fr) 390px;gap:14px;padding:14px;max-width:1500px;margin:0 auto}
.panel{background:var(--panel);border:1px solid var(--line);border-radius:8px;overflow:hidden}
.panel-head{height:48px;display:flex;align-items:center;justify-content:space-between;padding:0 14px;border-bottom:1px solid var(--line);font-weight:720}
.tools{display:flex;gap:8px;align-items:center}
.btn{min-height:32px;padding:0 10px;border-radius:7px;background:#e8eef8;color:#1e3a8a;font-weight:700}
.btn.primary{background:var(--primary);color:#fff}
.btn.danger{background:#fee4e2;color:#b42318}
.btn:disabled{opacity:.55;cursor:not-allowed}
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
.auth{position:fixed;inset:0;display:grid;place-items:center;background:rgba(245,247,251,.92);z-index:10}
.auth-card{width:min(360px,calc(100vw - 32px));background:#fff;border:1px solid var(--line);border-radius:8px;padding:18px;box-shadow:0 18px 45px rgba(24,34,48,.12)}
.auth-card h1{font-size:18px;margin:0 0 12px}
.auth-card form{display:grid;gap:10px}
.auth-error{min-height:18px;color:var(--bad);font-size:13px}
.hidden{display:none!important}
@media (max-width:1100px){.layout{grid-template-columns:1fr}.log-box{height:360px}.table-wrap{max-height:none}}
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
<main class="layout">
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
        <button class="btn danger" id="clearLogsBtn" disabled>删除日志</button>
      </div>
    </div>
    <div class="detail" id="detail">
      <div class="muted">选择一个任务查看日志</div>
    </div>
    <div class="log-box" id="logs"><div class="log-line">等待任务日志...</div></div>
  </aside>
</main>
<script>
const $ = (id) => document.getElementById(id);
const state = {tasks:[], selected:"", events:null};
const statusText = {queued:"排队中",running:"运行中",paused:"已暂停",stopped:"已停止",succeeded:"已完成",failed:"失败"};

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
  if(!state.tasks.length){
    body.innerHTML = '<tr><td colspan="8" class="muted">暂无任务</td></tr>';
    return;
  }
  body.innerHTML = state.tasks.slice().reverse().map(t =>
    '<tr data-id="' + escapeHTML(t.id) + '" class="' + (state.selected === t.id ? "active" : "") + '">' +
      '<td>' + escapeHTML(t.platform) + '<br><span class="muted">' + escapeHTML(t.id) + '</span></td>' +
      '<td>' + escapeHTML(t.account) + '</td>' +
      '<td class="password">' + escapeHTML(t.password || "-") + '</td>' +
      '<td>' + escapeHTML(formatTime(t.createdAt)) + '</td>' +
      '<td>' + escapeHTML(runtime(t)) + '</td>' +
      '<td>' + escapeHTML(formatTime(t.endedAt)) + '</td>' +
      '<td><span class="status ' + escapeHTML(t.status) + '">' + escapeHTML(statusText[t.status] || t.status) + '</span></td>' +
      '<td class="message">' + escapeHTML(t.message || "") + '</td>' +
    '</tr>').join("");
  document.querySelectorAll("tr[data-id]").forEach(row => row.addEventListener("click", () => selectTask(row.dataset.id)));
}

function renderDetail(task){
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
  $("startBtn").disabled = !task || !(task.status === "queued" || task.status === "paused");
  $("pauseBtn").disabled = !task || task.status !== "running";
  $("stopBtn").disabled = !task || ["succeeded","failed","stopped"].includes(task.status);
  $("clearLogsBtn").disabled = !task;
}

function appendLog(log){
  const box = $("logs");
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
  const running = state.tasks.filter(t => t.status === "running").length;
  $("summary").textContent = "任务 " + state.tasks.length + " 个，运行中 " + running + " 个";
  renderDetail(state.tasks.find(t => t.id === state.selected));
}

async function selectTask(id){
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
  if(!confirm("确定删除所选任务的日志吗？")) return;
  try{
    const res = await fetch("/admin/tasks/" + encodeURIComponent(state.selected) + "/logs", {method:"DELETE"});
    const text = await res.text();
    if(!res.ok) throw new Error(text || res.statusText);
    if(state.events){ state.events.close(); state.events = null; }
    $("logs").innerHTML = '<div class="log-line">日志已删除</div>';
    $("logs").dataset.empty = "true";
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
  await loadTasks();
}

async function logout(){
  await fetch("/admin/logout",{method:"POST"});
  if(state.events){ state.events.close(); state.events = null; }
  state.tasks = [];
  state.selected = "";
  renderTasks();
  showLogin();
}

$("loginForm").addEventListener("submit", async (event) => {
  event.preventDefault();
  try{
    await login($("adminKey").value);
  }catch(err){
    $("authError").textContent = err.message;
  }
});
$("refresh").addEventListener("click", loadTasks);
$("startBtn").addEventListener("click", () => control("resume"));
$("pauseBtn").addEventListener("click", () => control("pause"));
$("stopBtn").addEventListener("click", () => control("stop"));
$("clearLogsBtn").addEventListener("click", clearLogs);
$("logoutBtn").addEventListener("click", logout);
loadTasks();
setInterval(loadTasks, 3500);
</script>
</body>
</html>`
