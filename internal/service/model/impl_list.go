package model

import (
	"context"

	"tool-agent/internal/common"
	"tool-agent/internal/dao"
	"tool-agent/utils"
)

// ListResponseData 模型列表响应数据
type ListResponseData struct {
	List   []ModelItem `json:"list"`
	Count  int64       `json:"count"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
}

// ModelItem 模型列表项
type ModelItem struct {
	ID            uint   `json:"id"`
	DataType      int8   `json:"data_type"`
	ModelID       string `json:"model_id"`
	ModelName     string `json:"model_name"`
	ModelPath     string `json:"model_path"`
	RunPath       string `json:"run_path"`
	ParamsPath    string `json:"params_path"`
	TrainCallback string `json:"train_callback"`
	InferCallback string `json:"infer_callback"`
	InferPath     string `json:"infer_path"`
	Status        int8   `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// List 查询模型列表
func (s *Service) List(ctx context.Context, keyword string, offset, limit int) (common.ServiceResult, error) {
	logger := utils.SugarContext(ctx)

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	list, count, err := s.modelDao.List(ctx, dao.ListQuery{
		Keyword: keyword,
		Offset:  offset,
		Limit:   limit,
	})
	if err != nil {
		logger.Errorw("ModelService List modelDao.List error", "error", err)
		return common.ServiceResult{}, err
	}

	items := make([]ModelItem, 0, len(list))
	for _, m := range list {
		items = append(items, ModelItem{
			ModelName:     m.ModelName,
			ModelPath:     m.ModelPath,
			RunPath:       m.RunPath,
			ParamsPath:    m.ParamsPath,
			TrainCallback: m.TrainCallback,
			InferCallback: m.InferCallback,
			InferPath:     m.InferPath,
			CreatedAt:     m.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:     m.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return common.NewServiceResult(ListResponseData{
		List:   items,
		Count:  count,
		Offset: offset,
		Limit:  limit,
	}), nil
}
