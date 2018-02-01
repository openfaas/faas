package handlers

// ServiceQuery provides interface for replica querying/setting
type ServiceQuery interface {
	GetReplicas(service string) (currentReplicas uint64, maxReplicas uint64, minReplicas uint64, err error)
	SetReplicas(service string, count uint64) error
}
