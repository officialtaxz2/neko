package legacy

import "net/http"

const webCodecsMediaBackend = "webcodecs-ws"

func webCodecsMediaSelected(request *http.Request) bool {
	values, exists := request.URL.Query()["media"]
	return exists && len(values) == 1 && values[0] == webCodecsMediaBackend
}
