package web

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	haiqikejiAggregation "github.com/yatori-dev/yatori-go-core/aggregation/haiqikeji"
	xuexitongAggregation "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	xuexitongPoint "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong/point"
	yinghuaAggregation "github.com/yatori-dev/yatori-go-core/aggregation/yinghua"
	haiqikejiApi "github.com/yatori-dev/yatori-go-core/api/haiqikeji"
	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	yinghuaApi "github.com/yatori-dev/yatori-go-core/api/yinghua"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
)

type Server struct {
	store    *taskStore
	accounts *accountStore
}

const adminCookieName = "yatori_admin_key"
const userCookieName = "yatori_user_session"

type taskRequest struct {
	Platform   string   `json:"platform"`
	Account    string   `json:"account"`
	Password   string   `json:"password"`
	PreURL     string   `json:"preUrl"`
	CourseIDs  []string `json:"courseIds"`
	AIURL      string   `json:"aiUrl"`
	AIModel    string   `json:"aiModel"`
	AIKey      string   `json:"aiKey"`
	AIType     string   `json:"aiType"`
	Message    string   `json:"message"`
	LicenseKey string   `json:"licenseKey"`
}

type CourseOption struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Meta     string `json:"meta,omitempty"`
	Progress string `json:"progress,omitempty"`
	Ended    bool   `json:"ended"`
}

type TaskSummary struct {
	ID             string     `json:"id"`
	UserID         string     `json:"userId,omitempty"`
	Username       string     `json:"username,omitempty"`
	LicenseKey     string     `json:"licenseKey,omitempty"`
	Platform       string     `json:"platform"`
	Account        string     `json:"account"`
	CourseIDs      []string   `json:"courseIds"`
	Status         TaskStatus `json:"status"`
	Message        string     `json:"message"`
	CreatedAt      time.Time  `json:"createdAt"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	EndedAt        *time.Time `json:"endedAt,omitempty"`
	RuntimeSeconds int64      `json:"runtimeSeconds"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func NewServer() *Server {
	return &Server{store: newTaskStore(), accounts: newAccountStore()}
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.indexHandler)
	mux.HandleFunc("/auth/register", s.authRegisterHandler)
	mux.HandleFunc("/auth/login", s.authLoginHandler)
	mux.HandleFunc("/auth/logout", s.authLogoutHandler)
	mux.HandleFunc("/auth/me", s.authMeHandler)
	mux.HandleFunc("/license/verify", s.licenseVerifyHandler)
	mux.HandleFunc("/admin", s.adminHandler)
	mux.HandleFunc("/admin/users-page", s.adminPageHandler("users"))
	mux.HandleFunc("/admin/licenses-page", s.adminPageHandler("licenses"))
	mux.HandleFunc("/admin/delete-tasks", s.adminPageHandler("delete"))
	mux.HandleFunc("/admin/login", s.adminLoginHandler)
	mux.HandleFunc("/admin/logout", s.adminLogoutHandler)
	mux.HandleFunc("/admin/stats", s.adminStatsHandler)
	mux.HandleFunc("/admin/users", s.adminUsersHandler)
	mux.HandleFunc("/admin/users/", s.adminUserHandler)
	mux.HandleFunc("/admin/licenses", s.adminLicensesHandler)
	mux.HandleFunc("/admin/licenses/", s.adminLicenseHandler)
	mux.HandleFunc("/admin/logs", s.adminLogsHandler)
	mux.HandleFunc("/admin/tasks", s.adminTasksHandler)
	mux.HandleFunc("/admin/tasks/", s.adminTaskHandler)
	mux.HandleFunc("/courses", s.coursesHandler)
	mux.HandleFunc("/submit", s.submitHandler)
	mux.HandleFunc("/task-query", s.taskQueryHandler)
	mux.HandleFunc("/tasks", s.tasksHandler)
	mux.HandleFunc("/tasks/", s.taskHandler)
	log.Printf("web server listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func adminKey() string {
	if key := strings.TrimSpace(os.Getenv("ADMIN_KEY")); key != "" {
		return key
	}
	return "yatori-admin"
}

func (s *Server) currentUser(r *http.Request) (UserSummary, bool) {
	cookie, err := r.Cookie(userCookieName)
	if err != nil {
		return UserSummary{}, false
	}
	return s.accounts.userBySession(strings.TrimSpace(cookie.Value))
}

func (s *Server) requireUser(w http.ResponseWriter, r *http.Request) (UserSummary, bool) {
	user, ok := s.currentUser(r)
	if ok {
		return user, true
	}
	http.Error(w, "请先登录", http.StatusUnauthorized)
	return UserSummary{}, false
}

func (s *Server) authRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := s.accounts.register(payload.Username, payload.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "user": user})
}

func (s *Server) authLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	token, user, err := s.accounts.login(payload.Username, payload.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     userCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 30,
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "user": user})
}

func (s *Server) authLogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if cookie, err := r.Cookie(userCookieName); err == nil {
		s.accounts.logout(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     userCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) authMeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "user": user})
}

func (s *Server) licenseVerifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	var payload struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key, err := s.accounts.verifyLicense(user.ID, payload.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "license": key})
}

func (s *Server) isAdminRequest(r *http.Request) bool {
	key := strings.TrimSpace(r.Header.Get("X-Admin-Key"))
	if key == "" {
		key = strings.TrimSpace(r.URL.Query().Get("key"))
	}
	if key == "" {
		if cookie, err := r.Cookie(adminCookieName); err == nil {
			key = strings.TrimSpace(cookie.Value)
		}
	}
	expected := adminKey()
	if key == "" || expected == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(key), []byte(expected)) == 1
}

func (s *Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	if s.isAdminRequest(r) {
		return true
	}
	http.Error(w, "需要后台密钥", http.StatusUnauthorized)
	return false
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl := newIndexTemplate()
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/admin" {
		http.NotFound(w, r)
		return
	}
	s.renderAdminPage(w, "tasks")
}

func (s *Server) adminPageHandler(page string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.renderAdminPage(w, page)
	}
}

func (s *Server) renderAdminPage(w http.ResponseWriter, page string) {
	tmpl := newAdminTemplate()
	if err := tmpl.Execute(w, map[string]string{"Page": page}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	key := strings.TrimSpace(payload.Key)
	if key == "" || subtle.ConstantTimeCompare([]byte(key), []byte(adminKey())) != 1 {
		http.Error(w, "密钥错误", http.StatusUnauthorized)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    key,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7,
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) adminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     adminCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) adminStatsHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	stats := AppStats{
		UserCount:    len(s.accounts.listUsers()),
		RunCount:     s.accounts.totalRunCount(),
		TotalTasks:   s.store.count(),
		RunningTasks: s.store.runningCount(),
		TotalLogs:    s.store.logCount(),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(stats)
}

func (s *Server) adminUsersHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s.accounts.listUsers())
	case http.MethodPost:
		var payload struct {
			Username  string `json:"username"`
			Password  string `json:"password"`
			Disabled  *bool  `json:"disabled"`
			CreatedAt string `json:"createdAt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		createdAt, err := parseOptionalTime(payload.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		user, err := s.accounts.upsertUser("", payload.Username, payload.Password, payload.Disabled, createdAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "user": user})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) adminUserHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/admin/users/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodPut:
		var payload struct {
			Username  string `json:"username"`
			Password  string `json:"password"`
			Disabled  *bool  `json:"disabled"`
			CreatedAt string `json:"createdAt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		createdAt, err := parseOptionalTime(payload.CreatedAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		user, err := s.accounts.upsertUser(id, payload.Username, payload.Password, payload.Disabled, createdAt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "user": user})
	case http.MethodDelete:
		if !s.accounts.deleteUser(id) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) adminLicensesHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(s.accounts.listLicenses())
	case http.MethodPost:
		var payload struct {
			Key     string `json:"key"`
			Note    string `json:"note"`
			Active  *bool  `json:"active"`
			MaxUses int    `json:"maxUses"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		active := true
		if payload.Active != nil {
			active = *payload.Active
		}
		key, err := s.accounts.addLicense(LicenseKey{Key: payload.Key, Note: payload.Note, Active: active, MaxUses: payload.MaxUses})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "license": key})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) adminLicenseHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	key := strings.TrimPrefix(r.URL.Path, "/admin/licenses/")
	if key == "" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.accounts.deleteLicense(key) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *Server) adminLogsHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	before, err := parseOptionalTime(r.URL.Query().Get("before"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	removed := s.store.clearLogsByFilter(
		strings.TrimSpace(r.URL.Query().Get("platform")),
		strings.TrimSpace(r.URL.Query().Get("account")),
		before,
	)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "removed": removed})
}

func (s *Server) adminTasksHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.store.list())
}

func (s *Server) adminTaskHandler(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/admin/tasks/")
	if strings.HasSuffix(path, "/logs") {
		id := strings.TrimSuffix(path, "/logs")
		switch r.Method {
		case http.MethodGet:
			s.taskLogsHandler(w, r, id)
		case http.MethodDelete:
			if !s.store.clearLogs(id) {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	if strings.HasSuffix(path, "/events") {
		s.taskEventsHandler(w, r, strings.TrimSuffix(path, "/events"))
		return
	}
	if path != "" && r.Method == http.MethodDelete {
		if !s.store.delete(path) {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
		return
	}
	http.NotFound(w, r)
}

func (s *Server) submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	req, err := parseTaskRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if _, err := s.accounts.verifyLicense(user.ID, req.LicenseKey); err != nil {
		http.Error(w, "卡密验证失败: "+err.Error(), http.StatusForbidden)
		return
	}
	if req.Platform == "" || req.Account == "" || req.Password == "" {
		http.Error(w, "平台、账号和密码不能为空", http.StatusBadRequest)
		return
	}
	if req.Platform != "xuexitong" && req.PreURL == "" {
		http.Error(w, "海奇科技和英华需要填写平台地址", http.StatusBadRequest)
		return
	}
	task := s.store.enqueue(Task{
		UserID:     user.ID,
		Username:   user.Username,
		LicenseKey: req.LicenseKey,
		Platform:   req.Platform,
		Account:    req.Account,
		Password:   req.Password,
		PreURL:     req.PreURL,
		CourseIDs:  req.CourseIDs,
		AIURL:      req.AIURL,
		AIModel:    req.AIModel,
		AIKey:      req.AIKey,
		AIType:     req.AIType,
		Message:    req.Message,
	})
	if _, err := s.accounts.consumeLicense(user.ID, req.LicenseKey); err != nil {
		s.store.delete(task.ID)
		http.Error(w, "卡密扣次失败: "+err.Error(), http.StatusForbidden)
		return
	}
	s.accounts.recordRun(user.ID)
	s.store.appendLog(task.ID, "info", "任务已进入队列")
	go s.runTask(task.ID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "task": task})
}

func parseTaskRequest(r *http.Request) (taskRequest, error) {
	var req taskRequest
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return taskRequest{}, err
		}
	} else {
		if err := r.ParseForm(); err != nil {
			return taskRequest{}, err
		}
		req = taskRequest{
			Platform:   r.Form.Get("platform"),
			Account:    r.Form.Get("account"),
			Password:   r.Form.Get("password"),
			PreURL:     r.Form.Get("preUrl"),
			CourseIDs:  r.Form["courseIds"],
			AIURL:      r.Form.Get("aiUrl"),
			AIModel:    r.Form.Get("aiModel"),
			AIKey:      r.Form.Get("aiKey"),
			AIType:     r.Form.Get("aiType"),
			Message:    r.Form.Get("message"),
			LicenseKey: r.Form.Get("licenseKey"),
		}
		if len(req.CourseIDs) == 0 {
			req.CourseIDs = r.Form["courseId"]
		}
	}
	req.Platform = strings.TrimSpace(req.Platform)
	req.Account = strings.TrimSpace(req.Account)
	req.Password = strings.TrimSpace(req.Password)
	req.PreURL = strings.TrimSpace(req.PreURL)
	req.AIURL = strings.TrimSpace(req.AIURL)
	req.AIModel = strings.TrimSpace(req.AIModel)
	req.AIKey = strings.TrimSpace(req.AIKey)
	req.AIType = strings.TrimSpace(req.AIType)
	req.Message = strings.TrimSpace(req.Message)
	req.LicenseKey = strings.TrimSpace(req.LicenseKey)
	req.CourseIDs = cleanStringList(req.CourseIDs)
	return req, nil
}

func cleanStringList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func (s *Server) coursesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req, err := parseTaskRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Platform == "" || req.Account == "" || req.Password == "" {
		http.Error(w, "平台、账号和密码不能为空", http.StatusBadRequest)
		return
	}
	if req.Platform != "xuexitong" && req.PreURL == "" {
		http.Error(w, "海奇科技和英华需要填写平台地址", http.StatusBadRequest)
		return
	}
	courses, err := fetchCourses(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(courses)
}

func fetchCourses(req taskRequest) ([]CourseOption, error) {
	switch req.Platform {
	case "haiqikeji":
		cache := &haiqikejiApi.HqkjUserCache{PreUrl: req.PreURL, Account: req.Account, Password: req.Password}
		if err := haiqikejiAggregation.HqkjLoginAction(cache); err != nil {
			return nil, fmt.Errorf("海奇科技登录失败: %w", err)
		}
		courses, err := haiqikejiAggregation.HqkjCourseListAction(cache)
		if err != nil {
			return nil, fmt.Errorf("海奇科技拉取课程失败: %w", err)
		}
		options := make([]CourseOption, 0, len(courses))
		for _, course := range courses {
			meta := strings.TrimSpace(strings.Join(nonEmptyStrings(course.LecturerName, course.PeriodName, formatEndDate(course.EndDate), course.CollegeId), " · "))
			ended := courseEnded(course.EndDate)
			progress := "进行中"
			if ended {
				progress = "已结束"
			}
			options = append(options, CourseOption{ID: course.Id, Name: course.Name, Meta: meta, Progress: progress, Ended: ended})
		}
		return options, nil
	case "yinghua":
		cache := &yinghuaApi.YingHuaUserCache{PreUrl: req.PreURL, Account: req.Account, Password: req.Password}
		if err := yinghuaAggregation.YingHuaLoginAction(cache); err != nil {
			return nil, fmt.Errorf("英华登录失败: %w", err)
		}
		courses, err := yinghuaAggregation.CourseListAction(cache)
		if err != nil {
			return nil, fmt.Errorf("英华拉取课程失败: %w", err)
		}
		options := make([]CourseOption, 0, len(courses))
		for _, course := range courses {
			ended := courseEnded(course.EndDate)
			meta := strings.TrimSpace(strings.Join(nonEmptyStrings(fmt.Sprintf("%d/%d 视频", course.VideoLearned, course.VideoCount), formatEndDate(course.EndDate)), " · "))
			progress := fmt.Sprintf("%.0f%%", course.Progress)
			if ended {
				progress += " · 已结束"
			}
			options = append(options, CourseOption{ID: course.Id, Name: course.Name, Meta: meta, Progress: progress, Ended: ended})
		}
		return options, nil
	case "xuexitong":
		cache := &xuexitongApi.XueXiTUserCache{Name: req.Account, Password: req.Password}
		if err := xuexitongAggregation.XueXiTLoginAction(cache); err != nil {
			return nil, fmt.Errorf("学习通登录失败: %w", err)
		}
		courses, err := xuexitongAggregation.XueXiTPullCourseAction(cache)
		if err != nil {
			return nil, fmt.Errorf("学习通拉取课程失败: %w", err)
		}
		options := make([]CourseOption, 0, len(courses))
		for _, course := range courses {
			meta := strings.TrimSpace(strings.Join(nonEmptyStrings(course.CourseTeacher, "class "+course.Key), " · "))
			ended := course.State == 1
			progress := fmt.Sprintf("%d/%d", course.JobFinishCount, course.JobCount)
			if course.State == 1 {
				progress += " · 已结束"
			}
			options = append(options, CourseOption{ID: course.CourseID, Name: course.CourseName, Meta: meta, Progress: progress, Ended: ended})
		}
		return options, nil
	default:
		return nil, fmt.Errorf("unsupported platform %s", req.Platform)
	}
}

func nonEmptyStrings(values ...string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func formatEndDate(end time.Time) string {
	if end.IsZero() {
		return ""
	}
	return "结束 " + end.Format("2006-01-02")
}

func courseEnded(end time.Time) bool {
	if end.IsZero() {
		return false
	}
	return time.Now().After(end.AddDate(0, 0, 1))
}

func parseOptionalTime(value string) (*time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.ParseInLocation(layout, value, time.Local)
		if err == nil {
			return &parsed, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("时间格式错误: %w", lastErr)
}

func (s *Server) tasksHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summarizeTasks(s.store.listForUser(user.ID)))
}

func (s *Server) taskQueryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	platform := strings.TrimSpace(r.URL.Query().Get("platform"))
	account := strings.TrimSpace(r.URL.Query().Get("account"))
	if platform == "" || account == "" {
		http.Error(w, "平台和账号不能为空", http.StatusBadRequest)
		return
	}
	user, ok := s.requireUser(w, r)
	if !ok {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summarizeTasks(s.store.findForUser(user.ID, platform, account)))
}

func summarizeTask(task Task) TaskSummary {
	return TaskSummary{
		ID:             task.ID,
		UserID:         task.UserID,
		Username:       task.Username,
		LicenseKey:     task.LicenseKey,
		Platform:       task.Platform,
		Account:        task.Account,
		CourseIDs:      task.CourseIDs,
		Status:         task.Status,
		Message:        task.Message,
		CreatedAt:      task.CreatedAt,
		StartedAt:      task.StartedAt,
		EndedAt:        task.EndedAt,
		RuntimeSeconds: task.RuntimeSeconds,
		UpdatedAt:      task.UpdatedAt,
	}
}

func summarizeTasks(tasks []Task) []TaskSummary {
	summaries := make([]TaskSummary, 0, len(tasks))
	for _, task := range tasks {
		summaries = append(summaries, summarizeTask(task))
	}
	return summaries
}

func (s *Server) taskHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if strings.HasSuffix(path, "/logs") {
		id := strings.TrimSuffix(path, "/logs")
		if !s.canAccessTask(w, r, id) {
			return
		}
		s.taskLogsHandler(w, r, id)
		return
	}
	if strings.HasSuffix(path, "/events") {
		id := strings.TrimSuffix(path, "/events")
		if !s.canAccessTask(w, r, id) {
			return
		}
		s.taskEventsHandler(w, r, id)
		return
	}
	if strings.HasSuffix(path, "/control") {
		id := strings.TrimSuffix(path, "/control")
		if !s.canAccessTask(w, r, id) {
			return
		}
		s.taskControlHandler(w, r, id)
		return
	}
	id := path
	if id == "" {
		http.NotFound(w, r)
		return
	}
	task, ok := s.store.get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if !s.canAccessTask(w, r, id) {
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(summarizeTask(task))
}

func (s *Server) canAccessTask(w http.ResponseWriter, r *http.Request, id string) bool {
	task, ok := s.store.get(id)
	if !ok {
		http.NotFound(w, r)
		return false
	}
	if s.isAdminRequest(r) {
		return true
	}
	user, ok := s.requireUser(w, r)
	if !ok {
		return false
	}
	if task.UserID == "" || task.UserID == user.ID {
		return true
	}
	http.Error(w, "无权访问该任务", http.StatusForbidden)
	return false
}

func (s *Server) taskControlHandler(w http.ResponseWriter, r *http.Request, id string) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	payload.Action = strings.TrimSpace(payload.Action)
	var ok bool
	switch payload.Action {
	case "start", "resume":
		ok = s.store.start(id)
		if ok {
			s.store.appendLogOnly(id, "info", "任务已启动/继续")
		}
	case "pause":
		ok = s.store.pause(id)
		if ok {
			s.store.appendLogOnly(id, "info", "任务已请求暂停")
		}
	case "stop":
		ok = s.store.stop(id)
		if ok {
			s.store.appendLogOnly(id, "info", "任务已请求停止")
		}
	default:
		http.Error(w, "unsupported action", http.StatusBadRequest)
		return
	}
	if !ok {
		http.NotFound(w, r)
		return
	}
	task, _ := s.store.get(id)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (s *Server) taskLogsHandler(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := s.store.get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task.Logs)
}

func (s *Server) taskEventsHandler(w http.ResponseWriter, r *http.Request, id string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	if _, ok := s.store.get(id); !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	last := 0
	if after := r.URL.Query().Get("after"); after != "" {
		if n, err := strconv.Atoi(after); err == nil && n > 0 {
			last = n
		}
	}
	if lastEventID := r.Header.Get("Last-Event-ID"); lastEventID != "" {
		if n, err := strconv.Atoi(lastEventID); err == nil && n >= last {
			last = n + 1
		}
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			task, ok := s.store.get(id)
			if !ok {
				return
			}
			for last < len(task.Logs) {
				payload, _ := json.Marshal(task.Logs[last])
				fmt.Fprintf(w, "id: %d\ndata: %s\n\n", last, payload)
				last++
			}
			if task.Status == TaskSucceeded || task.Status == TaskFailed || task.Status == TaskStopped {
				fmt.Fprint(w, "event: done\ndata: {}\n\n")
				flusher.Flush()
				return
			}
			flusher.Flush()
		}
	}
}

func (s *Server) runTask(id string) {
	if !s.store.start(id) {
		return
	}
	s.logTask(id, "info", "开始执行")
	defer func() {
		s.store.appendLogOnly(id, "info", "执行结束")
	}()

	task, ok := s.store.get(id)
	if !ok {
		return
	}
	var err error
	switch task.Platform {
	case "haiqikeji":
		err = s.runHaiqiTask(id, task)
	case "yinghua":
		err = s.runYinghuaTask(id, task)
	case "xuexitong":
		err = s.runXueXiTongTask(id, task)
	default:
		err = fmt.Errorf("unsupported platform %s", task.Platform)
	}
	if err != nil {
		if isTaskStopped(err) {
			s.logTask(id, "info", "任务已停止")
			return
		}
		s.logTask(id, "error", err.Error())
		s.store.finish(id, TaskFailed, err.Error())
		return
	}
	s.logTask(id, "success", "任务已完成")
	s.store.finish(id, TaskSucceeded, "任务已完成")
}

func (s *Server) logTask(id, level, message string) {
	s.store.appendLog(id, level, message)
}

func isTaskStopped(err error) bool {
	return err != nil && err.Error() == "task stopped"
}

func (s *Server) waitIfPausedOrStopped(id string) error {
	for {
		task, ok := s.store.get(id)
		if !ok {
			return fmt.Errorf("task stopped")
		}
		switch task.Status {
		case TaskStopped:
			return fmt.Errorf("task stopped")
		case TaskPaused:
			time.Sleep(500 * time.Millisecond)
			continue
		default:
			return nil
		}
	}
}

func (s *Server) interruptibleSleep(id string, d time.Duration) error {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if err := s.waitIfPausedOrStopped(id); err != nil {
			return err
		}
		remaining := time.Until(deadline)
		if remaining > 500*time.Millisecond {
			remaining = 500 * time.Millisecond
		}
		time.Sleep(remaining)
	}
	return s.waitIfPausedOrStopped(id)
}

func (s *Server) runHaiqiTask(id string, task Task) error {
	if err := s.waitIfPausedOrStopped(id); err != nil {
		return err
	}
	cache := &haiqikejiApi.HqkjUserCache{PreUrl: task.PreURL, Account: task.Account, Password: task.Password}
	s.logTask(id, "info", "海奇科技登录中")
	if err := haiqikejiAggregation.HqkjLoginAction(cache); err != nil {
		return fmt.Errorf("海奇科技登录失败: %w", err)
	}
	s.logTask(id, "success", "海奇科技登录成功")
	courses, err := haiqikejiAggregation.HqkjCourseListAction(cache)
	if err != nil {
		return fmt.Errorf("海奇科技拉取课程失败: %w", err)
	}
	courses = filterHaiqiCourses(courses, task.CourseIDs)
	if len(courses) == 0 {
		return fmt.Errorf("海奇科技未找到匹配课程")
	}
	s.logTask(id, "info", fmt.Sprintf("准备执行 %d 门课程", len(courses)))
	processed := 0
	for _, course := range courses {
		if err := s.waitIfPausedOrStopped(id); err != nil {
			return err
		}
		if courseEnded(course.EndDate) {
			s.logTask(id, "info", fmt.Sprintf("跳过已结束课程: %s (%s)", course.Name, course.EndDate.Format("2006-01-02")))
			continue
		}
		s.logTask(id, "info", "拉取课程节点: "+course.Name)
		nodes, err := haiqikejiAggregation.HqkjNodeListAction(cache, course)
		if err != nil {
			return fmt.Errorf("海奇科技拉取节点失败: %w", err)
		}
		s.logTask(id, "info", fmt.Sprintf("%s 共 %d 个节点", course.Name, len(nodes)))
		for _, node := range nodes {
			if err := s.waitIfPausedOrStopped(id); err != nil {
				return err
			}
			if node.TabVideo != 1 && node.TabVideo != 0 {
				continue
			}
			s.logTask(id, "info", "提交学习进度: "+node.Name)
			_, err := haiqikejiAggregation.HqkjSubmitFastStudyTimeAction(cache, node)
			if err != nil {
				return fmt.Errorf("海奇科技刷课失败: %w", err)
			}
			processed++
			if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
				return err
			}
		}
	}
	if processed == 0 {
		return fmt.Errorf("海奇科技未找到可执行课程或节点")
	}
	return nil
}

func filterHaiqiCourses(courses []haiqikejiAggregation.HqkjCourse, ids []string) []haiqikejiAggregation.HqkjCourse {
	if len(ids) == 0 {
		return courses
	}
	selected := stringSet(ids)
	out := make([]haiqikejiAggregation.HqkjCourse, 0, len(courses))
	for _, course := range courses {
		if _, ok := selected[course.Id]; ok {
			out = append(out, course)
		}
	}
	return out
}

func (s *Server) runYinghuaTask(id string, task Task) error {
	if err := s.waitIfPausedOrStopped(id); err != nil {
		return err
	}
	cache := &yinghuaApi.YingHuaUserCache{PreUrl: task.PreURL, Account: task.Account, Password: task.Password}
	s.logTask(id, "info", "英华登录中")
	if err := yinghuaAggregation.YingHuaLoginAction(cache); err != nil {
		return fmt.Errorf("英华登录失败: %w", err)
	}
	s.logTask(id, "success", "英华登录成功")
	courses, err := yinghuaAggregation.CourseListAction(cache)
	if err != nil {
		return fmt.Errorf("英华拉取课程失败: %w", err)
	}
	courses = filterYinghuaCourses(courses, task.CourseIDs)
	if len(courses) == 0 {
		return fmt.Errorf("英华未找到匹配课程")
	}
	s.logTask(id, "info", fmt.Sprintf("准备执行 %d 门课程", len(courses)))
	processed := 0
	for _, course := range courses {
		if err := s.waitIfPausedOrStopped(id); err != nil {
			return err
		}
		if courseEnded(course.EndDate) {
			s.logTask(id, "info", fmt.Sprintf("跳过已结束课程: %s (%s)", course.Name, course.EndDate.Format("2006-01-02")))
			continue
		}
		s.logTask(id, "info", "拉取课程节点: "+course.Name)
		nodes, err := yinghuaAggregation.VideosListAction(cache, course)
		if err != nil {
			return fmt.Errorf("英华拉取节点失败: %w", err)
		}
		for _, node := range nodes {
			if err := s.waitIfPausedOrStopped(id); err != nil {
				return err
			}
			if !node.TabVideo {
				continue
			}
			s.logTask(id, "info", "提交学习进度: "+node.Name)
			_, err := yinghuaAggregation.SubmitStudyTimeAction(cache, node.Id, node.Id, 60)
			if err != nil {
				return fmt.Errorf("英华刷课失败: %w", err)
			}
			processed++
			if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
				return err
			}
		}
	}
	if processed == 0 {
		return fmt.Errorf("英华未找到可执行课程或节点")
	}
	return nil
}

func filterYinghuaCourses(courses []yinghuaAggregation.YingHuaCourse, ids []string) []yinghuaAggregation.YingHuaCourse {
	if len(ids) == 0 {
		return courses
	}
	selected := stringSet(ids)
	out := make([]yinghuaAggregation.YingHuaCourse, 0, len(courses))
	for _, course := range courses {
		if _, ok := selected[course.Id]; ok {
			out = append(out, course)
		}
	}
	return out
}

func (s *Server) runXueXiTongTask(id string, task Task) error {
	if err := s.waitIfPausedOrStopped(id); err != nil {
		return err
	}
	cache := &xuexitongApi.XueXiTUserCache{Name: task.Account, Password: task.Password}
	s.logTask(id, "info", "学习通登录中")
	if err := xuexitongAggregation.XueXiTLoginAction(cache); err != nil {
		return fmt.Errorf("学习通登录失败: %w", err)
	}
	s.logTask(id, "success", "学习通登录成功")
	courses, err := xuexitongAggregation.XueXiTPullCourseAction(cache)
	if err != nil {
		return fmt.Errorf("学习通拉取课程失败: %w", err)
	}
	courses = filterXueXiTongCourses(courses, task.CourseIDs)
	if len(courses) == 0 {
		return fmt.Errorf("学习通未找到匹配课程")
	}

	processed := 0
	discovered := 0
	scanErrors := 0
	unfinishedHints := 0
	s.logTask(id, "info", fmt.Sprintf("准备执行 %d 门课程", len(courses)))
	for _, course := range courses {
		if err := s.waitIfPausedOrStopped(id); err != nil {
			return err
		}
		if course.State == 1 {
			s.logTask(id, "info", "跳过已结束课程: "+course.CourseName)
			continue
		}
		if !course.IsStart {
			s.logTask(id, "info", "跳过未开课课程: "+course.CourseName)
			continue
		}
		if course.JobCount > course.JobFinishCount {
			unfinishedHints++
		}
		s.logTask(id, "info", fmt.Sprintf("处理课程: %s (列表进度 %d/%d，仅供参考)", course.CourseName, course.JobFinishCount, course.JobCount))

		classID, err := strconv.Atoi(course.Key)
		if err != nil {
			scanErrors++
			s.logTask(id, "error", fmt.Sprintf("课程 classId 解析失败: %s (%v)", course.CourseName, err))
			continue
		}
		chapter, ok, err := xuexitongAggregation.PullCourseChapterAction(cache, course.Cpi, classID)
		if err != nil || !ok {
			scanErrors++
			s.logTask(id, "error", fmt.Sprintf("无法拉取课程章节: %s (%v)", course.CourseName, err))
			continue
		}
		s.logTask(id, "info", fmt.Sprintf("章节拉取成功: %s，共 %d 个章节", course.CourseName, len(chapter.Knowledge)))

		nodes := make([]int, 0, len(chapter.Knowledge))
		for _, item := range chapter.Knowledge {
			nodes = append(nodes, item.ID)
		}
		if len(nodes) == 0 {
			s.logTask(id, "info", "跳过无章节课程: "+course.CourseName)
			continue
		}
		courseID, err := strconv.Atoi(course.CourseID)
		if err != nil {
			scanErrors++
			s.logTask(id, "error", fmt.Sprintf("课程 courseId 解析失败: %s (%v)", course.CourseName, err))
			continue
		}
		userID, err := strconv.Atoi(cache.UserID)
		if err != nil {
			userID = 0
		}
		pointAction, err := xuexitongAggregation.ChapterFetchPointAction(cache, nodes, &chapter, classID, userID, course.Cpi, courseID)
		if err != nil {
			scanErrors++
			s.logTask(id, "error", fmt.Sprintf("无法拉取章节任务点状态: %s (%v)", course.CourseName, err))
			continue
		}

		for index := range pointAction.Knowledge {
			if err := s.waitIfPausedOrStopped(id); err != nil {
				return err
			}
			if index < 0 || index >= len(nodes) {
				continue
			}
			chapterItem := pointAction.Knowledge[index]
			chapterName := xueXiTongChapterLabel(chapterItem)
			if chapterItem.PointTotal > 0 {
				s.logTask(id, "info", fmt.Sprintf("检查章节: %s (任务点 %d/%d)", chapterName, chapterItem.PointFinished, chapterItem.PointTotal))
				if chapterItem.PointFinished >= chapterItem.PointTotal {
					s.logTask(id, "info", "跳过已完成章节: "+chapterName)
					continue
				}
			} else {
				s.logTask(id, "info", "检查章节: "+chapterName)
			}

			_, fetchCards, err := xuexitongAggregation.ChapterFetchCardsAction(cache, &chapter, nodes, index, courseID, classID, course.Cpi)
			if err != nil {
				scanErrors++
				s.logTask(id, "error", fmt.Sprintf("章节卡片拉取失败: %s (%v)", chapterName, err))
				continue
			}
			videoDTOs, workDTOs, documentDTOs, hyperlinkDTOs, liveDTOs, _ := xuexitongApi.ParsePointDto(fetchCards)
			parsed := len(videoDTOs) + len(workDTOs) + len(documentDTOs) + len(hyperlinkDTOs) + len(liveDTOs)
			if parsed == 0 {
				s.logTask(id, "info", "章节未解析到可执行卡片: "+chapterName)
				continue
			}
			discovered += parsed
			s.logTask(id, "info", fmt.Sprintf("解析到卡片: %s，视频/音频 %d，作业 %d，文档 %d，链接 %d，直播 %d", chapterName, len(videoDTOs), len(workDTOs), len(documentDTOs), len(hyperlinkDTOs), len(liveDTOs)))

			for _, videoDTO := range videoDTOs {
				if err := s.waitIfPausedOrStopped(id); err != nil {
					return err
				}
				videoDTO.Logger = log.New(io.Discard, "", 0)
				card, enc, err := xuexitongAggregation.PageMobileChapterCardAction(cache, classID, courseID, videoDTO.KnowledgeID, videoDTO.CardIndex, course.Cpi)
				if err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("视频/音频移动卡片拉取失败: %s (%v)", chapterName, err))
					continue
				}
				if _, err := videoDTO.AttachmentsDetection(card); err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("视频/音频附件解析失败: %s (%v)", chapterName, err))
					continue
				}
				if !videoDTO.IsJob {
					s.logTask(id, "info", fmt.Sprintf("跳过非任务点视频/音频: %s (%d)", chapterName, videoDTO.KnowledgeID))
					continue
				}
				videoDTO.Enc = enc
				s.logTask(id, "info", fmt.Sprintf("开始执行视频/音频任务点: %s (%d)", chapterName, videoDTO.KnowledgeID))
				xuexitongPoint.ExecuteVideoTest(cache, &videoDTO, classID, course.Cpi)
				processed++
				s.logTask(id, "success", fmt.Sprintf("视频/音频任务点完成: %s (%d)", chapterName, videoDTO.KnowledgeID))
				if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
					return err
				}
			}
			for _, documentDTO := range documentDTOs {
				if err := s.waitIfPausedOrStopped(id); err != nil {
					return err
				}
				card, _, err := xuexitongAggregation.PageMobileChapterCardAction(cache, classID, courseID, documentDTO.KnowledgeID, documentDTO.CardIndex, course.Cpi)
				if err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("文档移动卡片拉取失败: %s (%v)", chapterName, err))
					continue
				}
				if _, err := documentDTO.AttachmentsDetection(card); err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("文档附件解析失败: %s (%v)", chapterName, err))
					continue
				}
				if !documentDTO.IsJob {
					s.logTask(id, "info", fmt.Sprintf("跳过非任务点文档: %s (%d)", chapterName, documentDTO.KnowledgeID))
					continue
				}
				if _, err := xuexitongPoint.ExecuteDocument(cache, &documentDTO); err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("文档任务点执行失败: %s (%v)", chapterName, err))
					continue
				}
				processed++
				s.logTask(id, "success", fmt.Sprintf("文档任务点完成: %s (%d)", chapterName, documentDTO.KnowledgeID))
				if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
					return err
				}
			}
			for _, workDTO := range workDTOs {
				if err := s.waitIfPausedOrStopped(id); err != nil {
					return err
				}
				if task.AIURL == "" || task.AIModel == "" {
					s.logTask(id, "info", fmt.Sprintf("检测到作业任务点但未配置 AI，跳过: %s (%d)", chapterName, workDTO.KnowledgeID))
					continue
				}
				mobileCard, _, err := xuexitongAggregation.PageMobileChapterCardAction(cache, classID, courseID, workDTO.KnowledgeID, workDTO.CardIndex, course.Cpi)
				if err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("作业移动卡片拉取失败: %s (%v)", chapterName, err))
					continue
				}
				isJob, err := workDTO.AttachmentsDetection(mobileCard)
				if err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("作业附件解析失败: %s (%v)", chapterName, err))
					continue
				}
				workDTO.IsJob = isJob
				if !workDTO.IsJob {
					s.logTask(id, "info", fmt.Sprintf("跳过非任务点作业: %s (%d)", chapterName, workDTO.KnowledgeID))
					continue
				}
				questionAction, err := xuexitongAggregation.ParseWorkQuestionAction(cache, &workDTO)
				if err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("作业题目解析失败: %s (%v)", chapterName, err))
					continue
				}
				for i := range questionAction.Choice {
					q := &questionAction.Choice[i]
					message := xuexitongAggregation.AIProblemMessage(questionAction.Title, q.Text, xuexitongApi.ExamTurn{XueXChoiceQue: *q})
					q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
				}
				for i := range questionAction.Judge {
					q := &questionAction.Judge[i]
					message := xuexitongAggregation.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{XueXJudgeQue: *q})
					q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
				}
				for i := range questionAction.Fill {
					q := &questionAction.Fill[i]
					message := xuexitongAggregation.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{XueXFillQue: *q})
					q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
				}
				for i := range questionAction.Short {
					q := &questionAction.Short[i]
					message := xuexitongAggregation.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{XueXShortQue: *q})
					q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
				}
				for i := range questionAction.TermExplanation {
					q := &questionAction.TermExplanation[i]
					message := xuexitongAggregation.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{XueXTermExplanationQue: *q})
					q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
				}
				for i := range questionAction.Essay {
					q := &questionAction.Essay[i]
					message := xuexitongAggregation.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{XueXEssayQue: *q})
					q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
				}
				if _, err := xuexitongAggregation.WorkNewSubmitAnswerAction(cache, questionAction, true); err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("作业答案提交失败: %s (%v)", chapterName, err))
					continue
				}
				processed++
				s.logTask(id, "success", fmt.Sprintf("作业任务点完成: %s (%d)", chapterName, workDTO.KnowledgeID))
				if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
					return err
				}
			}
			for _, hyperlinkDTO := range hyperlinkDTOs {
				if err := s.waitIfPausedOrStopped(id); err != nil {
					return err
				}
				if !hyperlinkDTO.IsSet {
					continue
				}
				if _, err := xuexitongPoint.ExecuteHyperlink(cache, &hyperlinkDTO); err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("链接任务点执行失败: %s (%v)", chapterName, err))
					continue
				}
				processed++
				s.logTask(id, "success", fmt.Sprintf("链接任务点完成: %s (%d)", chapterName, hyperlinkDTO.KnowledgeID))
				if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
					return err
				}
			}
			for _, liveDTO := range liveDTOs {
				if err := s.waitIfPausedOrStopped(id); err != nil {
					return err
				}
				if !liveDTO.IsSet {
					continue
				}
				if _, err := xuexitongPoint.ExecuteLive(cache, &liveDTO); err != nil {
					scanErrors++
					s.logTask(id, "error", fmt.Sprintf("直播任务点执行失败: %s (%v)", chapterName, err))
					continue
				}
				processed++
				s.logTask(id, "success", fmt.Sprintf("直播任务点完成: %s (%d)", chapterName, liveDTO.KnowledgeID))
				if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
					return err
				}
			}
		}
	}
	if processed == 0 {
		switch {
		case discovered > 0:
			return fmt.Errorf("学习通解析到 %d 个卡片，但没有成功执行任务点，请查看上方失败或跳过日志", discovered)
		case scanErrors > 0:
			return fmt.Errorf("学习通扫描任务点失败 %d 次，请查看章节/卡片日志", scanErrors)
		case unfinishedHints > 0:
			return fmt.Errorf("课程列表显示仍有未完成任务点，但章节卡片未解析到可执行项，请查看实时日志")
		default:
			s.logTask(id, "info", "没有需要执行的学习通任务点，课程可能已完成、已结束或没有未完成任务")
			return nil
		}
	}
	s.logTask(id, "success", fmt.Sprintf("学习通已执行 %d 个任务点", processed))
	return nil
}

func (s *Server) runXueXiTongTaskLegacy(id string, task Task) error {
	if err := s.waitIfPausedOrStopped(id); err != nil {
		return err
	}
	cache := &xuexitongApi.XueXiTUserCache{Name: task.Account, Password: task.Password}
	s.logTask(id, "info", "学习通登录中")
	if err := xuexitongAggregation.XueXiTLoginAction(cache); err != nil {
		return fmt.Errorf("学习通登录失败: %w", err)
	}
	s.logTask(id, "success", "学习通登录成功")
	courses, err := xuexitongAggregation.XueXiTPullCourseAction(cache)
	if err != nil {
		return fmt.Errorf("学习通拉取课程失败: %w", err)
	}
	courses = filterXueXiTongCourses(courses, task.CourseIDs)
	if len(courses) == 0 {
		return fmt.Errorf("学习通未找到匹配课程")
	}

	processed := 0
	s.logTask(id, "info", fmt.Sprintf("准备执行 %d 门课程", len(courses)))
	for _, course := range courses {
		if err := s.waitIfPausedOrStopped(id); err != nil {
			return err
		}
		if course.State == 1 {
			s.logTask(id, "info", "跳过已结束课程: "+course.CourseName)
			continue
		}
		if !course.IsStart {
			s.logTask(id, "info", "跳过未开课课程: "+course.CourseName)
			continue
		}
		s.logTask(id, "info", fmt.Sprintf("处理课程: %s (列表进度 %d/%d，仅供参考)", course.CourseName, course.JobFinishCount, course.JobCount))
		classID, err := strconv.Atoi(course.Key)
		if err != nil {
			continue
		}
		chapter, ok, err := xuexitongAggregation.PullCourseChapterAction(cache, course.Cpi, classID)
		if err != nil || !ok {
			s.logTask(id, "info", "跳过无法拉取章节的课程: "+course.CourseName)
			continue
		}
		var nodes []int
		for _, item := range chapter.Knowledge {
			nodes = append(nodes, item.ID)
		}
		if len(nodes) == 0 {
			s.logTask(id, "info", "跳过无章节任务点的课程: "+course.CourseName)
			continue
		}
		courseID, err := strconv.Atoi(course.CourseID)
		if err != nil {
			continue
		}
		userID, err := strconv.Atoi(cache.UserID)
		if err != nil {
			userID = 0
		}
		pointAction, err := xuexitongAggregation.ChapterFetchPointAction(cache, nodes, &chapter, classID, userID, course.Cpi, courseID)
		if err != nil {
			s.logTask(id, "info", "跳过无法拉取任务点状态的课程: "+course.CourseName)
			continue
		}
		for index := range pointAction.Knowledge {
			if err := s.waitIfPausedOrStopped(id); err != nil {
				return err
			}
			if index < 0 || index >= len(nodes) {
				continue
			}
			_, fetchCards, err := xuexitongAggregation.ChapterFetchCardsAction(cache, &chapter, nodes, index, courseID, classID, course.Cpi)
			if err != nil {
				continue
			}
			videoDTOs, workDTOs, documentDTOs, hyperlinkDTOs, liveDTOs, _ := xuexitongApi.ParsePointDto(fetchCards)
			if len(videoDTOs) == 0 && len(workDTOs) == 0 && len(documentDTOs) == 0 && len(hyperlinkDTOs) == 0 && len(liveDTOs) == 0 {
				continue
			}
			for _, videoDTO := range videoDTOs {
				if err := s.waitIfPausedOrStopped(id); err != nil {
					return err
				}
				if !videoDTO.IsJob {
					continue
				}
				card, enc, err := xuexitongAggregation.PageMobileChapterCardAction(cache, classID, courseID, videoDTO.KnowledgeID, videoDTO.CardIndex, course.Cpi)
				if err != nil {
					continue
				}
				videoDTO.AttachmentsDetection(card)
				videoDTO.Enc = enc
				xuexitongPoint.ExecuteVideoTest(cache, &videoDTO, classID, course.Cpi)
				processed++
				s.logTask(id, "info", fmt.Sprintf("视频任务点完成: %d", videoDTO.KnowledgeID))
				if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
					return err
				}
			}
			for _, documentDTO := range documentDTOs {
				if err := s.waitIfPausedOrStopped(id); err != nil {
					return err
				}
				if !documentDTO.IsJob {
					continue
				}
				card, _, err := xuexitongAggregation.PageMobileChapterCardAction(cache, classID, courseID, documentDTO.KnowledgeID, documentDTO.CardIndex, course.Cpi)
				if err != nil {
					continue
				}
				documentDTO.AttachmentsDetection(card)
				if _, err := xuexitongPoint.ExecuteDocument(cache, &documentDTO); err != nil {
					continue
				}
				processed++
				s.logTask(id, "info", fmt.Sprintf("文档任务点完成: %d", documentDTO.KnowledgeID))
				if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
					return err
				}
			}
			if task.AIURL != "" && task.AIModel != "" {
				for _, workDTO := range workDTOs {
					if err := s.waitIfPausedOrStopped(id); err != nil {
						return err
					}
					if !workDTO.IsJob {
						continue
					}
					mobileCard, _, err := xuexitongAggregation.PageMobileChapterCardAction(cache, classID, courseID, workDTO.KnowledgeID, workDTO.CardIndex, course.Cpi)
					if err != nil {
						continue
					}
					workDTO.AttachmentsDetection(mobileCard)
					questionAction, err := xuexitongAggregation.ParseWorkQuestionAction(cache, &workDTO)
					if err != nil {
						continue
					}
					for i := range questionAction.Choice {
						q := &questionAction.Choice[i]
						message := xuexitongAggregation.AIProblemMessage(questionAction.Title, q.Text, xuexitongApi.ExamTurn{XueXChoiceQue: *q})
						q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
					}
					for i := range questionAction.Judge {
						q := &questionAction.Judge[i]
						message := xuexitongAggregation.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{XueXJudgeQue: *q})
						q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
					}
					for i := range questionAction.Fill {
						q := &questionAction.Fill[i]
						message := xuexitongAggregation.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{XueXFillQue: *q})
						q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
					}
					for i := range questionAction.Short {
						q := &questionAction.Short[i]
						message := xuexitongAggregation.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{XueXShortQue: *q})
						q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
					}
					for i := range questionAction.TermExplanation {
						q := &questionAction.TermExplanation[i]
						message := xuexitongAggregation.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{XueXTermExplanationQue: *q})
						q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
					}
					for i := range questionAction.Essay {
						q := &questionAction.Essay[i]
						message := xuexitongAggregation.AIProblemMessage(q.Type.String(), q.Text, xuexitongApi.ExamTurn{XueXEssayQue: *q})
						q.AnswerAIGet(cache.UserID, task.AIURL, task.AIModel, ctype.AiType(task.AIType), message, task.AIKey)
					}
					_, _ = xuexitongAggregation.WorkNewSubmitAnswerAction(cache, questionAction, true)
					processed++
					s.logTask(id, "info", fmt.Sprintf("作业任务点完成: %d", workDTO.KnowledgeID))
					if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
						return err
					}
				}
			}
			for _, hyperlinkDTO := range hyperlinkDTOs {
				if err := s.waitIfPausedOrStopped(id); err != nil {
					return err
				}
				if !hyperlinkDTO.IsSet {
					continue
				}
				if _, err := xuexitongPoint.ExecuteHyperlink(cache, &hyperlinkDTO); err != nil {
					continue
				}
				processed++
				s.logTask(id, "info", fmt.Sprintf("链接任务点完成: %d", hyperlinkDTO.KnowledgeID))
				if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
					return err
				}
			}
			for _, liveDTO := range liveDTOs {
				if err := s.waitIfPausedOrStopped(id); err != nil {
					return err
				}
				if !liveDTO.IsSet {
					continue
				}
				if _, err := xuexitongPoint.ExecuteLive(cache, &liveDTO); err != nil {
					continue
				}
				processed++
				s.logTask(id, "info", fmt.Sprintf("直播任务点完成: %d", liveDTO.KnowledgeID))
				if err := s.interruptibleSleep(id, 2*time.Second); err != nil {
					return err
				}
			}
		}
	}
	if processed == 0 {
		s.logTask(id, "info", "没有需要执行的学习通任务点，可能课程已完成、已结束或没有未完成任务")
		return nil
	}
	return nil
}

func filterXueXiTongCourses(courses []xuexitongAggregation.XueXiTCourse, ids []string) []xuexitongAggregation.XueXiTCourse {
	if len(ids) == 0 {
		return courses
	}
	selected := stringSet(ids)
	out := make([]xuexitongAggregation.XueXiTCourse, 0, len(courses))
	for _, course := range courses {
		if _, ok := selected[course.CourseID]; ok {
			out = append(out, course)
		}
	}
	return out
}

func xueXiTongChapterLabel(item xuexitongAggregation.KnowledgeItem) string {
	parts := nonEmptyStrings(item.Label, item.Name)
	if len(parts) == 0 {
		return fmt.Sprintf("章节 %d", item.ID)
	}
	return strings.Join(parts, " ")
}

func stringSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	return set
}

func countXueXiTongTaskPoints(videoDTOs []xuexitongApi.PointVideoDto, documentDTOs []xuexitongApi.PointDocumentDto, hyperlinkDTOs []xuexitongApi.PointHyperlinkDto, liveDTOs []xuexitongApi.PointLiveDto) int {
	count := 0
	count += len(videoDTOs)
	count += len(documentDTOs)
	count += len(hyperlinkDTOs)
	count += len(liveDTOs)
	return count
}
