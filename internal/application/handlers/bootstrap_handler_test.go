package handlers

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/bnema/archup/internal/domain/ports/mocks"
	"go.uber.org/mock/gomock"
)

func TestBootstrapHandler_DownloadFiles_RejectsHTTPErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockFS := mocks.NewMockFileSystem(ctrl)
	mockHTTP := mocks.NewMockHTTPClient(ctrl)
	mockLogger := mocks.NewMockLogger(ctrl)

	mockLogger.EXPECT().Info(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any(), gomock.Any()).AnyTimes()
	mockFS.EXPECT().MkdirAll(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	mockHTTP.EXPECT().
		Get("https://raw.githubusercontent.com/bnema/archup/v1.2.3/install/base.packages").
		Return(newMockResponse(ctrl, http.StatusNotFound, []byte("not found")), nil)

	handler := NewBootstrapHandler(
		mockFS,
		mockHTTP,
		mockLogger,
		"https://github.com/bnema/archup",
		"https://raw.githubusercontent.com/bnema/archup/v1.2.3",
		"v1.2.3",
	)

	err := handler.downloadFiles(context.Background())
	if err == nil {
		t.Fatal("expected downloadFiles to fail on non-2xx response")
	}

	if !strings.Contains(err.Error(), "unexpected status 404") {
		t.Fatalf("expected status error, got %v", err)
	}
}
