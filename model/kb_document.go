package model

// KbDocument 知识库文档模型，存储文档内容及其向量表示
type KbDocument struct {
	BaseModelFieldId
	Title     string `gorm:"column:title;type:varchar(255);not null;comment:文档标题" json:"title"`
	Content   string `gorm:"column:content;type:longtext;not null;comment:文档内容" json:"content"`
	Embedding string `gorm:"column:embedding;type:longtext;comment:向量JSON" json:"-"`
	BaseModelFieldTime
}

func (m *KbDocument) TableName() string {
	return "kb_document"
}
