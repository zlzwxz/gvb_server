package routers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"gvb-server/config"
	"gvb-server/global"

	"github.com/gin-gonic/gin"
)

func TestRegisterUploadStaticRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	oldConfig := global.Config
	global.Config = &config.Config{
		Upload: config.Upload{
			Path: filepath.Join("uploads", "file"),
		},
	}
	defer func() {
		global.Config = oldConfig
	}()

	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	tempDir := t.TempDir()
	if err = os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(oldWD)
	}()

	fixtures := map[string]string{
		filepath.Join("uploads", "avatar", "default.jpeg"):              "avatar",
		filepath.Join("uploads", "file", "u_1", "demo.png"):             "image",
		filepath.Join("uploads", "file", "attachments", "u_1", "a.txt"): "attachment",
	}
	for path, content := range fixtures {
		if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err = os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write fixture %s: %v", path, err)
		}
	}

	router := gin.New()
	registerUploadStaticRoute(router)

	testCases := []struct {
		name string
		path string
		want int
	}{
		{name: "serve avatar", path: "/uploads/avatar/default.jpeg", want: http.StatusOK},
		{name: "serve image under upload path", path: "/uploads/file/u_1/demo.png", want: http.StatusOK},
		{name: "block attachment direct access", path: "/uploads/file/attachments/u_1/a.txt", want: http.StatusForbidden},
		{name: "block path traversal", path: "/uploads/../settings.yaml", want: http.StatusForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, tc.path, nil)
			router.ServeHTTP(recorder, request)
			if recorder.Code != tc.want {
				t.Fatalf("path %s returned %d, want %d", tc.path, recorder.Code, tc.want)
			}
		})
	}
}
