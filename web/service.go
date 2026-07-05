package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	haiqikejiAggregation "github.com/yatori-dev/yatori-go-core/aggregation/haiqikeji"
	haiqikejiApi "github.com/yatori-dev/yatori-go-core/api/haiqikeji"
	xuexitongAggregation "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	xuexitongPoint "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong/point"
	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	yinghuaAggregation "github.com/yatori-dev/yatori-go-core/aggregation/yinghua"
	yinghuaApi "github.com/yatori-dev/yatori-go-core/api/yinghua"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
)

type Server struct {
	store *taskStore
}

func NewServer() *Server {
	return &Server{store: newTaskStore()}
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.indexHandler)
	mux.HandleFunc("/submit", s.submitHandler)
	mux.HandleFunc("/tasks", s.tasksHandler)
	mux.HandleFunc("/tasks/", s.taskHandler)
	log.Printf("web server listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl := template.Must(template.New("index").Parse(`<!DOCTYPE html>
<html>
<head><meta charset="utf-8"><title>Yatori Web</title><style>body{font-family:Arial,sans-serif;max-width:760px;margin:32px auto;padding:16px}label{display:block;margin-top:12px}input,select,textarea{width:100%;padding:8px;margin-top:6px}button{margin-top:16px;padding:8px 12px}</style></head>
<body>
<h2>Yatori Web 任务提交</h2>
<form action="/submit" method="post">
<label>平台<select name="platform"><option value="haiqikeji">海旗</option><option value="yinghua">英华</option><option value="xuexitong">学习通</option></select></label>
<label>账号<input name="account" required></label>
<label>密码<input name="password" type="password" required></label>
<label>平台地址<input name="preUrl" placeholder="海旗/英华填写站点地址；学习通可留空或填写登录后地址" ></label>
<label>AI地址<input name="aiUrl" placeholder="可选：例如 https://api.example.com/v1"></label>
<label>AI模型<input name="aiModel" placeholder="可选：例如 gpt-4o"></label>
<label>AI密钥<input name="aiKey" placeholder="可选：填写你的 AI 调用密钥"></label>
<label>AI类型<select name="aiType"><option value="">默认</option><option value="OPENAI">OpenAI</option><option value="DEEPSEEK">DeepSeek</option><option value="CHATGLM">ChatGLM</option><option value="TONGYI">通义</option><option value="DOUBAO">豆包</option><option value="OTHER">Other</option></select></label>
<label>说明<textarea name="message" placeholder="可选：自定义说明"></textarea></label>
<button type="submit">提交一键刷课任务</button>
</form>
<hr>
<h3>最近任务</h3>
<div id="tasks">加载中...</div>
<script>
fetch('/tasks').then(r=>r.json()).then(data=>{const el=document.getElementById('tasks'); if(!data.length){el.innerHTML='暂无任务';return;} el.innerHTML=data.map(t=>'<div style="border:1px solid #ddd;padding:10px;margin-top:8px"><div><b>'+t.platform+'</b> - '+t.account+'</div><div>状态:'+t.status+' | 任务ID:'+t.id+'</div><div>'+ (t.message || '') +'</div></div>').join('');}).catch(()=>{document.getElementById('tasks').innerHTML='加载失败';});
</script>
</body></html>`))
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	platform := strings.TrimSpace(r.Form.Get("platform"))
	account := strings.TrimSpace(r.Form.Get("account"))
	password := strings.TrimSpace(r.Form.Get("password"))
	preURL := strings.TrimSpace(r.Form.Get("preUrl"))
	aiURL := strings.TrimSpace(r.Form.Get("aiUrl"))
	aiModel := strings.TrimSpace(r.Form.Get("aiModel"))
	aiKey := strings.TrimSpace(r.Form.Get("aiKey"))
	aiType := strings.TrimSpace(r.Form.Get("aiType"))
	message := strings.TrimSpace(r.Form.Get("message"))
	if platform == "" || account == "" || password == "" {
		http.Error(w, "平台、账号和密码不能为空", http.StatusBadRequest)
		return
	}
	if platform != "xuexitong" && preURL == "" {
		http.Error(w, "海旗和英华需要填写平台地址", http.StatusBadRequest)
		return
	}
	task := s.store.enqueue(Task{Platform: platform, Account: account, Password: password, PreURL: preURL, AIURL: aiURL, AIModel: aiModel, AIKey: aiKey, AIType: aiType, Message: message})
	go s.runTask(task.ID)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "task": task})
}

func (s *Server) tasksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(s.store.list())
}

func (s *Server) taskHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/tasks/")
	if id == "" {
		http.NotFound(w, r)
		return
	}
	task, ok := s.store.get(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (s *Server) runTask(id string) {
	if !s.store.update(id, func(task *Task) {
		task.Status = TaskRunning
		task.Message = "开始执行"
	}) {
		return
	}
	defer func() {
		s.store.update(id, func(task *Task) {
			task.Message = "执行结束"
		})
	}()

	task, ok := s.store.get(id)
	if !ok {
		return
	}
	var err error
	switch task.Platform {
	case "haiqikeji":
		err = s.runHaiqiTask(task)
	case "yinghua":
		err = s.runYinghuaTask(task)
	case "xuexitong":
		err = s.runXueXiTongTask(task)
	default:
		err = fmt.Errorf("unsupported platform %s", task.Platform)
	}
	if err != nil {
		s.store.update(id, func(task *Task) {
			task.Status = TaskFailed
			task.Message = err.Error()
		})
		return
	}
	s.store.update(id, func(task *Task) {
		task.Status = TaskSucceeded
		task.Message = "任务已完成"
	})
}

func (s *Server) runHaiqiTask(task Task) error {
	cache := &haiqikejiApi.HqkjUserCache{PreUrl: task.PreURL, Account: task.Account, Password: task.Password}
	if err := haiqikejiAggregation.HqkjLoginAction(cache); err != nil {
		return fmt.Errorf("海旗登录失败: %w", err)
	}
	courses, err := haiqikejiAggregation.HqkjCourseListAction(cache)
	if err != nil {
		return fmt.Errorf("海旗拉取课程失败: %w", err)
	}
	if len(courses) == 0 {
		return fmt.Errorf("海旗未找到课程")
	}
	for _, course := range courses {
		nodes, err := haiqikejiAggregation.HqkjNodeListAction(cache, course)
		if err != nil {
			return fmt.Errorf("海旗拉取节点失败: %w", err)
		}
		for _, node := range nodes {
			if node.TabVideo != 1 && node.TabVideo != 0 {
				continue
			}
			_, err := haiqikejiAggregation.HqkjSubmitFastStudyTimeAction(cache, node)
			if err != nil {
				return fmt.Errorf("海旗刷课失败: %w", err)
			}
			time.Sleep(2 * time.Second)
		}
	}
	return nil
}

func (s *Server) runYinghuaTask(task Task) error {
	cache := &yinghuaApi.YingHuaUserCache{PreUrl: task.PreURL, Account: task.Account, Password: task.Password}
	if err := yinghuaAggregation.YingHuaLoginAction(cache); err != nil {
		return fmt.Errorf("英华登录失败: %w", err)
	}
	courses, err := yinghuaAggregation.CourseListAction(cache)
	if err != nil {
		return fmt.Errorf("英华拉取课程失败: %w", err)
	}
	if len(courses) == 0 {
		return fmt.Errorf("英华未找到课程")
	}
	for _, course := range courses {
		nodes, err := yinghuaAggregation.VideosListAction(cache, course)
		if err != nil {
			return fmt.Errorf("英华拉取节点失败: %w", err)
		}
		for _, node := range nodes {
			if !node.TabVideo {
				continue
			}
			_, err := yinghuaAggregation.SubmitStudyTimeAction(cache, node.Id, node.Id, 60)
			if err != nil {
				return fmt.Errorf("英华刷课失败: %w", err)
			}
			time.Sleep(2 * time.Second)
		}
	}
	return nil
}

func (s *Server) runXueXiTongTask(task Task) error {
	cache := &xuexitongApi.XueXiTUserCache{Name: task.Account, Password: task.Password}
	if err := xuexitongAggregation.XueXiTLoginAction(cache); err != nil {
		return fmt.Errorf("学习通登录失败: %w", err)
	}
	courses, err := xuexitongAggregation.XueXiTPullCourseAction(cache)
	if err != nil {
		return fmt.Errorf("学习通拉取课程失败: %w", err)
	}
	if len(courses) == 0 {
		return fmt.Errorf("学习通未找到课程")
	}

	processed := 0
	for _, course := range courses {
		if !course.IsStart {
			continue
		}
		classID, err := strconv.Atoi(course.Key)
		if err != nil {
			continue
		}
		chapter, ok, err := xuexitongAggregation.PullCourseChapterAction(cache, course.Cpi, classID)
		if err != nil || !ok {
			continue
		}
		var nodes []int
		for _, item := range chapter.Knowledge {
			nodes = append(nodes, item.ID)
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
			continue
		}
		for index := range pointAction.Knowledge {
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
				time.Sleep(2 * time.Second)
			}
			for _, documentDTO := range documentDTOs {
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
				time.Sleep(2 * time.Second)
			}
			if task.AIURL != "" && task.AIModel != "" {
				for _, workDTO := range workDTOs {
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
					time.Sleep(2 * time.Second)
				}
			}
			for _, hyperlinkDTO := range hyperlinkDTOs {
				if !hyperlinkDTO.IsSet {
					continue
				}
				if _, err := xuexitongPoint.ExecuteHyperlink(cache, &hyperlinkDTO); err != nil {
					continue
				}
				processed++
				time.Sleep(2 * time.Second)
			}
			for _, liveDTO := range liveDTOs {
				if !liveDTO.IsSet {
					continue
				}
				if _, err := xuexitongPoint.ExecuteLive(cache, &liveDTO); err != nil {
					continue
				}
				processed++
				time.Sleep(2 * time.Second)
			}
		}
	}
	if processed == 0 {
		return fmt.Errorf("学习通未找到可执行任务点")
	}
	return nil
}

func countXueXiTongTaskPoints(videoDTOs []xuexitongApi.PointVideoDto, documentDTOs []xuexitongApi.PointDocumentDto, hyperlinkDTOs []xuexitongApi.PointHyperlinkDto, liveDTOs []xuexitongApi.PointLiveDto) int {
	count := 0
	count += len(videoDTOs)
	count += len(documentDTOs)
	count += len(hyperlinkDTOs)
	count += len(liveDTOs)
	return count
}
