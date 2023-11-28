package my_cache

import (
	"log"
	"net/http"
)

const defaultBasePath = "/_geecache/"

type httpApi struct {
	basePath string
	self     string
}

func NewHttpApi(self string) *httpApi {
	return &httpApi{basePath: defaultBasePath, self: self}

}
func (h *httpApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// /<basepath> ?<groupname>&&<key> required
	log.Print("url:", r.URL.Path)

	groupName := r.URL.Query().Get("groupName")
	key := r.URL.Query().Get("key")

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteView())
}
