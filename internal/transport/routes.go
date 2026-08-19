package transport

import "net/http"

type Route struct {
	Method, Pattern, Summary string
	Auth                     bool
	Handler                  http.HandlerFunc
}
type Registry struct{ routes []Route }

func (r *Registry) Add(route Route) { r.routes = append(r.routes, route) }
func (r Registry) Routes() []Route  { return append([]Route(nil), r.routes...) }
func (r Registry) Mount(mux *http.ServeMux) {
	for _, route := range r.routes {
		method := route.Method
		handler := route.Handler
		mux.HandleFunc(route.Pattern, func(w http.ResponseWriter, req *http.Request) {
			if req.Method != method {
				w.Header().Set("Allow", method)
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handler(w, req)
		})
	}
}
func Health(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/health", Summary: "Liveness probe", Handler: handler}
}
func Ready(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/ready", Summary: "Readiness probe", Handler: handler}
}
func Metrics(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/metrics", Summary: "Prometheus metrics", Handler: handler}
}
func CreateWebhook(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodPost, Pattern: "/api/v1/webhooks", Summary: "Receive a webhook event", Handler: handler}
}
func ListEvents(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/api/v1/events", Summary: "List and filter events", Handler: handler}
}
func GetEvent(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/api/v1/events/{id}", Summary: "Get event details", Handler: handler}
}
func ReplayEvent(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodPost, Pattern: "/api/v1/events/{id}/replay", Summary: "Replay one event", Handler: handler}
}
func ReplayEvents(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodPost, Pattern: "/api/v1/events/replay", Summary: "Replay matching events", Handler: handler}
}
func ScheduleEvent(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodPost, Pattern: "/api/v1/events/{id}/schedule", Summary: "Schedule delivery", Handler: handler}
}
func CancelSchedule(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodDelete, Pattern: "/api/v1/events/{id}/schedule", Summary: "Cancel scheduled delivery", Handler: handler}
}
func ListTasks(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/api/v1/delivery-tasks", Summary: "List delivery tasks", Handler: handler}
}
func GetTask(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/api/v1/delivery-tasks/{id}", Summary: "Get delivery task", Handler: handler}
}
func RetryTask(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodPost, Pattern: "/api/v1/delivery-tasks/{id}/retry", Summary: "Retry a task", Handler: handler}
}
func CancelTask(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodPost, Pattern: "/api/v1/delivery-tasks/{id}/cancel", Summary: "Cancel a task", Handler: handler}
}
func ListDeadLetters(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/api/v1/dead-letters", Summary: "List dead letters", Handler: handler}
}
func GetDeadLetter(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/api/v1/dead-letters/{id}", Summary: "Get dead letter", Handler: handler}
}
func RetryDeadLetter(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodPost, Pattern: "/api/v1/dead-letters/{id}/retry", Summary: "Retry a dead letter", Handler: handler}
}
func DiscardDeadLetter(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodPost, Pattern: "/api/v1/dead-letters/{id}/discard", Summary: "Discard a dead letter", Handler: handler}
}
func Statistics(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/api/v1/statistics", Summary: "Get service statistics", Handler: handler}
}
func AuditLogs(handler http.HandlerFunc) Route {
	return Route{Method: http.MethodGet, Pattern: "/api/v1/audit-logs", Summary: "List audit records", Handler: handler}
}
func DefaultRouteNames() map[string]string {
	return map[string]string{"health": "health", "ready": "ready", "metrics": "metrics", "webhooks": "webhooks", "events": "events", "event": "event", "event-replay": "event-replay", "batch-replay": "batch-replay", "schedule": "schedule", "task-list": "task-list", "task-retry": "task-retry", "dead-letter-list": "dead-letter-list", "dead-letter-retry": "dead-letter-retry", "statistics": "statistics", "audit-logs": "audit-logs"}
}
func IsWriteMethod(method string) bool {
	return method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch || method == http.MethodDelete
}
func IsSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
func AllowedMethods() []string {
	return []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodOptions}
}
