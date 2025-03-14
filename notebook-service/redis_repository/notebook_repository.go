package redis_repository

type NotebookRepository interface {
	CheckCacheExists(string) (bool, error)
	StoreNotebooks(string, []string) error
	AddNotebook(string, string) error
	GetNotebooks(string) ([]string, error)
	DeleteNotebook(string, string) error
}
