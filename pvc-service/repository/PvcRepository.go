package repository

type PvcRepository interface {
	CachePvcList(pvcList []string) error
	CheckPvcExistsInCache(pvcName string) (bool, error)
	CreatePvc(pvcName string) error
	DeletePvc(pvcName string) error
}
