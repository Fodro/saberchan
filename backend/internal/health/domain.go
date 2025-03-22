package health

type Service interface {
	Readiness() error
}