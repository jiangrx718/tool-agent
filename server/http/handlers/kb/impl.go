package kb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"

	"tool-agent/internal/kbagent"
	kbService "tool-agent/internal/service/kb"
	"tool-agent/server/http/response"
	"tool-agent/utils"
)

// NewKBHandler 创建知识库处理器
func NewKBHandler(engine *gin.Engine) *KBHandler {
	return &KBHandler{
		engine:  engine,
		service: kbService.NewKBService(),
	}
}

// KBHandler 知识库处理器
type KBHandler struct {
	engine  *gin.Engine
	service kbService.ServiceIFace
}

// RegisterRoutes 注册知识库路由
func (h *KBHandler) RegisterRoutes(routerGroup *gin.RouterGroup) {
	g := routerGroup.Group("/kb")

	// 文档管理
	g.POST("/documents", h.CreateDocument)
	g.GET("/documents", h.ListDocuments)
	g.POST("/documents/delete", h.DeleteDocument)

	// ReAct 智能体问答
	g.POST("/agent/chat", h.AgentChat)
	g.GET("/agent/stream", h.AgentStream)

	// RAG 检索增强问答
	g.POST("/rag/ask", h.RagAsk)
	g.GET("/rag/stream", h.RagStream)
}

// ---- 请求/响应类型 ----

type createDocReq struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type deleteDocReq struct {
	ID uint64 `json:"id" binding:"required"`
}

type chatReq struct {
	Question string `json:"question" binding:"required"`
}

type chatResp struct {
	Answer  string        `json:"answer"`
	Sources []*sourceInfo `json:"sources,omitempty"`
}

type sourceInfo struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Score float64 `json:"score"`
}

// ---- 文档管理接口 ----

func (h *KBHandler) CreateDocument(c *gin.Context) {
	var req createDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParameterError(c, err)
		return
	}

	id, err := h.service.CreateDocument(c.Request.Context(), req.Title, req.Content)
	if err != nil {
		utils.Sugar().Errorf("[kb] create document: %v", err)
		response.Failed(c, response.CodeInternalErr, "文档入库失败: "+err.Error(), nil)
		return
	}
	response.Successful(c, gin.H{"id": id})
}

func (h *KBHandler) ListDocuments(c *gin.Context) {
	docs, err := h.service.ListDocuments(c.Request.Context())
	if err != nil {
		utils.Sugar().Errorf("[kb] list documents: %v", err)
		response.InternalError(c)
		return
	}
	response.Successful(c, docs)
}

func (h *KBHandler) DeleteDocument(c *gin.Context) {
	var req deleteDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParameterError(c, err)
		return
	}

	if err := h.service.DeleteDocument(c.Request.Context(), req.ID); err != nil {
		utils.Sugar().Errorf("[kb] delete document: %v", err)
		response.InternalError(c)
		return
	}
	response.Successful(c, nil)
}

// ---- ReAct 智能体接口 ----

func (h *KBHandler) AgentChat(c *gin.Context) {
	var req chatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParameterError(c, err)
		return
	}

	answer, sources, err := h.service.AgentChat(c.Request.Context(), req.Question)
	if err != nil {
		utils.Sugar().Errorf("[kb] agent chat: %v", err)
		response.Failed(c, response.CodeInternalErr, "智能体调用失败: "+err.Error(), nil)
		return
	}
	response.Successful(c, chatResp{Answer: answer, Sources: toSourceInfos(sources)})
}

func (h *KBHandler) AgentStream(c *gin.Context) {
	question := c.Query("question")
	if question == "" {
		response.ParameterError(c, errors.New("question is required"))
		return
	}

	reader, err := h.service.AgentStream(c.Request.Context(), question)
	if err != nil {
		utils.Sugar().Errorf("[kb] agent stream: %v", err)
		response.Failed(c, response.CodeInternalErr, "智能体流式调用失败: "+err.Error(), nil)
		return
	}
	defer reader.Close()
	streamSSE(c, reader)
}

// ---- RAG 接口 ----

func (h *KBHandler) RagAsk(c *gin.Context) {
	var req chatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParameterError(c, err)
		return
	}

	answer, sources, err := h.service.RagAsk(c.Request.Context(), req.Question)
	if err != nil {
		utils.Sugar().Errorf("[kb] rag ask: %v", err)
		response.Failed(c, response.CodeInternalErr, "RAG 调用失败: "+err.Error(), nil)
		return
	}
	response.Successful(c, chatResp{Answer: answer, Sources: toSourceInfos(sources)})
}

func (h *KBHandler) RagStream(c *gin.Context) {
	question := c.Query("question")
	if question == "" {
		response.ParameterError(c, errors.New("question is required"))
		return
	}

	reader, err := h.service.RagStream(c.Request.Context(), question)
	if err != nil {
		utils.Sugar().Errorf("[kb] rag stream: %v", err)
		response.Failed(c, response.CodeInternalErr, "RAG 流式调用失败: "+err.Error(), nil)
		return
	}
	defer reader.Close()
	streamSSE(c, reader)
}

// ---- 辅助 ----

func toSourceInfos(sources []*kbagent.Source) []*sourceInfo {
	if sources == nil {
		return nil
	}
	out := make([]*sourceInfo, len(sources))
	for i, s := range sources {
		out[i] = &sourceInfo{ID: s.ID, Title: s.Title, Score: s.Score}
	}
	return out
}

// streamSSE 将 Eino StreamReader 以 SSE 格式推送给客户端
func streamSSE(c *gin.Context, reader *schema.StreamReader[*schema.Message]) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	for {
		msg, err := reader.Recv()
		if errors.Is(err, io.EOF) {
			fmt.Fprintf(c.Writer, "data: [DONE]\n\n")
			flusher.Flush()
			return
		}
		if err != nil {
			errData, _ := json.Marshal(map[string]string{"error": err.Error()})
			fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", errData)
			flusher.Flush()
			return
		}
		if msg != nil && msg.Content != "" {
			data, _ := json.Marshal(map[string]string{"content": msg.Content})
			fmt.Fprintf(c.Writer, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
